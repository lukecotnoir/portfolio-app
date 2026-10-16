package main

import (
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	stBody    = lipgloss.NewStyle().Foreground(lipgloss.Color("#d6cfbd"))
	stDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("#6f6a5c"))
	stFaint   = lipgloss.NewStyle().Foreground(lipgloss.Color("#45413a"))
	stAccent  = lipgloss.NewStyle().Foreground(lipgloss.Color("#d9a05b"))
	stBright  = lipgloss.NewStyle().Foreground(lipgloss.Color("#f2ecdd"))
	stRed     = lipgloss.NewStyle().Foreground(lipgloss.Color("#b3492e"))
	stKey     = lipgloss.NewStyle().Foreground(lipgloss.Color("#d9a05b"))
	stArt     = lipgloss.NewStyle().Foreground(lipgloss.Color("#8a6435"))
	stArtDark = lipgloss.NewStyle().Foreground(lipgloss.Color("#d6cfbd"))
	stBanner  = lipgloss.NewStyle().Foreground(lipgloss.Color("#f2ecdd"))

	styleFor = map[string]*lipgloss.Style{
		"body": &stBody, "dim": &stDim, "faint": &stFaint, "accent": &stAccent,
		"bright": &stBright, "red": &stRed, "key": &stKey, "art": &stArt,
		"artdark": &stArtDark, "banner": &stBanner,
	}
)

func renderLine(runs []Run) string {
	var b strings.Builder
	for _, r := range runs {
		st := styleFor[r.S]
		if st == nil {
			st = &stBody
		}
		b.WriteString(st.Render(r.T))
	}
	return b.String()
}

func echoLine(cmd string) string {
	return stAccent.Render("guest@lcotnoir") + stDim.Render(":~$ ") + stBright.Render(cmd)
}

type mascotMsg struct{}

func mascotTick() tea.Cmd {
	return tea.Tick(400*time.Millisecond, func(time.Time) tea.Msg { return mascotMsg{} })
}

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;:]*[a-zA-Z]")

func visibleWidth(s string) int {
	return len([]rune(ansiRe.ReplaceAllString(s, "")))
}

// overlayMascot draws art rows flush-right at the top of the view, only on
// rows short enough to leave him room — busy lines stay untouched and he
// just peeks over the empty ones.
func overlayMascot(view string, width int, art string) string {
	rows := strings.Split(art, "\n")
	mw := 0
	for _, r := range rows {
		if w := len([]rune(r)); w > mw {
			mw = w
		}
	}
	col := width - mw - 2
	if col < 5 {
		return view
	}
	lines := strings.Split(view, "\n")
	for i, r := range rows {
		if i >= len(lines) {
			break
		}
		vw := visibleWidth(lines[i])
		if vw > col {
			continue
		}
		lines[i] += strings.Repeat(" ", col-vw) + stDim.Render(r)
	}
	return strings.Join(lines, "\n")
}

type model struct {
	vp          viewport.Model
	ti          textinput.Model
	user        string
	history     []string
	hIdx        int
	lines       []string
	height      int
	quitting    bool
	mascotOn    bool
	mascotBlink bool
	doubleTake  bool
	hasBlink    bool
	blinkDone   bool
	mascotOff   time.Time
	nextPeek    time.Time
	blinkAt     time.Time
	blinkEnd    time.Time
}

func newModel(user string, authed bool) model {
	ti := textinput.New()
	ti.Prompt = stAccent.Render("guest@lcotnoir") + stDim.Render(":~$ ")
	ti.TextStyle = stBright
	ti.Cursor.Style = stAccent
	ti.Focus()

	vp := viewport.New(80, 20)

	m := model{vp: vp, ti: ti, user: user, hIdx: -1}
	m.nextPeek = time.Now().Add(6 * time.Second)

	var boot []string
	for _, runs := range content.Boot {
		boot = append(boot, renderLine(runs))
	}
	boot = append(boot,
		"",
		stBanner.Render(content.Banner),
		"",
	)
	for _, runs := range content.AfterBanner {
		boot = append(boot, renderLine(runs))
	}
	boot = append(boot, stDim.Render("web version: https://"+host()+ " — you are inside the ssh build"))
	m.lines = boot
	m.refresh()
	return m
}

func host() string {
	if h := envHost(); h != "" {
		return h
	}
	return "lcotnoir.dev"
}

func (m *model) refresh() {
	m.vp.SetContent(strings.Join(m.lines, "\n"))
	m.vp.GotoBottom()
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, mascotTick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			m.vp.Width = msg.Width
		}
		if msg.Height > 0 {
			m.height = msg.Height
			m.vp.Height = msg.Height - 2
			if m.vp.Height < 1 {
				m.vp.Height = 1
			}
		}
		m.refresh()
		return m, nil

	case mascotMsg:
		now := time.Now()
		if m.mascotOn {
			if m.mascotBlink && now.After(m.blinkEnd) {
				m.mascotBlink = false
			}
			if m.hasBlink && !m.blinkDone && now.After(m.blinkAt) {
				m.mascotBlink = true
				m.blinkDone = true
			}
			if now.After(m.mascotOff) {
				m.mascotOn = false
				if m.doubleTake {
					m.nextPeek = now.Add(500 * time.Millisecond)
				} else {
					m.nextPeek = now.Add(time.Duration(14+rand.Intn(24)) * time.Second)
				}
			}
		} else if now.After(m.nextPeek) {
			m.mascotOn = true
			hold := 1800 + rand.Intn(1800)
			m.mascotOff = now.Add(time.Duration(hold) * time.Millisecond)
			m.doubleTake = rand.Intn(100) < 20
			m.hasBlink = rand.Intn(100) < 60
			m.blinkDone = false
			m.mascotBlink = false
			if m.hasBlink {
				m.blinkAt = now.Add(time.Duration(500+rand.Intn(hold-700)) * time.Millisecond)
				m.blinkEnd = m.blinkAt.Add(180 * time.Millisecond)
			}
		}
		return m, mascotTick()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			raw := strings.TrimSpace(m.ti.Value())
			m.lines = append(m.lines, echoLine(raw))
			m.ti.SetValue("")
			if raw == "" {
				m.refresh()
				return m, nil
			}
			m.history = append(m.history, raw)
			m.hIdx = len(m.history)
			low := strings.ToLower(raw)
			if low == "clear" || low == "cls" {
				m.lines = nil
				m.refresh()
				return m, nil
			}
			out, quit := dispatch(raw)
			m.lines = append(m.lines, out...)
			m.refresh()
			if quit {
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil

		case tea.KeyUp:
			if m.hIdx > 0 {
				m.hIdx--
				m.ti.SetValue(m.history[m.hIdx])
			}
			return m, nil
		case tea.KeyDown:
			if m.hIdx < len(m.history) {
				m.hIdx++
				if m.hIdx == len(m.history) {
					m.ti.SetValue("")
				} else {
					m.ti.SetValue(m.history[m.hIdx])
				}
			}
			return m, nil
		case tea.KeyCtrlL:
			m.lines = nil
			m.refresh()
			return m, nil
		case tea.KeyTab:
			if v := complete(m.ti.Value()); v != "" {
				m.ti.SetValue(v)
			}
			return m, nil
		case tea.KeyCtrlC:
			m.lines = append(m.lines, echoLine(m.ti.Value()+"^C"))
			m.ti.SetValue("")
			m.refresh()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.ti, cmd = m.ti.Update(msg)
	var vcmd tea.Cmd
	m.vp, vcmd = m.vp.Update(msg)
	return m, tea.Batch(cmd, vcmd)
}

func (m model) View() string {
	view := m.vp.View() + "\n" + m.ti.View()
	if m.mascotOn {
		art := content.Mascot
		if m.mascotBlink && content.MascotBlink != "" {
			art = content.MascotBlink
		}
		view = overlayMascot(view, m.vp.Width, art)
	}
	return view
}
