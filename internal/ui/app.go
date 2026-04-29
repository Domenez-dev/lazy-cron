package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/domenez-dev/lazy-cron/internal/cron"
	"github.com/domenez-dev/lazy-cron/internal/views"
)

type Screen int

const (
	ScreenList Screen = iota
	ScreenForm
	ScreenConfirm
)

type App struct {
	manager *cron.Manager
	screen  Screen
	list    *views.ListView
	form    *views.FormView
	confirm *views.ConfirmView
	width   int
	height  int
}

func NewApp() (*App, error) {
	mgr := &cron.Manager{}
	if err := mgr.Load(); err != nil {
		return nil, err
	}
	return &App{
		manager: mgr,
		screen:  ScreenList,
		list:    views.NewListView(mgr),
	}, nil
}

func (a *App) Init() tea.Cmd { return nil }

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		a.list.SetSize(msg.Width, msg.Height)
		if a.form != nil {
			a.form.SetSize(msg.Width, msg.Height)
		}
		if a.confirm != nil {
			a.confirm.SetSize(msg.Width, msg.Height)
		}
		return a, nil

	case views.OpenAddMsg:
		a.form = views.NewFormView(a.manager, 0)
		a.form.SetSize(a.width, a.height)
		a.screen = ScreenForm
		return a, a.form.Init()

	case views.OpenEditMsg:
		a.form = views.NewFormView(a.manager, msg.JobID)
		a.form.SetSize(a.width, a.height)
		a.screen = ScreenForm
		return a, a.form.Init()

	case views.ConfirmDelMsg:
		a.confirm = views.NewConfirmView(a.manager, msg.JobID)
		a.confirm.SetSize(a.width, a.height)
		a.screen = ScreenConfirm
		return a, nil

	case views.DoneFormMsg:
		a.screen = ScreenList
		if msg.Cancelled {
			a.list.SetMessage("Cancelled", true)
			return a, nil
		}
		var err error
		if msg.EditID == 0 {
			err = a.manager.AddUserJob(msg.Schedule, msg.Command, msg.Comment)
			if err != nil {
				a.list.SetMessage(fmt.Sprintf("Error adding job: %s", err), false)
			} else {
				a.list.SetMessage("Job added successfully", true)
			}
		} else {
			for _, j := range a.manager.Jobs {
				if j.ID == msg.EditID {
					j.Schedule = msg.Schedule
					j.Command = msg.Command
					j.Comment = msg.Comment
					break
				}
			}
			err = a.manager.WriteUserCrontab()
			if err != nil {
				a.list.SetMessage(fmt.Sprintf("Error saving job: %s", err), false)
			} else {
				a.list.SetMessage("Job updated successfully", true)
			}
		}
		return a, nil

	case views.ConfirmResultMsg:
		a.screen = ScreenList
		if !msg.Confirmed {
			a.list.SetMessage("Cancelled", true)
			return a, nil
		}
		if err := a.manager.DeleteUserJob(msg.JobID); err != nil {
			a.list.SetMessage(fmt.Sprintf("Error deleting job: %s", err), false)
		} else {
			a.list.SetMessage("Job deleted", true)
		}
		return a, nil
	}

	switch a.screen {
	case ScreenList:
		m, cmd := a.list.Update(msg)
		a.list = m.(*views.ListView)
		return a, cmd
	case ScreenForm:
		if a.form != nil {
			m, cmd := a.form.Update(msg)
			a.form = m.(*views.FormView)
			return a, cmd
		}
	case ScreenConfirm:
		if a.confirm != nil {
			m, cmd := a.confirm.Update(msg)
			a.confirm = m.(*views.ConfirmView)
			return a, cmd
		}
	}
	return a, nil
}

func (a *App) View() string {
	switch a.screen {
	case ScreenForm:
		if a.form != nil {
			return a.form.View()
		}
	case ScreenConfirm:
		if a.confirm != nil {
			return a.confirm.View()
		}
	}
	return a.list.View()
}
