package tui

import (
	"testing"
	"time"

	"fast-go/internal/speedtest"
)

func sampleModel(mode ViewMode, state State) Model {
	m := InitialModel(mode == ViewModeMini)
	m.width = 80
	m.height = 24
	m.state = state
	m.viewMode = mode
	m.actualSpeed = 245.3
	m.displaySpeed = AnimFloat{Current: 245.3, Target: 245.3}
	m.displayGauge = AnimFloat{Current: 0.73, Target: 0.73}
	m.bytesRecv = 18 * 1024 * 1024
	m.elapsed = 7 * time.Second
	for i := 0; i < 40; i++ {
		m.history = append(m.history, 100+float64(i%13)*15)
	}
	if state == StateDone {
		m.displayGauge = AnimFloat{Current: 1.0, Target: 1.0}
		m.result = &speedtest.SpeedResult{
			DownloadMbps:  245.3,
			BytesReceived: 25 * 1024 * 1024,
			Duration:      8200 * time.Millisecond,
			Timestamp:     time.Now(),
		}
	}
	return m
}

func TestViewSnapshot(t *testing.T) {
	for _, tc := range []struct {
		name  string
		mode  ViewMode
		state State
	}{
		{"normal-testing", ViewModeNormal, StateTesting},
		{"normal-done", ViewModeNormal, StateDone},
		{"mini-testing", ViewModeMini, StateTesting},
		{"mini-done", ViewModeMini, StateDone},
	} {
		m := sampleModel(tc.mode, tc.state)
		t.Logf("=== %s ===\n%s\n", tc.name, m.View())
	}
}
