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
}

type speedSampleMsg speedtest.SpeedSample
type tickMsg struct{}
type startTestMsg struct {
	speedCh chan speedtest.SpeedSample
	cancel  context.CancelFunc
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
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.tickCmd(),
		m.startTestCmd(),
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

		case "s":
			if m.result != nil {
				f, err := os.Create(fmt.Sprintf("fast-result-%s.json", time.Now().Format("20060102-150405")))
				if err == nil {
					fmt.Fprintf(f, `{"download_mbps":%.1f,"bytes_received":%d,"duration_ms":%d,"timestamp":"%s"}`+"\n",
						m.result.DownloadMbps, m.result.BytesReceived,
						m.result.Duration.Milliseconds(), m.result.Timestamp.Format(time.RFC3339))
					f.Close()
				}
			}
		}

	case startTestMsg:
		m.state = StateTesting
		m.speedCh = msg.speedCh
		m.stopTest = msg.cancel
		return m, m.nextSampleCmd()

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
		title := lipgloss.NewStyle().Bold(true).Foreground(prim).Width(m.width).Align(lipgloss.Center).
			Render("FAST-GO")
		sub := lipgloss.NewStyle().Foreground(sec).Width(m.width).Align(lipgloss.Center).
			Render("Local Speed Test")
		fmt.Fprintln(&b, title)
		fmt.Fprintln(&b, sub)
		fmt.Fprintln(&b)
		loading := lipgloss.NewStyle().Foreground(acc).Width(m.width).Align(lipgloss.Center).
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
		Width(m.width).Align(lipgloss.Center).
		Render("◆ " + m.mood.String() + " ◆")
	fmt.Fprintln(b, moodLabel)
	fmt.Fprintln(b)

	speedStr := util.FormatMbps(m.displaySpeed.Current)
	speedStyle := lipgloss.NewStyle().Bold(true).Foreground(prim)
	bigNum := speedStyle.Render(speedStr)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(prim).
		Padding(0, 2).
		Width(m.width - 6).
		Align(lipgloss.Center)

	boxedNum := borderStyle.Render(bigNum)
	fmt.Fprintln(b, boxedNum)
	fmt.Fprintln(b)

	gaugeWidth := m.width - 10
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
	gauge := strings.Repeat("█", filled) + strings.Repeat("░", gaugeWidth-filled)
	gaugeStr := lipgloss.NewStyle().Foreground(sec).Render(gauge)
	pctStr := lipgloss.NewStyle().Foreground(acc).Render(fmt.Sprintf("%3.0f%%", m.displayGauge.Current*100))

	fmt.Fprintf(b, "  %s  %s\n", gaugeStr, pctStr)

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
	miniLine := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).
		Render(fmt.Sprintf("%s    %s    %s", miniDown, miniUp, miniPing))
	fmt.Fprintln(b, miniLine)
}

func (m Model) renderResult(b *strings.Builder, prim, sec, acc lipgloss.Color) {
	fmt.Fprintln(b)
	complete := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00")).
		Width(m.width).Align(lipgloss.Center).
		Render("✓ Test Complete")
	fmt.Fprintln(b, complete)

	if m.showDetails && m.result != nil {
		fmt.Fprintln(b)
		detail := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).
			Width(m.width).Align(lipgloss.Center).
			Render(fmt.Sprintf("Download: %s    Data: %s    Duration: %s",
				util.FormatMbps(m.result.DownloadMbps),
				util.FormatBytes(m.result.BytesReceived),
				util.FormatDuration(m.result.Duration)))
		fmt.Fprintln(b, detail)
	}

	fmt.Fprintln(b)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Width(m.width).Align(lipgloss.Center)

	hints := []string{"[r] Retry", "[m] Mood", "[d] Details", "[s] Save", "[q] Quit"}
	fmt.Fprintln(b, hintStyle.Render(strings.Join(hints, "  ")))
}
