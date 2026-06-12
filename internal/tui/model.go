package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"fast-go/internal/speedtest"
	"fast-go/internal/util"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type State int

const (
	StateIdle State = iota
	StateTesting
	StateDone
	StateError
)

type ViewMode int

const (
	ViewModeCompact ViewMode = iota
	ViewModeDS
	ViewModeWide
	viewModeCount
)

func (vm ViewMode) String() string {
	switch vm {
	case ViewModeCompact:
		return "GB"
	case ViewModeDS:
		return "DS"
	case ViewModeWide:
		return "WIDE"
	}
	return "WIDE"
}

type Model struct {
	state    State
	width    int
	height   int
	errMsg   string

	actualSpeed float64
	progress    float64
	bytesRecv   int64
	elapsed     time.Duration
	result      *speedtest.SpeedResult

	displaySpeed AnimFloat
	displayGauge AnimFloat

	mood    Mood
	palette Palette

	showDetails bool
	moodIndex   int

	stopTest context.CancelFunc
	speedCh  chan speedtest.SpeedSample

	lat      float64
	lon      float64
	viewMode ViewMode
}

type speedSampleMsg speedtest.SpeedSample
type tickMsg struct{}
type startTestMsg struct {
	speedCh chan speedtest.SpeedSample
	cancel  context.CancelFunc
}

type locationMsg struct {
	lat float64
	lon float64
}

func InitialModel() Model {
	mood := RandomMood()
	return Model{
		state:       StateIdle,
		mood:        mood,
		palette:     mood.GeneratePalette(),
		moodIndex:   int(mood),
		showDetails: false,
		width:       80,
		height:      24,
		viewMode:    ViewModeWide,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.tickCmd(),
		m.startTestCmd(),
		m.fetchLocationCmd(),
	)
}

func (m Model) startTestCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		ch := make(chan speedtest.SpeedSample, 100)
		go speedtest.Run(ctx, ch)
		return startTestMsg{speedCh: ch, cancel: cancel}
	}
}

func (m Model) nextSampleCmd() tea.Cmd {
	if m.speedCh == nil {
		return nil
	}
	return func() tea.Msg {
		sample, ok := <-m.speedCh
		if !ok {
			return nil
		}
		return speedSampleMsg(sample)
	}
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second/30, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m Model) fetchLocationCmd() tea.Cmd {
	return func() tea.Msg {
		lat, lon, err := util.FetchLocation()
		if err != nil {
			return nil
		}
		return locationMsg{lat: lat, lon: lon}
	}
}

func (m Model) viewWidth() int {
	w := int(float64(m.width) * 0.8)
	if w < 30 {
		w = 30
	}
	switch m.viewMode {
	case ViewModeCompact:
		if w > 34 {
			w = 34
		}
	case ViewModeDS:
		if w > 50 {
			w = 50
		}
	}
	return w
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.stopTest != nil {
				m.stopTest()
			}
			return m, tea.Quit

		case "r":
			if m.stopTest != nil {
				m.stopTest()
			}
			m.state = StateIdle
			m.actualSpeed = 0
			m.progress = 0
			m.bytesRecv = 0
			m.elapsed = 0
			m.result = nil
			m.displaySpeed = AnimFloat{}
			m.displayGauge = AnimFloat{}
			m.speedCh = nil
			mood := RandomMood()
			m.mood = mood
			m.moodIndex = int(mood)
			m.palette = mood.GeneratePalette()
			return m, m.startTestCmd()

		case "m":
			m.moodIndex = (m.moodIndex + 1) % int(moodCount)
			m.mood = Mood(m.moodIndex)
			m.palette = m.mood.GeneratePalette()

		case "d":
			m.showDetails = !m.showDetails

		case "v":
			m.viewMode = ViewMode((int(m.viewMode) + 1) % int(viewModeCount))

		case "s":
			if m.result != nil {
				f, err := os.OpenFile("fast-results.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err == nil {
					fmt.Fprintf(f, `{"download_mbps":%.1f,"bytes_received":%d,"duration_ms":%d,"timestamp":"%s","lat":%.4f,"lon":%.4f}`+"\n",
						m.result.DownloadMbps, m.result.BytesReceived,
						m.result.Duration.Milliseconds(), m.result.Timestamp.Format(time.RFC3339),
						m.lat, m.lon)
					f.Close()
				}
			}
		}

	case startTestMsg:
		m.state = StateTesting
		m.speedCh = msg.speedCh
		m.stopTest = msg.cancel
		return m, m.nextSampleCmd()

	case locationMsg:
		m.lat = msg.lat
		m.lon = msg.lon
		return m, nil

	case speedSampleMsg:
		s := speedtest.SpeedSample(msg)
		if s.Error != nil {
			m.state = StateError
			m.errMsg = s.Error.Error()
			return m, nil
		}

		m.actualSpeed = s.Speed
		m.progress = s.Progress
		m.bytesRecv = s.BytesRecv
		m.elapsed = s.Elapsed

		if s.Done {
			m.state = StateDone
			m.result = s.Result
			if m.result != nil {
				m.actualSpeed = m.result.DownloadMbps
				m.progress = 1.0
			}
			return m, nil
		}

		return m, m.nextSampleCmd()

	case tickMsg:
		dt := 1.0 / 30.0
		m.displaySpeed.Update(dt)
		m.displayGauge.Update(dt)

		if m.state == StateTesting {
			m.displaySpeed.Target = m.actualSpeed
			m.displayGauge.Target = m.progress
		}

		return m, m.tickCmd()
	}

	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	prim := lipgloss.Color(m.palette.Primary)
	sec := lipgloss.Color(m.palette.Secondary)
	acc := lipgloss.Color(m.palette.Accent)

	switch m.state {
	case StateIdle:
		fmt.Fprint(&b, "\n\n\n\n")
		title := lipgloss.NewStyle().Bold(true).Foreground(prim).Width(m.viewWidth()).Align(lipgloss.Center).
			Render("FAST-GO")
		sub := lipgloss.NewStyle().Foreground(sec).Width(m.viewWidth()).Align(lipgloss.Center).
			Render("Local Speed Test")
		fmt.Fprintln(&b, title)
		fmt.Fprintln(&b, sub)
		fmt.Fprintln(&b)
		loading := lipgloss.NewStyle().Foreground(acc).Width(m.viewWidth()).Align(lipgloss.Center).
			Render("Initializing...")
		fmt.Fprintln(&b, loading)

	case StateTesting:
		m.renderTest(&b, prim, sec, acc)

	case StateDone:
		m.renderTest(&b, prim, sec, acc)
		m.renderResult(&b, prim, sec, acc)

	case StateError:
		fmt.Fprint(&b, "\n\n\n\n")
		errStr := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF0000")).
			Width(m.width).Align(lipgloss.Center).Render("✗ ERROR")
		msg := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666")).
			Width(m.width).Align(lipgloss.Center).Render(m.errMsg)
		fmt.Fprintln(&b, errStr)
		fmt.Fprintln(&b, msg)
		fmt.Fprintln(&b)
		hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
			Width(m.width).Align(lipgloss.Center).Render("[r] Retry  [q] Quit")
		fmt.Fprintln(&b, hint)
	}

	return b.String()
}

func (m Model) renderTest(b *strings.Builder, prim, sec, acc lipgloss.Color) {
	fmt.Fprint(b, "\n\n")

	moodLabel := lipgloss.NewStyle().Foreground(acc).Bold(true).
		Width(m.viewWidth()).Align(lipgloss.Center).
		Render("◆ " + m.mood.String() + " ◆  [" + m.viewMode.String() + "]")
	fmt.Fprintln(b, moodLabel)
	fmt.Fprintln(b)

	speedStr := util.FormatMbps(m.displaySpeed.Current)
	speedStyle := lipgloss.NewStyle().Bold(true).Foreground(prim)
	bigNum := speedStyle.Render(speedStr)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(prim).
		Padding(0, 2).
		Width(m.viewWidth() - 6).
		Align(lipgloss.Center)

	boxedNum := borderStyle.Render(bigNum)
	fmt.Fprintln(b, boxedNum)
	fmt.Fprintln(b)

	gaugeWidth := m.viewWidth() - 10
	if gaugeWidth < 10 {
		gaugeWidth = 10
	}
	filled := int(m.displayGauge.Current * float64(gaugeWidth))
	if filled > gaugeWidth {
		filled = gaugeWidth
	}
	if filled < 0 {
		filled = 0
	}
	// Row 1: leader bar (slightly ahead for visual pop)
	row1Filled := filled + 2
	if row1Filled > gaugeWidth {
		row1Filled = gaugeWidth
	}
	row1 := strings.Repeat("█", row1Filled) + strings.Repeat("░", gaugeWidth-row1Filled)
	// Row 2: actual progress
	row2 := strings.Repeat("█", filled) + strings.Repeat("░", gaugeWidth-filled)

	gaugeRow1 := lipgloss.NewStyle().Foreground(prim).Render(row1)
	gaugeRow2 := lipgloss.NewStyle().Foreground(sec).Render(row2)
	pctStr := lipgloss.NewStyle().Foreground(acc).Render(fmt.Sprintf("%3.0f%%", m.displayGauge.Current*100))

	fmt.Fprintf(b, "  %s  %s\n", gaugeRow1, pctStr)
	fmt.Fprintf(b, "  %s\n", gaugeRow2)

	progressStr := fmt.Sprintf("  Testing... %s / %s  (%s)",
		util.FormatBytes(m.bytesRecv),
		util.FormatBytes(int64(25*1024*1024)),
		util.FormatDuration(m.elapsed))
	progressLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(progressStr)
	fmt.Fprintln(b, progressLine)

	fmt.Fprintln(b)
	miniDown := lipgloss.NewStyle().Foreground(prim).Render(fmt.Sprintf("↓ %s", util.FormatMbps(m.actualSpeed)))
	miniUp := lipgloss.NewStyle().Foreground(sec).Render("↑ --")
	miniPing := lipgloss.NewStyle().Foreground(acc).Render("ping --")
	miniLine := lipgloss.NewStyle().Width(m.viewWidth()).Align(lipgloss.Center).
		Render(fmt.Sprintf("%s    %s    %s", miniDown, miniUp, miniPing))
	fmt.Fprintln(b, miniLine)
}

func (m Model) renderResult(b *strings.Builder, prim, sec, acc lipgloss.Color) {
	fmt.Fprintln(b)
	complete := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00")).
		Width(m.viewWidth()).Align(lipgloss.Center).
		Render("✓ Test Complete")
	fmt.Fprintln(b, complete)

	if m.result != nil {
		fmt.Fprintln(b)
		avgStr := lipgloss.NewStyle().Foreground(sec).
			Width(m.viewWidth()).Align(lipgloss.Center).
			Render(fmt.Sprintf("Average: %s  |  Rating: %s",
				util.FormatMbps(m.result.DownloadMbps),
				util.SpeedRating(m.result.DownloadMbps)))
		fmt.Fprintln(b, avgStr)
	}

	if m.showDetails && m.result != nil {
		fmt.Fprintln(b)
		detail := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).
			Width(m.viewWidth()).Align(lipgloss.Center).
			Render(fmt.Sprintf("Download: %s    Data: %s    Duration: %s",
				util.FormatMbps(m.result.DownloadMbps),
				util.FormatBytes(m.result.BytesReceived),
				util.FormatDuration(m.result.Duration)))
		fmt.Fprintln(b, detail)
	}

	fmt.Fprintln(b)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Width(m.viewWidth()).Align(lipgloss.Center)

	hints := []string{"[r] Retry", "[m] Mood", "[v] View", "[d] Details", "[s] Save", "[q] Quit"}
	fmt.Fprintln(b, hintStyle.Render(strings.Join(hints, "  ")))
}
