package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/domenez-dev/lazy-cron/internal/cron"
	"github.com/domenez-dev/lazy-cron/internal/styles"
)

type ConfirmView struct {
	jobID   int
	manager *cron.Manager
	width   int
	height  int
}

type ConfirmResultMsg struct {
	Confirmed bool
	JobID     int
}

func NewConfirmView(mgr *cron.Manager, jobID int) *ConfirmView {
	return &ConfirmView{manager: mgr, jobID: jobID}
}

func (c *ConfirmView) SetSize(w, h int) { c.width = w; c.height = h }

func (c *ConfirmView) Init() tea.Cmd { return nil }

func (c *ConfirmView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y", "enter":
			return c, func() tea.Msg { return ConfirmResultMsg{Confirmed: true, JobID: c.jobID} }
		case "n", "N", "esc", "q":
			return c, func() tea.Msg { return ConfirmResultMsg{Confirmed: false, JobID: c.jobID} }
		}
	}
	return c, nil
}

func (c *ConfirmView) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render(" Delete Cron Job") + "\n\n")

	for _, j := range c.manager.Jobs {
		if j.ID == c.jobID {
			b.WriteString(styles.MutedStyle.Render("  Job #"+fmt.Sprintf("%d", j.ID)) + "\n")
			b.WriteString(styles.NormalRowStyle.Render("  Schedule: ") + styles.KeyStyle.Render(j.Schedule) + "\n")
			b.WriteString(styles.NormalRowStyle.Render("  Command:  ") + styles.KeyStyle.Render(j.Command) + "\n\n")
			break
		}
	}

	b.WriteString(styles.ErrorStyle.Render("  Are you sure you want to delete this job?") + "\n\n")

	hints := []string{
		styles.KeyStyle.Render("y") + styles.KeyDescStyle.Render(" yes, delete"),
		styles.KeyStyle.Render("n/esc") + styles.KeyDescStyle.Render(" cancel"),
	}
	b.WriteString(styles.HelpBarStyle.Width(c.width).Render(strings.Join(hints, "  ")))
	return b.String()
}
