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
	ViewModeNormal ViewMode = iota
	ViewModeMini
	viewModeCount
)

func (vm ViewMode) String() string {
	if vm == ViewModeMini {
		return "MINI"
	}
	return "NORMAL"
}

const (
	historyCap = 240
	dimColor   = "#666666"
	faintColor = "#444444"
	okColor    = "#34D399"
)

type Model struct {
	state  State
	width  int
	height int
	errMsg string

	actualSpeed float64
	progress    float64
	bytesRecv   int64
	elapsed     time.Duration
	result      *speedtest.SpeedResult

	displaySpeed AnimFloat
	displayGauge AnimFloat
	history      []float64
	frame        int

	mood    Mood
	palette Palette

	showDetails bool
	moodIndex   int
	savedAt     time.Time

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

func InitialModel(mini bool) Model {
	mood := RandomMood()
	mode := ViewModeNormal
	if mini {
		mode = ViewModeMini
	}
	return Model{
		state:       StateIdle,
		mood:        mood,
		palette:     mood.GeneratePalette(),
		moodIndex:   int(mood),
		showDetails: false,
		width:       80,
		height:      24,
		viewMode:    mode,
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

// panelWidth is the content width for the NORMAL layout.
func (m Model) panelWidth() int {
	w := m.width - 8
	if w > 64 {
		w = 64
	}
	if w < 36 {
		w = 36
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
			m.history = nil
			m.savedAt = time.Time{}
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
					m.savedAt = time.Now()
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

		m.history = append(m.history, s.Speed)
		if len(m.history) > historyCap {
			m.history = m.history[len(m.history)-historyCap:]
		}

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
		m.frame++
		m.displaySpeed.Update(dt)
		m.displayGauge.Update(dt)

		switch m.state {
		case StateTesting:
			m.displaySpeed.Target = m.actualSpeed
			m.displayGauge.Target = m.progress
		case StateDone:
			m.displaySpeed.Target = m.actualSpeed
			m.displayGauge.Target = 1.0
		}

		return m, m.tickCmd()
	}

	return m, nil
}

// speedParts splits a Mbps value into the numeric string and its unit.
func speedParts(v float64) (string, string) {
	if v >= 1000 {
		return fmt.Sprintf("%.2f", v/1000), "Gbps"
	}
	if v >= 1 {
		return fmt.Sprintf("%.1f", v), "Mbps"
	}
	return fmt.Sprintf("%.2f", v), "Mbps"
}

func (m Model) View() string {
	var content string

	switch m.state {
	case StateIdle:
		content = m.viewIdle()
	case StateError:
		content = m.viewError()
	default:
		if m.viewMode == ViewModeMini {
			content = m.viewMini()
		} else {
			content = m.viewNormal()
		}
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) viewIdle() string {
	prim := lipgloss.Color(m.palette.Primary)
	sec := lipgloss.Color(m.palette.Secondary)
	acc := lipgloss.Color(m.palette.Accent)

	title := lipgloss.NewStyle().Bold(true).Foreground(prim).Render("⚡ FAST-GO")
	sub := lipgloss.NewStyle().Foreground(sec).Render("LOCAL SPEED TEST")
	loading := lipgloss.NewStyle().Foreground(acc).Render(spinner(m.frame) + " connecting…")

	return lipgloss.JoinVertical(lipgloss.Center, title, sub, "", loading)
}

func (m Model) viewError() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5555")).Render("✗ ERROR")
	msg := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8888")).Render(m.errMsg)
	hint := m.renderKeys([][2]string{{"r", "retry"}, {"q", "quit"}})

	return lipgloss.JoinVertical(lipgloss.Center, title, msg, "", hint)
}

func (m Model) viewNormal() string {
	pw := m.panelWidth()
	prim := m.palette.Primary
	sec := m.palette.Secondary
	acc := m.palette.Accent

	center := lipgloss.NewStyle().Width(pw).Align(lipgloss.Center)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color(dimColor))

	// Header: name on the left, mood + mode on the right.
	left := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(prim)).Render("⚡ FAST-GO")
	right := lipgloss.NewStyle().Foreground(lipgloss.Color(acc)).
		Render("◆ " + m.mood.String() + " · " + m.viewMode.String())
	gap := pw - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	header := left + strings.Repeat(" ", gap) + right
	rule := lipgloss.NewStyle().Foreground(lipgloss.Color(faintColor)).
		Render(strings.Repeat("─", pw))

	// Big speed readout.
	numStr, unit := speedParts(m.displaySpeed.Current)
	digits := center.Render(RenderBigDigits(numStr, prim, sec))
	unitLine := center.Render(
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(sec)).Render(unit) +
			dim.Render("  ·  ") +
			lipgloss.NewStyle().Foreground(lipgloss.Color(acc)).Render("↓ DOWNLOAD"))

	// Speed history sparkline.
	sparkW := pw - 12
	spark := ""
	if len(m.history) > 1 && sparkW > 4 {
		spark = center.Render(sparkline(m.history, sparkW, sec, prim))
	}

	// Gradient gauge.
	gaugeW := pw - 12
	if gaugeW < 10 {
		gaugeW = 10
	}
	pct := lipgloss.NewStyle().Foreground(lipgloss.Color(acc)).
		Render(fmt.Sprintf(" %3.0f%%", m.displayGauge.Current*100))
	gauge := center.Render(gradientBar(gaugeW, m.displayGauge.Current, prim, acc) + pct)

	// Status line.
	var status string
	if m.state == StateTesting {
		status = center.Render(dim.Render(fmt.Sprintf("%s %s / %s · %s",
			spinner(m.frame),
			util.FormatBytes(m.bytesRecv),
			util.FormatBytes(int64(25*1024*1024)),
			util.FormatDuration(m.elapsed))))
	} else if m.result != nil {
		status = center.Render(
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(okColor)).Render("✓ COMPLETE") +
				dim.Render("  ·  ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(sec)).
					Render(util.FormatMbps(m.result.DownloadMbps)+" avg") +
				dim.Render("  ·  ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color(acc)).
					Render(util.SpeedRating(m.result.DownloadMbps)))
	}

	rows := []string{header, rule, "", digits, "", unitLine}
	if spark != "" {
		rows = append(rows, "", spark)
	}
	rows = append(rows, "", gauge, status)

	if m.showDetails && m.result != nil {
		detail := center.Render(dim.Render(fmt.Sprintf("data %s · duration %s · %d ms/sample",
			util.FormatBytes(m.result.BytesReceived),
			util.FormatDuration(m.result.Duration),
			200)))
		rows = append(rows, detail)
	}

	rows = append(rows, "", m.footer(pw))

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) viewMini() string {
	prim := m.palette.Primary
	sec := m.palette.Secondary
	acc := m.palette.Accent
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color(dimColor))

	inner := 34

	// Line 1: speed + sparkline.
	speed := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(prim)).
		Render("↓ " + util.FormatMbps(m.displaySpeed.Current))
	spark := sparkline(m.history, 10, sec, prim)
	gap1 := inner - lipgloss.Width(speed) - lipgloss.Width(spark)
	if gap1 < 1 {
		gap1 = 1
	}
	line1 := speed + strings.Repeat(" ", gap1) + spark

	// Line 2: gauge + status.
	var tail string
	if m.state == StateTesting {
		tail = lipgloss.NewStyle().Foreground(lipgloss.Color(acc)).
			Render(fmt.Sprintf(" %3.0f%%", m.displayGauge.Current*100)) +
			dim.Render(" · "+util.FormatDuration(m.elapsed))
	} else if m.result != nil {
		tail = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(okColor)).Render(" ✓ ") +
			lipgloss.NewStyle().Foreground(lipgloss.Color(acc)).
				Render(util.SpeedRating(m.result.DownloadMbps))
	}
	gaugeW := inner - lipgloss.Width(tail)
	if gaugeW < 6 {
		gaugeW = 6
	}
	line2 := gradientBar(gaugeW, m.displayGauge.Current, prim, acc) + tail

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(prim)).
		Padding(0, 1).
		Render(line1 + "\n" + line2)

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(prim)).Render("⚡ FAST-GO ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color(acc)).Render("◆ "+m.mood.String())

	footer := m.footer(inner + 4)

	return lipgloss.JoinVertical(lipgloss.Center, title, box, footer)
}

// renderKeys renders [key] label pairs as a single dim hint line.
func (m Model) renderKeys(keys [][2]string) string {
	acc := lipgloss.NewStyle().Foreground(lipgloss.Color(m.palette.Accent))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color(dimColor))
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, acc.Render(k[0])+dim.Render(" "+k[1]))
	}
	return strings.Join(parts, dim.Render(" · "))
}

func (m Model) footer(width int) string {
	var keys [][2]string
	if m.viewMode == ViewModeMini {
		keys = [][2]string{
			{"r", "retry"}, {"v", "normal"}, {"s", "save"}, {"q", "quit"},
		}
	} else {
		keys = [][2]string{
			{"r", "retry"}, {"m", "mood"}, {"v", "mini"},
			{"d", "details"}, {"s", "save"}, {"q", "quit"},
		}
	}

	line := m.renderKeys(keys)
	if !m.savedAt.IsZero() && time.Since(m.savedAt) < 2*time.Second {
		line = lipgloss.NewStyle().Foreground(lipgloss.Color(okColor)).
			Render("✓ saved → fast-results.jsonl")
	}
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(line)
}
