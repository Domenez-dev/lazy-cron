package cron

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	robfigcron "github.com/robfig/cron/v3"
)

// Source describes where a cron job came from.
type Source int

const (
	SourceUser   Source = iota // user crontab (crontab -l)
	SourceSystem               // /etc/cron.d/* or /etc/crontab
	SourceCrond                // /etc/cron.d/ directory entries
)

func (s Source) String() string {
	switch s {
	case SourceUser:
		return "user"
	case SourceSystem:
		return "system"
	case SourceCrond:
		return "cron.d"
	default:
		return "unknown"
	}
}

// Job represents a single cron job entry.
type Job struct {
	ID       int
	Schedule string // raw schedule string e.g. "*/5 * * * *"
	Command  string
	Comment  string // inline comment or description
	User     string // relevant for /etc/crontab and /etc/cron.d
	Source   Source
	FilePath string // file that contains this job
	LineNum  int    // 1-based line number in that file
	Disabled bool   // line starts with #
	Raw      string // original raw line
}

// NextRun returns the next scheduled time for this job.
// Returns zero time if the schedule is unparseable.
func (j *Job) NextRun() time.Time {
	if j.Disabled {
		return time.Time{}
	}
	parser := robfigcron.NewParser(
		robfigcron.Minute | robfigcron.Hour | robfigcron.Dom |
			robfigcron.Month | robfigcron.Dow | robfigcron.Descriptor,
	)
	sched, err := parser.Parse(j.Schedule)
	if err != nil {
		return time.Time{}
	}
	return sched.Next(time.Now())
}

// NextRunStr returns a human-friendly string for the next run time.
func (j *Job) NextRunStr() string {
	if j.Disabled {
		return "disabled"
	}
	t := j.NextRun()
	if t.IsZero() {
		return "invalid schedule"
	}
	d := time.Until(t).Round(time.Second)
	if d < 0 {
		return "overdue"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	return fmt.Sprintf("%dd %dh", int(d.Hours())/24, int(d.Hours())%24)
}

// Manager loads, saves, and mutates cron jobs.
type Manager struct {
	Jobs    []*Job
	counter int
}

// Load reads all cron jobs from user crontab and system files.
func (m *Manager) Load() error {
	m.Jobs = nil
	m.counter = 0

	// 1. User crontab
	if err := m.loadUserCrontab(); err != nil {
		// Non-fatal: user might not have a crontab yet
		_ = err
	}

	// 2. /etc/crontab
	_ = m.loadFile("/etc/crontab", SourceSystem, true)

	// 3. /etc/cron.d/*
	entries, _ := filepath.Glob("/etc/cron.d/*")
	for _, f := range entries {
		_ = m.loadFile(f, SourceCrond, true)
	}

	return nil
}

func (m *Manager) nextID() int {
	m.counter++
	return m.counter
}

func (m *Manager) loadUserCrontab() error {
	out, err := exec.Command("crontab", "-l").Output()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		job := parseUserLine(line, lineNum)
		if job == nil {
			continue
		}
		job.ID = m.nextID()
		job.Source = SourceUser
		job.FilePath = "crontab"
		m.Jobs = append(m.Jobs, job)
	}
	return nil
}

func (m *Manager) loadFile(path string, src Source, hasUser bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		var job *Job
		if hasUser {
			job = parseSystemLine(line, lineNum)
		} else {
			job = parseUserLine(line, lineNum)
		}
		if job == nil {
			continue
		}
		job.ID = m.nextID()
		job.Source = src
		job.FilePath = path
		m.Jobs = append(m.Jobs, job)
	}
	return nil
}

// parseUserLine parses a line from a user crontab (no user field).
func parseUserLine(line string, lineNum int) *Job {
	raw := line
	disabled := false

	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "##") {
		return nil
	}

	// Disabled job: starts with a single #
	if strings.HasPrefix(trimmed, "#") {
		disabled = true
		trimmed = strings.TrimSpace(trimmed[1:])
	}

	// Skip env vars and special keywords
	if strings.Contains(trimmed, "=") && !strings.Contains(trimmed, " ") {
		return nil
	}
	if strings.HasPrefix(trimmed, "@") {
		// @reboot @daily etc - single field schedules
		parts := strings.Fields(trimmed)
		if len(parts) < 2 {
			return nil
		}
		return &Job{
			Schedule: parts[0],
			Command:  strings.Join(parts[1:], " "),
			Disabled: disabled,
			Raw:      raw,
			LineNum:  lineNum,
		}
	}

	parts := strings.Fields(trimmed)
	if len(parts) < 6 {
		return nil
	}
	schedule := strings.Join(parts[:5], " ")
	command := strings.Join(parts[5:], " ")

	// Strip inline comment from command
	comment := ""
	if idx := strings.Index(command, " #"); idx != -1 {
		comment = strings.TrimSpace(command[idx+2:])
		command = strings.TrimSpace(command[:idx])
	}

	return &Job{
		Schedule: schedule,
		Command:  command,
		Comment:  comment,
		Disabled: disabled,
		Raw:      raw,
		LineNum:  lineNum,
	}
}

// parseSystemLine parses a line from /etc/crontab or /etc/cron.d (has user field).
func parseSystemLine(line string, lineNum int) *Job {
	job := parseUserLine(line, lineNum)
	if job == nil {
		return nil
	}
	// System crontabs have: schedule user command
	// After the 5 schedule fields, next is user, then command
	parts := strings.Fields(job.Command)
	if len(parts) >= 2 {
		job.User = parts[0]
		job.Command = strings.Join(parts[1:], " ")
	}
	return job
}

// WriteUserCrontab rewrites the user's crontab with current in-memory jobs.
func (m *Manager) WriteUserCrontab() error {
	var lines []string

	// Collect jobs that belong to user crontab, preserve their raw non-job lines
	// by re-running crontab -l and patching changed lines
	out, _ := exec.Command("crontab", "-l").Output()
	rawLines := strings.Split(string(out), "\n")

	// Build a map of lineNum -> job for user jobs
	userJobs := map[int]*Job{}
	for _, j := range m.Jobs {
		if j.Source == SourceUser {
			userJobs[j.LineNum] = j
		}
	}

	for i, raw := range rawLines {
		lineNum := i + 1
		if job, ok := userJobs[lineNum]; ok {
			lines = append(lines, jobToLine(job))
		} else {
			lines = append(lines, raw)
		}
	}

	content := strings.Join(lines, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(content)
	return cmd.Run()
}

// AddUserJob appends a new job to the user's crontab.
func (m *Manager) AddUserJob(schedule, command, comment string) error {
	job := &Job{
		ID:       m.nextID(),
		Schedule: schedule,
		Command:  command,
		Comment:  comment,
		Source:   SourceUser,
		FilePath: "crontab",
		Disabled: false,
	}

	// Read current crontab
	out, _ := exec.Command("crontab", "-l").Output()
	existing := strings.TrimRight(string(out), "\n")

	newLine := jobToLine(job)
	var content string
	if existing == "" {
		content = newLine + "\n"
	} else {
		content = existing + "\n" + newLine + "\n"
	}

	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(content)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Re-assign line number
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	job.LineNum = len(lines)
	job.Raw = newLine
	m.Jobs = append(m.Jobs, job)
	return nil
}

// DeleteUserJob removes a user crontab job by ID.
func (m *Manager) DeleteUserJob(id int) error {
	var target *Job
	for _, j := range m.Jobs {
		if j.ID == id {
			target = j
			break
		}
	}
	if target == nil || target.Source != SourceUser {
		return fmt.Errorf("job %d not found or not a user job", id)
	}

	out, _ := exec.Command("crontab", "-l").Output()
	rawLines := strings.Split(string(out), "\n")

	var newLines []string
	for i, line := range rawLines {
		if i+1 != target.LineNum {
			newLines = append(newLines, line)
		}
	}

	content := strings.Join(newLines, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(content)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Remove from in-memory list
	newJobs := make([]*Job, 0, len(m.Jobs)-1)
	for _, j := range m.Jobs {
		if j.ID != id {
			newJobs = append(newJobs, j)
		}
	}
	m.Jobs = newJobs
	return nil
}

// ToggleUserJob enables or disables a user crontab job.
func (m *Manager) ToggleUserJob(id int) error {
	var target *Job
	for _, j := range m.Jobs {
		if j.ID == id {
			target = j
			break
		}
	}
	if target == nil || target.Source != SourceUser {
		return fmt.Errorf("job %d not found or not a user job", id)
	}
	target.Disabled = !target.Disabled
	return m.WriteUserCrontab()
}

func jobToLine(j *Job) string {
	line := fmt.Sprintf("%s %s", j.Schedule, j.Command)
	if j.Comment != "" {
		line += " # " + j.Comment
	}
	if j.Disabled {
		line = "#" + line
	}
	return line
}
