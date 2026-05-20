package util

import (
	"fmt"
	"time"
)

func FormatMbps(v float64) string {
	if v >= 1000 {
		return fmt.Sprintf("%.2f Gbps", v/1000)
	}
	if v >= 1 {
		return fmt.Sprintf("%.1f Mbps", v)
	}
	return fmt.Sprintf("%.2f Mbps", v)
}

func FormatBytes(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	if b < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	}
	if b < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(b)/(1024*1024*1024))
}

func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0f ms", float64(d.Milliseconds()))
	}
	return fmt.Sprintf("%.1f s", d.Seconds())
}
