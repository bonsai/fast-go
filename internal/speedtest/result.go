package speedtest

import "time"

type SpeedResult struct {
	DownloadMbps    float64       `json:"download_mbps"`
	UploadMbps      float64       `json:"upload_mbps"`
	LatencyMs       float64       `json:"latency_ms"`
	LoadedLatencyMs float64       `json:"loaded_latency_ms"`
	BytesReceived   int64         `json:"bytes_received"`
	BytesSent       int64         `json:"bytes_sent"`
	Duration        time.Duration `json:"duration"`
	Timestamp       time.Time     `json:"timestamp"`
}

type SpeedSample struct {
	Speed     float64
	Progress  float64
	BytesRecv int64
	Elapsed   time.Duration
	Done      bool
	Error     error
	Result    *SpeedResult
}
