package speedtest

import (
	"context"
	"io"
	"math"
	"net/http"
	"sync"
	"time"
)

var defaultURLs = []string{
	"https://proof.ovh.net/files/100Mb.dat",
	"https://speedtest.tele2.net/100MB.zip",
	"https://speed.hetzner.de/100MB.bin",
}

const (
	workers        = 4
	chunkSize      = 64 * 1024
	reportInterval = 200 * time.Millisecond
	minDuration    = 5 * time.Second
	maxDuration    = 15 * time.Second
)

func Run(ctx context.Context, updates chan<- SpeedSample) {
	defer close(updates)

	url := pickURL(ctx)
	if url == "" {
		updates <- SpeedSample{Done: true, Error: errNoReachableURL()}
		return
	}

	reportCh := make(chan int64, 2000)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			buf := make([]byte, chunkSize)
			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					select {
					case reportCh <- int64(n):
					case <-ctx.Done():
						return
					}
				}
				if err == io.EOF {
					return
				}
				if err != nil {
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(reportCh)
	}()

	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()

	var totalBytes int64
	start := time.Now()
	var instSpeeds []float64
	chOpen := true

	for chOpen {
		select {
		case <-ctx.Done():
			return
		case n, ok := <-reportCh:
			if !ok {
				chOpen = false
				continue
			}
			totalBytes += n
		case <-ticker.C:
			elapsed := time.Since(start)

			avgMbps := float64(totalBytes) * 8 / elapsed.Seconds() / 1_000_000
			if avgMbps > 0 {
				instSpeeds = append(instSpeeds, avgMbps)
			}

			progress := math.Min(float64(totalBytes)/float64(25*1024*1024), 1.0)
			done := false

			if elapsed >= maxDuration {
				done = true
			}
			if elapsed >= minDuration && len(instSpeeds) >= 3 {
				sum, maxDev := 0.0, 0.0
				for _, s := range instSpeeds {
					sum += s
				}
				mean := sum / float64(len(instSpeeds))
				for _, s := range instSpeeds {
					dev := math.Abs(s - mean)
					if dev > maxDev {
						maxDev = dev
					}
				}
				if mean > 0 && maxDev/mean < 0.03 {
					done = true
				}
			}

			var result *SpeedResult
			if done && len(instSpeeds) > 0 {
				sum := 0.0
				for _, s := range instSpeeds {
					sum += s
				}
				result = &SpeedResult{
					DownloadMbps:  math.Round(sum/float64(len(instSpeeds))*10) / 10,
					BytesReceived: totalBytes,
					Duration:      elapsed,
					Timestamp:     time.Now(),
				}
			}

			updates <- SpeedSample{
				Speed:     math.Round(avgMbps*10) / 10,
				Progress:  progress,
				BytesRecv: totalBytes,
				Elapsed:   elapsed,
				Done:      done,
				Result:    result,
			}

			if done {
				return
			}
		}
	}
}

func pickURL(ctx context.Context) string {
	for _, u := range defaultURLs {
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, u, nil)
		if err != nil {
			continue
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return u
		}
	}
	return ""
}

type reachableURLError struct{}

func (e *reachableURLError) Error() string { return "no reachable test URL" }

func errNoReachableURL() *reachableURLError { return &reachableURLError{} }
