package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/domenez-dev/lazy-chrony/internal/cron"
	"github.com/domenez-dev/lazy-chrony/internal/styles"
)

type FormView struct {
	manager *cron.Manager
	editID  int
	inputs  []textinput.Model
	focused int
	width   int
	height  int
	errMsg  string
	isEdit  bool
}

const (
	fieldSchedule = 0
	fieldCommand  = 1
	fieldComment  = 2
)

var fieldLabels = []string{"Schedule", "Command", "Comment (optional)"}
var fieldHints = []string{
	"e.g.  */5 * * * *   or  @daily  or  @reboot",
	"e.g.  /usr/bin/backup.sh >> /var/log/backup.log 2>&1",
	"e.g.  nightly backup",
}

var presets = []struct{ label, value string }{
	{"@reboot", "@reboot"},
	{"@hourly", "@hourly"},
	{"@daily", "@daily"},
	{"@weekly", "@weekly"},
	{"@monthly", "@monthly"},
	{"every 5m", "*/5 * * * *"},
	{"every 15m", "*/15 * * * *"},
	{"every 30m", "*/30 * * * *"},
	{"midnight", "0 0 * * *"},
	{"noon", "0 12 * * *"},
}

type DoneFormMsg struct {
	Cancelled bool
	Schedule  string
	Command   string
	Comment   string
	EditID    int
}

func NewFormView(mgr *cron.Manager, editJobID int) *FormView {
	f := &FormView{
		manager: mgr,
		editID:  editJobID,
		isEdit:  editJobID != 0,
	}
	f.inputs = make([]textinput.Model, 3)
	for i := range f.inputs {
		t := textinput.New()
		t.Prompt = ""
		t.CharLimit = 256
		f.inputs[i] = t
	}
	f.inputs[fieldSchedule].Placeholder = "*/5 * * * *"
	f.inputs[fieldCommand].Placeholder = "/path/to/script.sh"
	f.inputs[fieldComment].Placeholder = "optional description"

	if editJobID != 0 {
		for _, j := range mgr.Jobs {
			if j.ID == editJobID {
				f.inputs[fieldSchedule].SetValue(j.Schedule)
				f.inputs[fieldCommand].SetValue(j.Command)
				f.inputs[fieldComment].SetValue(j.Comment)
				break
			}
		}
	}
	f.inputs[0].Focus()
	return f
}

func (f *FormView) SetSize(w, h int) { f.width = w; f.height = h }

func (f *FormView) Init() tea.Cmd { return textinput.Blink }

func (f *FormView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return f, func() tea.Msg { return DoneFormMsg{Cancelled: true} }
		case "tab", "down":
			f.nextField()
		case "shift+tab", "up":
			f.prevField()
		case "ctrl+p":
			if f.focused == fieldSchedule {
				cur := f.inputs[fieldSchedule].Value()
				next := presets[0].value
				for i, p := range presets {
					if p.value == cur && i+1 < len(presets) {
						next = presets[i+1].value
						break
					}
				}
				f.inputs[fieldSchedule].SetValue(next)
			}
		case "enter":
			if f.focused == fieldComment {
				return f.submit()
			}
			f.nextField()
		case "ctrl+s":
			return f.submit()
		}
	}

	var cmds []tea.Cmd
	for i := range f.inputs {
		var cmd tea.Cmd
		f.inputs[i], cmd = f.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return f, tea.Batch(cmds...)
}

func (f *FormView) nextField() {
	f.inputs[f.focused].Blur()
	f.focused = (f.focused + 1) % len(f.inputs)
	f.inputs[f.focused].Focus()
}

func (f *FormView) prevField() {
	f.inputs[f.focused].Blur()
	f.focused = (f.focused - 1 + len(f.inputs)) % len(f.inputs)
	f.inputs[f.focused].Focus()
}

func (f *FormView) submit() (tea.Model, tea.Cmd) {
	schedule := strings.TrimSpace(f.inputs[fieldSchedule].Value())
	command := strings.TrimSpace(f.inputs[fieldCommand].Value())
	comment := strings.TrimSpace(f.inputs[fieldComment].Value())
	if schedule == "" {
		f.errMsg = "Schedule cannot be empty"
		return f, nil
	}
	if command == "" {
		f.errMsg = "Command cannot be empty"
		return f, nil
	}
	f.errMsg = ""
	return f, func() tea.Msg {
		return DoneFormMsg{
			Cancelled: false,
			Schedule:  schedule,
			Command:   command,
			Comment:   comment,
			EditID:    f.editID,
		}
	}
}

func (f *FormView) View() string {
	var b strings.Builder

	title := "Add Cron Job"
	if f.isEdit {
		title = fmt.Sprintf("Edit Cron Job #%d", f.editID)
	}
	b.WriteString(styles.TitleStyle.Render(" "+title) + "\n\n")

	for i, label := range fieldLabels {
		labelStyle := styles.MutedStyle
		if i == f.focused {
			labelStyle = styles.KeyStyle
		}
		b.WriteString(labelStyle.Render("  "+label) + "\n")

		inputWidth := f.width - 6
		if inputWidth < 20 {
			inputWidth = 20
		}
		var inputStyle = styles.InputStyle.Width(inputWidth)
		if i == f.focused {
			inputStyle = styles.ActiveInputStyle.Width(inputWidth)
		}
		b.WriteString("  " + inputStyle.Render(f.inputs[i].View()) + "\n")
		b.WriteString(styles.MutedStyle.Render("  "+fieldHints[i]) + "\n\n")
	}

	if f.focused == fieldSchedule {
		var presetLabels []string
		for _, p := range presets {
			presetLabels = append(presetLabels, p.label)
		}
		b.WriteString(styles.MutedStyle.Render("  Presets (ctrl+p): "+strings.Join(presetLabels, ", ")) + "\n\n")
	}

	if f.errMsg != "" {
		b.WriteString(styles.ErrorStyle.PaddingLeft(2).Render("Error: "+f.errMsg) + "\n\n")
	}

	hints := []string{
		styles.KeyStyle.Render("tab") + styles.KeyDescStyle.Render(" next field"),
		styles.KeyStyle.Render("shift+tab") + styles.KeyDescStyle.Render(" prev field"),
		styles.KeyStyle.Render("ctrl+s") + styles.KeyDescStyle.Render(" save"),
		styles.KeyStyle.Render("esc") + styles.KeyDescStyle.Render(" cancel"),
	}
	b.WriteString(styles.HelpBarStyle.Width(f.width).Render(strings.Join(hints, "  ")))
	return b.String()
}
