package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/domenez-dev/lazy-chrony/internal/cron"
	"github.com/domenez-dev/lazy-chrony/internal/styles"
)

// ─────────────────────────────────────────────────────────────────────────────
// Field option types
// ─────────────────────────────────────────────────────────────────────────────

type fieldOpt struct {
	label string // display label shown in the cell
	value string // cron expression fragment
}

// fieldOptGroup marks the start of a named "mode group" within a field's flat option list.
type fieldOptGroup struct {
	name  string // e.g. "all", "every", "at", "range"
	start int    // index into the flat allOpts[field] slice
}

const (
	fldMinute  = 0
	fldHour    = 1
	fldDOM     = 2
	fldMonth   = 3
	fldWeekday = 4
)

var fldTitles = [5]string{"MINUTE", "HOUR", "DAY", "MONTH", "WEEKDAY"}

// Per-field accent colours (MINUTE=purple, HOUR=blue, DAY=green, MONTH=yellow, WEEKDAY=cyan)
var fldColors = [5]lipgloss.TerminalColor{
	styles.ColorAccent,
	styles.ColorBlue,
	styles.ColorGreen,
	styles.ColorYellow,
	styles.ColorCyan,
}

var allOpts [5][]fieldOpt
var allGroups [5][]fieldOptGroup

func init() {
	allOpts[fldMinute], allGroups[fldMinute] = buildMinuteOpts()
	allOpts[fldHour], allGroups[fldHour] = buildHourOpts()
	allOpts[fldDOM], allGroups[fldDOM] = buildDOMOpts()
	allOpts[fldMonth], allGroups[fldMonth] = buildMonthOpts()
	allOpts[fldWeekday], allGroups[fldWeekday] = buildWeekdayOpts()
}

// ── Minute ────────────────────────────────────────────────────────────────────

func buildMinuteOpts() ([]fieldOpt, []fieldOptGroup) {
	var opts []fieldOpt
	var groups []fieldOptGroup

	groups = append(groups, fieldOptGroup{"all", len(opts)})
	opts = append(opts, fieldOpt{"every min", "*"})

	groups = append(groups, fieldOptGroup{"every", len(opts)})
	for _, n := range []int{2, 3, 5, 10, 15, 20, 30} {
		opts = append(opts, fieldOpt{fmt.Sprintf("every %dmin", n), fmt.Sprintf("*/%d", n)})
	}

	groups = append(groups, fieldOptGroup{"at", len(opts)})
	for i := 0; i <= 59; i++ {
		opts = append(opts, fieldOpt{fmt.Sprintf("at :%02d", i), fmt.Sprintf("%d", i)})
	}

	groups = append(groups, fieldOptGroup{"range", len(opts)})
	for _, r := range [][2]int{{0, 4}, {0, 9}, {0, 14}, {0, 29}, {30, 59}, {15, 44}} {
		opts = append(opts, fieldOpt{
			fmt.Sprintf(":%02d–:%02d", r[0], r[1]),
			fmt.Sprintf("%d-%d", r[0], r[1]),
		})
	}
	return opts, groups
}

// ── Hour ──────────────────────────────────────────────────────────────────────

func buildHourOpts() ([]fieldOpt, []fieldOptGroup) {
	var opts []fieldOpt
	var groups []fieldOptGroup

	groups = append(groups, fieldOptGroup{"all", len(opts)})
	opts = append(opts, fieldOpt{"every hour", "*"})

	groups = append(groups, fieldOptGroup{"every", len(opts)})
	for _, n := range []int{2, 3, 4, 6, 8, 12} {
		opts = append(opts, fieldOpt{fmt.Sprintf("every %dh", n), fmt.Sprintf("*/%d", n)})
	}

	groups = append(groups, fieldOptGroup{"at", len(opts)})
	for i := 0; i <= 23; i++ {
		opts = append(opts, fieldOpt{fmt.Sprintf("%02d:xx", i), fmt.Sprintf("%d", i)})
	}

	groups = append(groups, fieldOptGroup{"range", len(opts)})
	for _, r := range [][2]int{{0, 5}, {6, 11}, {9, 17}, {8, 18}, {12, 23}, {0, 11}} {
		opts = append(opts, fieldOpt{
			fmt.Sprintf("%02d–%02dh", r[0], r[1]),
			fmt.Sprintf("%d-%d", r[0], r[1]),
		})
	}
	return opts, groups
}

// ── DOM ───────────────────────────────────────────────────────────────────────

var domOrdinals = [32]string{
	"", "1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th",
	"10th", "11th", "12th", "13th", "14th", "15th", "16th", "17th", "18th", "19th",
	"20th", "21st", "22nd", "23rd", "24th", "25th", "26th", "27th", "28th", "29th",
	"30th", "31st",
}

func buildDOMOpts() ([]fieldOpt, []fieldOptGroup) {
	var opts []fieldOpt
	var groups []fieldOptGroup

	groups = append(groups, fieldOptGroup{"all", len(opts)})
	opts = append(opts, fieldOpt{"every day", "*"})

	groups = append(groups, fieldOptGroup{"every", len(opts)})
	for _, n := range []int{2, 5, 7, 10, 14, 15} {
		opts = append(opts, fieldOpt{fmt.Sprintf("every %dd", n), fmt.Sprintf("*/%d", n)})
	}

	groups = append(groups, fieldOptGroup{"at", len(opts)})
	for i := 1; i <= 31; i++ {
		opts = append(opts, fieldOpt{domOrdinals[i], fmt.Sprintf("%d", i)})
	}

	groups = append(groups, fieldOptGroup{"range", len(opts)})
	for _, r := range [][2]int{{1, 7}, {1, 14}, {1, 15}, {15, 31}} {
		opts = append(opts, fieldOpt{
			fmt.Sprintf("%d–%dth", r[0], r[1]),
			fmt.Sprintf("%d-%d", r[0], r[1]),
		})
	}
	return opts, groups
}

// ── Month ─────────────────────────────────────────────────────────────────────

var monthLabels = [13]string{
	"", "January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

func buildMonthOpts() ([]fieldOpt, []fieldOptGroup) {
	var opts []fieldOpt
	var groups []fieldOptGroup

	groups = append(groups, fieldOptGroup{"all", len(opts)})
	opts = append(opts, fieldOpt{"every month", "*"})

	groups = append(groups, fieldOptGroup{"every", len(opts)})
	for _, n := range []int{2, 3, 4, 6} {
		opts = append(opts, fieldOpt{fmt.Sprintf("every %dmo", n), fmt.Sprintf("*/%d", n)})
	}

	groups = append(groups, fieldOptGroup{"at", len(opts)})
	for i := 1; i <= 12; i++ {
		opts = append(opts, fieldOpt{monthLabels[i], fmt.Sprintf("%d", i)})
	}

	groups = append(groups, fieldOptGroup{"range", len(opts)})
	for _, r := range [][2]int{{1, 3}, {4, 6}, {7, 9}, {10, 12}, {1, 6}} {
		opts = append(opts, fieldOpt{
			monthLabels[r[0]][:3] + "–" + monthLabels[r[1]][:3],
			fmt.Sprintf("%d-%d", r[0], r[1]),
		})
	}
	return opts, groups
}

// ── Weekday ───────────────────────────────────────────────────────────────────

var weekdayFull = [7]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
var weekdayShort = [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

func buildWeekdayOpts() ([]fieldOpt, []fieldOptGroup) {
	var opts []fieldOpt
	var groups []fieldOptGroup

	groups = append(groups, fieldOptGroup{"any", len(opts)})
	opts = append(opts, fieldOpt{"any day", "*"})

	groups = append(groups, fieldOptGroup{"on", len(opts)})
	for _, i := range []int{1, 2, 3, 4, 5, 6, 0} {
		opts = append(opts, fieldOpt{weekdayFull[i] + "s", fmt.Sprintf("%d", i)})
	}

	groups = append(groups, fieldOptGroup{"range", len(opts)})
	dowRangePresets := [][2]int{{1, 5}, {0, 5}, {1, 6}, {0, 6}}
	dowRangeLabels := []string{"Mon–Fri", "Sun–Fri", "Mon–Sat", "Sun–Sat"}
	for i, r := range dowRangePresets {
		opts = append(opts, fieldOpt{
			dowRangeLabels[i],
			fmt.Sprintf("%d-%d", r[0], r[1]),
		})
	}
	return opts, groups
}

// ─────────────────────────────────────────────────────────────────────────────
// Raw-mode presets
// ─────────────────────────────────────────────────────────────────────────────

var rawPresets = []struct{ label, value string }{
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
	{"weekdays 9am", "0 9 * * 1-5"},
}

// ─────────────────────────────────────────────────────────────────────────────
// Form sections
// ─────────────────────────────────────────────────────────────────────────────

type FormSection int

const (
	SecBuilder FormSection = iota
	SecCommand
	SecComment
)

// ─────────────────────────────────────────────────────────────────────────────
// FormView
// ─────────────────────────────────────────────────────────────────────────────

type FormView struct {
	manager *cron.Manager
	editID  int
	isEdit  bool

	// Schedule builder state
	builderSel   [5]int // selected option index per field (into allOpts[field])
	builderFocus int    // 0-4: which field is focused
	rawMode      bool   // true = show raw text input for schedule

	// Text inputs
	schedInput textinput.Model // used when rawMode == true
	cmdInput   textinput.Model
	cmpInput   textinput.Model

	section FormSection
	width   int
	height  int
	errMsg  string
}

// DoneFormMsg is sent when the form is submitted or cancelled.
type DoneFormMsg struct {
	Cancelled bool
	Schedule  string
	Command   string
	Comment   string
	EditID    int
}

// ─────────────────────────────────────────────────────────────────────────────
// Constructor
// ─────────────────────────────────────────────────────────────────────────────

func NewFormView(mgr *cron.Manager, editJobID int) *FormView {
	f := &FormView{
		manager: mgr,
		editID:  editJobID,
		isEdit:  editJobID != 0,
		section: SecBuilder,
	}

	f.schedInput = textinput.New()
	f.schedInput.Prompt = ""
	f.schedInput.CharLimit = 128
	f.schedInput.Placeholder = "*/5 * * * *"

	f.cmdInput = textinput.New()
	f.cmdInput.Prompt = ""
	f.cmdInput.CharLimit = 256
	f.cmdInput.Placeholder = "/path/to/script.sh"

	f.cmpInput = textinput.New()
	f.cmpInput.Prompt = ""
	f.cmpInput.CharLimit = 128
	f.cmpInput.Placeholder = "optional description"

	if editJobID != 0 {
		for _, j := range mgr.Jobs {
			if j.ID == editJobID {
				f.cmdInput.SetValue(j.Command)
				f.cmpInput.SetValue(j.Comment)
				if !f.syncBuilderFromExpr(j.Schedule) {
					f.schedInput.SetValue(j.Schedule)
					f.rawMode = true
					f.schedInput.Focus()
				}
				break
			}
		}
	}
	return f
}

// ─────────────────────────────────────────────────────────────────────────────
// Builder helpers
// ─────────────────────────────────────────────────────────────────────────────

// syncBuilderFromExpr parses a 5-field cron expression into builder state.
func (f *FormView) syncBuilderFromExpr(expr string) bool {
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return false
	}
	var newSel [5]int
	for i := 0; i < 5; i++ {
		found := false
		for j, opt := range allOpts[i] {
			if opt.value == parts[i] {
				newSel[i] = j
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	f.builderSel = newSel
	return true
}

// builderExpr returns the current 5-field cron expression from builder state.
func (f *FormView) builderExpr() string {
	parts := make([]string, 5)
	for i := 0; i < 5; i++ {
		parts[i] = allOpts[i][f.builderSel[i]].value
	}
	return strings.Join(parts, " ")
}

// currentSchedule returns the schedule from builder or raw input.
func (f *FormView) currentSchedule() string {
	if f.rawMode {
		return strings.TrimSpace(f.schedInput.Value())
	}
	return f.builderExpr()
}

// currentGroupIdx returns the index of the mode group the focused field is in.
func (f *FormView) currentGroupIdx(fld int) int {
	idx := f.builderSel[fld]
	groups := allGroups[fld]
	cur := 0
	for gi := len(groups) - 1; gi >= 0; gi-- {
		if idx >= groups[gi].start {
			cur = gi
			break
		}
	}
	return cur
}

// currentGroupName returns the name of the mode group for field fld.
func (f *FormView) currentGroupName(fld int) string {
	return allGroups[fld][f.currentGroupIdx(fld)].name
}

// cycleMode jumps the focused field's selection to the first option of the next group.
func (f *FormView) cycleMode() {
	fld := f.builderFocus
	groups := allGroups[fld]
	cur := f.currentGroupIdx(fld)
	next := (cur + 1) % len(groups)
	f.builderSel[fld] = groups[next].start
}

// ─────────────────────────────────────────────────────────────────────────────
// Tea interface
// ─────────────────────────────────────────────────────────────────────────────

func (f *FormView) SetSize(w, h int) {
	f.width = w
	f.height = h
	// Set textinput widths to match the lipgloss container.
	// Container: Width(inputW) with Padding(0,1) → text area = inputW - 2
	inputW := w - 6 // 2 margin + 2 border + 2 padding
	if inputW < 20 {
		inputW = 20
	}
	textW := inputW - 2
	if textW < 10 {
		textW = 10
	}
	f.schedInput.Width = textW
	f.cmdInput.Width = textW
	f.cmpInput.Width = textW
}

func (f *FormView) Init() tea.Cmd { return textinput.Blink }

func (f *FormView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return f.handleKey(msg)
	}
	// Forward non-key events (cursor blink, etc.) to the focused input.
	var cmds []tea.Cmd
	if f.rawMode && f.section == SecBuilder {
		var cmd tea.Cmd
		f.schedInput, cmd = f.schedInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	if f.section == SecCommand {
		var cmd tea.Cmd
		f.cmdInput, cmd = f.cmdInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	if f.section == SecComment {
		var cmd tea.Cmd
		f.cmpInput, cmd = f.cmpInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	return f, tea.Batch(cmds...)
}

func (f *FormView) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "esc":
		return f, func() tea.Msg { return DoneFormMsg{Cancelled: true} }
	case "ctrl+s":
		return f.submit()
	}
	switch f.section {
	case SecBuilder:
		return f.handleBuilderKey(key, msg)
	case SecCommand:
		return f.handleCmdKey(key, msg)
	case SecComment:
		return f.handleCmpKey(key, msg)
	}
	return f, nil
}

func (f *FormView) handleBuilderKey(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if f.rawMode {
		switch key {
		case "ctrl+e":
			f.syncBuilderFromExpr(strings.TrimSpace(f.schedInput.Value()))
			f.rawMode = false
			f.schedInput.Blur()
			return f, nil
		case "ctrl+p":
			// Cycle through raw presets
			cur := strings.TrimSpace(f.schedInput.Value())
			next := rawPresets[0].value
			for i, p := range rawPresets {
				if p.value == cur && i+1 < len(rawPresets) {
					next = rawPresets[i+1].value
					break
				}
			}
			f.schedInput.SetValue(next)
			return f, nil
		case "tab", "enter":
			f.schedInput.Blur()
			f.section = SecCommand
			f.cmdInput.Focus()
			return f, textinput.Blink
		case "shift+tab":
			f.schedInput.Blur()
			f.section = SecComment
			f.cmpInput.Focus()
			return f, textinput.Blink
		}
		var cmd tea.Cmd
		f.schedInput, cmd = f.schedInput.Update(msg)
		return f, cmd
	}

	// Builder navigation mode
	switch key {
	case "ctrl+e":
		f.schedInput.SetValue(f.builderExpr())
		f.rawMode = true
		f.schedInput.Focus()
		return f, textinput.Blink
	case "tab", "enter":
		f.section = SecCommand
		f.cmdInput.Focus()
		return f, textinput.Blink
	case "shift+tab":
		f.section = SecComment
		f.cmpInput.Focus()
		return f, textinput.Blink
	case "h", "left":
		if f.builderFocus > 0 {
			f.builderFocus--
		}
	case "l", "right":
		if f.builderFocus < 4 {
			f.builderFocus++
		}
	case "j", "down":
		max := len(allOpts[f.builderFocus]) - 1
		if f.builderSel[f.builderFocus] < max {
			f.builderSel[f.builderFocus]++
		}
	case "k", "up":
		if f.builderSel[f.builderFocus] > 0 {
			f.builderSel[f.builderFocus]--
		}
	case "m":
		// Cycle to the first option of the next mode group
		f.cycleMode()
	}
	return f, nil
}

func (f *FormView) handleCmdKey(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key {
	case "tab", "enter":
		f.cmdInput.Blur()
		f.section = SecComment
		f.cmpInput.Focus()
		return f, textinput.Blink
	case "shift+tab":
		f.cmdInput.Blur()
		f.section = SecBuilder
		if f.rawMode {
			f.schedInput.Focus()
			return f, textinput.Blink
		}
		return f, nil
	}
	var cmd tea.Cmd
	f.cmdInput, cmd = f.cmdInput.Update(msg)
	return f, cmd
}

func (f *FormView) handleCmpKey(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key {
	case "tab":
		f.cmpInput.Blur()
		f.section = SecBuilder
		if f.rawMode {
			f.schedInput.Focus()
			return f, textinput.Blink
		}
		return f, nil
	case "shift+tab":
		f.cmpInput.Blur()
		f.section = SecCommand
		f.cmdInput.Focus()
		return f, textinput.Blink
	case "enter":
		return f.submit()
	}
	var cmd tea.Cmd
	f.cmpInput, cmd = f.cmpInput.Update(msg)
	return f, cmd
}

func (f *FormView) submit() (tea.Model, tea.Cmd) {
	schedule := f.currentSchedule()
	command := strings.TrimSpace(f.cmdInput.Value())
	comment := strings.TrimSpace(f.cmpInput.Value())
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

// ─────────────────────────────────────────────────────────────────────────────
// View
// ─────────────────────────────────────────────────────────────────────────────

func (f *FormView) View() string {
	hints := f.renderHints()
	body := f.renderBody()

	// Pin the hints bar to the very bottom of the terminal.
	if f.height > 0 {
		bodyH := lipgloss.Height(body)
		hintsH := lipgloss.Height(hints)
		gap := f.height - bodyH - hintsH
		if gap > 0 {
			body += strings.Repeat("\n", gap)
		}
	}
	return body + hints
}

func (f *FormView) renderBody() string {
	var b strings.Builder

	title := " Add Cron Job"
	if f.isEdit {
		title = fmt.Sprintf(" Edit Cron Job #%d", f.editID)
	}
	b.WriteString(styles.TitleStyle.Render(title) + "\n\n")

	// ── Schedule ────────────────────────────────────────────────────────────
	schedActive := f.section == SecBuilder
	b.WriteString(f.renderSectionTitle("SCHEDULE", schedActive))
	if f.rawMode {
		b.WriteString(f.renderRawSchedule(schedActive))
	} else {
		b.WriteString(f.renderBuilder(schedActive))
	}

	// Live description
	expr := f.currentSchedule()
	desc := cron.DescribeSchedule(expr)
	if expr != "" {
		b.WriteString(styles.MutedStyle.Render("  cron: ") + styles.CronExprStyle.Render(expr) + "\n")
	}
	if desc != "" {
		b.WriteString(styles.DescriptionStyle.PaddingLeft(2).Render("✦ "+desc) + "\n")
	}
	b.WriteString("\n")

	// ── Command ─────────────────────────────────────────────────────────────
	inputW := f.width - 6
	if inputW < 20 {
		inputW = 20
	}
	cmdActive := f.section == SecCommand
	b.WriteString(f.renderSectionTitle("COMMAND", cmdActive))
	cmdStyle := styles.InputStyle.Width(inputW)
	if cmdActive {
		cmdStyle = styles.ActiveInputStyle.Width(inputW)
	}
	b.WriteString("  " + cmdStyle.Render(f.cmdInput.View()) + "\n")
	b.WriteString(styles.MutedStyle.Render("  e.g. /usr/bin/backup.sh >> /var/log/backup.log 2>&1") + "\n\n")

	// ── Comment ─────────────────────────────────────────────────────────────
	cmpActive := f.section == SecComment
	b.WriteString(f.renderSectionTitle("COMMENT  (optional)", cmpActive))
	cmpStyle := styles.InputStyle.Width(inputW)
	if cmpActive {
		cmpStyle = styles.ActiveInputStyle.Width(inputW)
	}
	b.WriteString("  " + cmpStyle.Render(f.cmpInput.View()) + "\n")
	b.WriteString(styles.MutedStyle.Render("  e.g. nightly backup job") + "\n\n")

	// ── Error ────────────────────────────────────────────────────────────────
	if f.errMsg != "" {
		b.WriteString(styles.ErrorStyle.PaddingLeft(2).Render("✖  "+f.errMsg) + "\n\n")
	}

	return b.String()
}

// ─────────────────────────────────────────────────────────────────────────────
// Render helpers
// ─────────────────────────────────────────────────────────────────────────────

func (f *FormView) renderSectionTitle(title string, active bool) string {
	var fg lipgloss.TerminalColor
	if active {
		fg = styles.ColorAccent
	} else {
		fg = styles.ColorMuted
	}
	titleStr := lipgloss.NewStyle().Foreground(fg).Bold(active).Render(title)
	lineLen := f.width - len([]rune(title)) - 5
	if lineLen < 1 {
		lineLen = 1
	}
	sep := styles.MutedStyle.Render(strings.Repeat("─", lineLen))
	return "  " + titleStr + " " + sep + "\n"
}

func (f *FormView) renderBuilder(active bool) string {
	var b strings.Builder

	// Compute cell content width dynamically
	available := f.width - 2
	cellW := available/5 - 2 // subtract 2 for RoundedBorder (left+right wall)
	if cellW < 11 {
		cellW = 11
	}
	if cellW > 18 {
		cellW = 18
	}
	totalCellW := cellW + 2

	// ── Header row (field titles + mode tag for focused field) ────────────
	var headers []string
	for i, title := range fldTitles {
		label := title
		if active && i == f.builderFocus {
			label = title + " [" + f.currentGroupName(i) + "]"
		}
		var hs lipgloss.Style
		if active {
			fg := fldColors[i]
			if i == f.builderFocus {
				hs = lipgloss.NewStyle().Foreground(fg).Bold(true).Width(totalCellW).Align(lipgloss.Center)
			} else {
				hs = lipgloss.NewStyle().Foreground(fg).Width(totalCellW).Align(lipgloss.Center)
			}
		} else {
			hs = styles.MutedStyle.Copy().Width(totalCellW).Align(lipgloss.Center)
		}
		headers = append(headers, hs.Render(label))
	}
	b.WriteString("  " + strings.Join(headers, "") + "\n")

	// ── Cell row ──────────────────────────────────────────────────────────
	cells := make([]string, 5)
	for i := 0; i < 5; i++ {
		label := allOpts[i][f.builderSel[i]].label
		var cs lipgloss.Style
		if active {
			fc := fldColors[i]
			if i == f.builderFocus {
				cs = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(fc).
					Foreground(fc).
					Bold(true).
					Width(cellW).
					Align(lipgloss.Center)
			} else {
				cs = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(fc).
					Foreground(fc).
					Width(cellW).
					Align(lipgloss.Center)
			}
		} else {
			cs = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(styles.ColorBorder).
				Foreground(styles.ColorMuted).
				Width(cellW).
				Align(lipgloss.Center)
		}
		cells[i] = cs.Render(label)
	}
	// JoinHorizontal returns a multi-line string; indent every line.
	cellRow := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	for _, line := range strings.Split(cellRow, "\n") {
		b.WriteString("  " + line + "\n")
	}

	if active {
		b.WriteString(styles.MutedStyle.Render("  j/k scroll  ·  h/l field  ·  m mode  ·  ctrl+e manual") + "\n")
	}
	b.WriteString("\n")
	return b.String()
}

func (f *FormView) renderRawSchedule(active bool) string {
	var b strings.Builder
	inputW := f.width - 6
	if inputW < 20 {
		inputW = 20
	}
	var inputStyle lipgloss.Style
	if active {
		inputStyle = styles.ActiveInputStyle.Width(inputW)
	} else {
		inputStyle = styles.InputStyle.Width(inputW)
	}
	b.WriteString("  " + inputStyle.Render(f.schedInput.View()) + "\n")
	b.WriteString(styles.MutedStyle.Render("  e.g.  */5 * * * *   ·   @daily   ·   @reboot") + "\n")
	if active {
		// Show preset labels
		var pLabels []string
		for _, p := range rawPresets {
			pLabels = append(pLabels, styles.KeyStyle.Render(p.label))
		}
		b.WriteString(styles.MutedStyle.Render("  ctrl+p presets: ") + strings.Join(pLabels, styles.MutedStyle.Render("  ")) + "\n")
		b.WriteString(styles.MutedStyle.Render("  ctrl+e  switch back to builder") + "\n")
	}
	b.WriteString("\n")
	return b.String()
}

func (f *FormView) renderHints() string {
	var parts []string
	switch f.section {
	case SecBuilder:
		if !f.rawMode {
			parts = []string{
				styles.KeyStyle.Render("j/k") + styles.KeyDescStyle.Render(" scroll"),
				styles.KeyStyle.Render("h/l") + styles.KeyDescStyle.Render(" field"),
				styles.KeyStyle.Render("m") + styles.KeyDescStyle.Render(" mode"),
				styles.KeyStyle.Render("ctrl+e") + styles.KeyDescStyle.Render(" manual"),
				styles.KeyStyle.Render("tab") + styles.KeyDescStyle.Render(" next"),
				styles.KeyStyle.Render("ctrl+s") + styles.KeyDescStyle.Render(" save"),
				styles.KeyStyle.Render("esc") + styles.KeyDescStyle.Render(" cancel"),
			}
		} else {
			parts = []string{
				styles.KeyStyle.Render("ctrl+p") + styles.KeyDescStyle.Render(" preset"),
				styles.KeyStyle.Render("ctrl+e") + styles.KeyDescStyle.Render(" builder"),
				styles.KeyStyle.Render("tab") + styles.KeyDescStyle.Render(" next"),
				styles.KeyStyle.Render("ctrl+s") + styles.KeyDescStyle.Render(" save"),
				styles.KeyStyle.Render("esc") + styles.KeyDescStyle.Render(" cancel"),
			}
		}
	default:
		parts = []string{
			styles.KeyStyle.Render("tab") + styles.KeyDescStyle.Render(" next"),
			styles.KeyStyle.Render("shift+tab") + styles.KeyDescStyle.Render(" prev"),
			styles.KeyStyle.Render("ctrl+s") + styles.KeyDescStyle.Render(" save"),
			styles.KeyStyle.Render("esc") + styles.KeyDescStyle.Render(" cancel"),
		}
	}
	return styles.HelpBarStyle.Width(f.width).Render(strings.Join(parts, "  "))
}
