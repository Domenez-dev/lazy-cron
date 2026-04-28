package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/domenez-dev/lazy-chrony/internal/cron"
	"github.com/domenez-dev/lazy-chrony/internal/styles"
)

type ListKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Top     key.Binding
	Bottom  key.Binding
	Add     key.Binding
	Edit    key.Binding
	Delete  key.Binding
	Toggle  key.Binding
	Refresh key.Binding
	Filter  key.Binding
	Quit    key.Binding
	Help    key.Binding
}

var DefaultListKeys = ListKeyMap{
	Up:      key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/up", "up")),
	Down:    key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/down", "down")),
	Top:     key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "top")),
	Bottom:  key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),
	Add:     key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
	Edit:    key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	Delete:  key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Toggle:  key.NewBinding(key.WithKeys(" ", "t"), key.WithHelp("space/t", "toggle")),
	Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	Filter:  key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
	Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
}

type FilterMode int

const (
	FilterAll FilterMode = iota
	FilterUser
	FilterSystem
	FilterEnabled
	FilterDisabled
)

type ListView struct {
	manager    *cron.Manager
	cursor     int
	offset     int
	width      int
	height     int
	filterMode FilterMode
	message    string
	messageOK  bool
	showHelp   bool
}

type (
	RefreshMsg struct{}
	StatusMsg  struct {
		Text string
		OK   bool
	}
	OpenAddMsg    struct{}
	OpenEditMsg   struct{ JobID int }
	ConfirmDelMsg struct{ JobID int }
)

func NewListView(mgr *cron.Manager) *ListView {
	return &ListView{manager: mgr}
}

func (l *ListView) SetSize(w, h int) { l.width = w; l.height = h }

func (l *ListView) SetMessage(msg string, ok bool) { l.message = msg; l.messageOK = ok }

func (l *ListView) Init() tea.Cmd { return nil }

func (l *ListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return l.handleKey(msg)
	case StatusMsg:
		l.message = msg.Text
		l.messageOK = msg.OK
	}
	return l, nil
}

func (l *ListView) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	jobs := l.visibleJobs()
	n := len(jobs)

	switch {
	case key.Matches(msg, DefaultListKeys.Quit):
		return l, tea.Quit
	case key.Matches(msg, DefaultListKeys.Up):
		if l.cursor > 0 {
			l.cursor--
		}
	case key.Matches(msg, DefaultListKeys.Down):
		if l.cursor < n-1 {
			l.cursor++
		}
	case key.Matches(msg, DefaultListKeys.Top):
		l.cursor = 0
	case key.Matches(msg, DefaultListKeys.Bottom):
		if n > 0 {
			l.cursor = n - 1
		}
	case key.Matches(msg, DefaultListKeys.Help):
		l.showHelp = !l.showHelp
	case key.Matches(msg, DefaultListKeys.Refresh):
		_ = l.manager.Load()
		l.message = "Refreshed"
		l.messageOK = true
		visible := l.visibleJobs()
		if l.cursor >= len(visible) {
			l.cursor = maxInt(0, len(visible)-1)
		}
	case key.Matches(msg, DefaultListKeys.Toggle):
		if n == 0 {
			break
		}
		job := jobs[l.cursor]
		if job.Source != cron.SourceUser {
			l.message = "Cannot toggle system jobs (requires root)"
			l.messageOK = false
			break
		}
		if err := l.manager.ToggleUserJob(job.ID); err != nil {
			l.message = fmt.Sprintf("Error: %s", err)
			l.messageOK = false
		} else {
			state := "enabled"
			if job.Disabled {
				state = "disabled"
			}
			l.message = fmt.Sprintf("Job %s", state)
			l.messageOK = true
		}
	case key.Matches(msg, DefaultListKeys.Add):
		return l, func() tea.Msg { return OpenAddMsg{} }
	case key.Matches(msg, DefaultListKeys.Edit):
		if n == 0 {
			break
		}
		job := jobs[l.cursor]
		if job.Source != cron.SourceUser {
			l.message = "Cannot edit system jobs (requires root)"
			l.messageOK = false
			break
		}
		return l, func() tea.Msg { return OpenEditMsg{JobID: job.ID} }
	case key.Matches(msg, DefaultListKeys.Delete):
		if n == 0 {
			break
		}
		job := jobs[l.cursor]
		if job.Source != cron.SourceUser {
			l.message = "Cannot delete system jobs (requires root)"
			l.messageOK = false
			break
		}
		return l, func() tea.Msg { return ConfirmDelMsg{JobID: job.ID} }
	case key.Matches(msg, DefaultListKeys.Filter):
		l.filterMode = (l.filterMode + 1) % 5
		l.cursor = 0
		l.message = fmt.Sprintf("Filter: %s", l.filterModeName())
		l.messageOK = true
	}

	listH := l.listHeight()
	if l.cursor < l.offset {
		l.offset = l.cursor
	}
	if l.cursor >= l.offset+listH {
		l.offset = l.cursor - listH + 1
	}
	return l, nil
}

func (l *ListView) filterModeName() string {
	switch l.filterMode {
	case FilterUser:
		return "user"
	case FilterSystem:
		return "system"
	case FilterEnabled:
		return "enabled"
	case FilterDisabled:
		return "disabled"
	default:
		return "all"
	}
}

func (l *ListView) visibleJobs() []*cron.Job {
	var out []*cron.Job
	for _, j := range l.manager.Jobs {
		switch l.filterMode {
		case FilterUser:
			if j.Source != cron.SourceUser {
				continue
			}
		case FilterSystem:
			if j.Source == cron.SourceUser {
				continue
			}
		case FilterEnabled:
			if j.Disabled {
				continue
			}
		case FilterDisabled:
			if !j.Disabled {
				continue
			}
		}
		out = append(out, j)
	}
	return out
}

func (l *ListView) listHeight() int {
	used := 5
	if l.showHelp {
		used += 7
	}
	if l.message != "" {
		used++
	}
	h := l.height - used
	if h < 1 {
		h = 1
	}
	return h
}

func (l *ListView) View() string {
	var b strings.Builder

	title := styles.TitleStyle.Render(" lazy-chrony ")
	version := styles.MutedStyle.Render("v" + styles.AppVersion)
	filter := styles.KeyStyle.Render("[/]") + styles.KeyDescStyle.Render(" filter:"+l.filterModeName())
	titleBar := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", version, "   ", filter)
	b.WriteString(titleBar + "\n")
	b.WriteString(l.renderHeader() + "\n")

	jobs := l.visibleJobs()
	listH := l.listHeight()

	if len(jobs) == 0 {
		b.WriteString(styles.MutedStyle.Render("  No cron jobs found. Press 'a' to add one."))
		for i := 1; i < listH; i++ {
			b.WriteString("\n")
		}
	} else {
		end := l.offset + listH
		if end > len(jobs) {
			end = len(jobs)
		}
		for i := l.offset; i < end; i++ {
			b.WriteString(l.renderRow(jobs[i], i == l.cursor) + "\n")
		}
		for i := end - l.offset; i < listH; i++ {
			b.WriteString("\n")
		}
	}

	if l.message != "" {
		var msgStyle lipgloss.Style
		if l.messageOK {
			msgStyle = styles.SuccessStyle
		} else {
			msgStyle = styles.ErrorStyle
		}
		b.WriteString(msgStyle.PaddingLeft(1).Render(l.message) + "\n")
	}

	if l.showHelp {
		b.WriteString(l.renderHelp() + "\n")
	}
	b.WriteString(l.renderHints())
	return b.String()
}

func (l *ListView) renderHeader() string {
	cols := fmt.Sprintf("%-4s %-2s %-8s %-22s %-36s %-14s",
		"#", "S", "SRC", "SCHEDULE", "COMMAND", "NEXT RUN")
	return styles.HeaderStyle.Width(l.width).Render(cols)
}

func (l *ListView) renderRow(j *cron.Job, selected bool) string {
	statusChar := "*"
	if j.Disabled {
		statusChar = "-"
	}

	srcText := "user"
	switch j.Source {
	case cron.SourceSystem:
		srcText = "system"
	case cron.SourceCrond:
		srcText = "cron.d"
	}

	// Plain row — no ANSI codes — for width-correct background rendering
	plain := fmt.Sprintf("%-4d %s %-8s %-22s %-36s %-14s",
		j.ID, statusChar, srcText,
		truncate(j.Schedule, 22),
		truncate(j.Command, 36),
		j.NextRunStr(),
	)

	if selected {
		return styles.SelectedRowStyle.Width(l.width).Render(plain)
	}
	if j.Disabled {
		return styles.MutedStyle.Width(l.width).Render(plain)
	}

	// Normal: compose with inline ANSI colors
	statusStyled := styles.EnabledStyle.Render(statusChar)
	var srcStyled string
	switch j.Source {
	case cron.SourceUser:
		srcStyled = styles.UserStyle.Render(fmt.Sprintf("%-8s", "user"))
	case cron.SourceSystem:
		srcStyled = styles.SystemStyle.Render(fmt.Sprintf("%-8s", "system"))
	case cron.SourceCrond:
		srcStyled = styles.CrondStyle.Render(fmt.Sprintf("%-8s", "cron.d"))
	}
	idStr := fmt.Sprintf("%-4d ", j.ID)
	restStr := fmt.Sprintf(" %-22s %-36s %-14s",
		truncate(j.Schedule, 22),
		truncate(j.Command, 36),
		j.NextRunStr(),
	)
	return styles.NormalRowStyle.Render(idStr) + statusStyled + " " + srcStyled + styles.NormalRowStyle.Render(restStr)
}

func (l *ListView) renderHelp() string {
	keys := []struct{ k, d string }{
		{"j/k", "navigate"}, {"g/G", "top/bottom"}, {"a", "add job"},
		{"e", "edit job"}, {"d", "delete job"}, {"space", "toggle enable"},
		{"r", "refresh"}, {"/", "cycle filter"}, {"?", "toggle help"}, {"q", "quit"},
	}
	var parts []string
	for _, kd := range keys {
		parts = append(parts, styles.KeyStyle.Render(kd.k)+" "+styles.KeyDescStyle.Render(kd.d))
	}
	content := strings.Join(parts[:5], "  ") + "\n" + strings.Join(parts[5:], "  ")
	return styles.HelpBarStyle.Width(l.width).Render(content)
}

func (l *ListView) renderHints() string {
	hints := []string{
		styles.KeyStyle.Render("a") + styles.KeyDescStyle.Render(" add"),
		styles.KeyStyle.Render("e") + styles.KeyDescStyle.Render(" edit"),
		styles.KeyStyle.Render("d") + styles.KeyDescStyle.Render(" del"),
		styles.KeyStyle.Render("space") + styles.KeyDescStyle.Render(" toggle"),
		styles.KeyStyle.Render("?") + styles.KeyDescStyle.Render(" help"),
		styles.KeyStyle.Render("q") + styles.KeyDescStyle.Render(" quit"),
	}
	return styles.HelpBarStyle.Width(l.width).Render(strings.Join(hints, "  "))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "~"
}

func stripAnsi(s string) string {
	var out strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
