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
	"github.com/muesli/reflow/wordwrap"
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

// overlayMascot draws art rows flush-right at the top of the scroll
// region, only on rows short enough to leave him room — busy lines stay
// untouched and he just peeks over the empty ones.
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
	menu        []string
	menuIdx     int
	mascotOn    bool
	mascotFrame int
	peekTicks   int
	doubleTake  bool
	hasBlink    bool
	blinkDone   bool
	mascotOff   time.Time
	nextPeek    time.Time
	blinkAt     time.Time
	blinkEnd    time.Time
	blinking    bool
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

	m.menu = []string{"about", "work", "skills", "edu", "interests", "contact", "resume", "exit"}

	var boot []string
	for _, runs := range content.Boot {
		boot = append(boot, renderLine(runs))
	}
	boot = append(boot, "")
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
	wrapped := make([]string, 0, len(m.lines))
	for _, l := range m.lines {
		wrapped = append(wrapped, wordwrap.String(l, m.vp.Width))
	}
	m.vp.SetContent(strings.Join(wrapped, "\n"))
	m.vp.GotoBottom()
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, mascotTick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			m.vp.Width = msg.Width - 18
			if m.vp.Width < 20 {
				m.vp.Width = 20
			}
		}
		if msg.Height > 0 {
			m.height = msg.Height
			m.vp.Height = msg.Height - 4
			if m.vp.Height < 1 {
				m.vp.Height = 1
			}
		}
		m.refresh()
		return m, nil

	case mascotMsg:
		now := time.Now()
		if m.mascotOn {
			m.peekTicks++
			// eyes: center ~0.8s -> left -> right -> center
			switch {
			case m.peekTicks < 2:
				m.mascotFrame = 0
			case m.peekTicks < 4:
				m.mascotFrame = 1
			case m.peekTicks < 6:
				m.mascotFrame = 2
			default:
				m.mascotFrame = 0
			}
			if m.blinking && now.After(m.blinkEnd) {
				m.blinking = false
			}
			if m.hasBlink && !m.blinkDone && now.After(m.blinkAt) {
				m.blinking = true
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
			m.mascotFrame = 0
			m.peekTicks = 0
			hold := 2600 + rand.Intn(1600)
			m.mascotOff = now.Add(time.Duration(hold) * time.Millisecond)
			m.doubleTake = rand.Intn(100) < 20
			m.hasBlink = rand.Intn(100) < 60
			m.blinkDone = false
			m.blinking = false
			if m.hasBlink {
				m.blinkAt = now.Add(time.Duration(2200+rand.Intn(hold-2400)) * time.Millisecond)
				m.blinkEnd = m.blinkAt.Add(180 * time.Millisecond)
			}
		}
		return m, mascotTick()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			raw := strings.TrimSpace(m.ti.Value())
			menuPick := raw == ""
			if menuPick {
				if m.menuIdx >= 0 && m.menuIdx < len(m.menu) {
					raw = m.menu[m.menuIdx]
				} else {
					return m, nil
				}
			}
			m.lines = append(m.lines, echoLine(raw))
			m.ti.SetValue("")
			m.history = append(m.history, raw)
			m.hIdx = len(m.history)
			low := strings.ToLower(raw)
			if low == "clear" || low == "cls" {
				m.lines = nil
				m.refresh()
				return m, nil
			}
			out, quit := dispatch(raw)
			if menuPick {
				// page mode: the pane shows just this section
				m.lines = out
			} else {
				m.lines = append(m.lines, out...)
			}
			m.refresh()
			if menuPick {
				m.vp.GotoTop()
			}
			if quit {
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil

		case tea.KeyUp:
			if m.menuIdx > 0 {
				m.menuIdx--
			}
			return m, nil
		case tea.KeyDown:
			if m.menuIdx < len(m.menu)-1 {
				m.menuIdx++
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
	hdr := " " + stAccent.Render("L U K E   C O T N O I R") + stDim.Render("   ·   ssh guest terminal") +
		"\n" + stFaint.Render(strings.Repeat("─", m.vp.Width+17))

	var mb strings.Builder
	for i, it := range m.menu {
		if i == m.menuIdx {
			mb.WriteString(stAccent.Render(" » ") + stBright.Render(it))
		} else {
			mb.WriteString(stFaint.Render("   ") + stDim.Render(it))
		}
		mb.WriteString("\n")
	}
	menuPane := lipgloss.NewStyle().Width(16).Render(mb.String())
	sep := stFaint.Render(strings.Repeat("│\n", m.vp.Height))

	body := m.vp.View()
	if m.mascotOn {
		arts := []string{content.Mascot, content.MascotLeft, content.MascotRight, content.MascotBlink}
		art := content.Mascot
		idx := m.mascotFrame
		if m.blinking {
			idx = 3
		}
		if idx >= 0 && idx < len(arts) && arts[idx] != "" {
			art = arts[idx]
		}
		body = overlayMascot(body, m.vp.Width, art)
	}
	return hdr + "\n" + lipgloss.JoinHorizontal(lipgloss.Top, menuPane, sep, body) + "\n" + m.ti.View()
}
