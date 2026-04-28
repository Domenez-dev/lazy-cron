package cron

import (
	"fmt"
	"strconv"
	"strings"
)

var descMonthFull = [13]string{"", "January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December"}

var descWeekday = [7]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

// DescribeSchedule returns a human-readable description of a cron expression.
func DescribeSchedule(expr string) string {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return ""
	}
	switch strings.ToLower(expr) {
	case "@reboot":
		return "At system startup"
	case "@hourly":
		return "Every hour at :00"
	case "@daily", "@midnight":
		return "Every day at midnight"
	case "@weekly":
		return "Every Sunday at midnight"
	case "@monthly":
		return "On the 1st of every month at midnight"
	case "@yearly", "@annually":
		return "Every year on January 1st at midnight"
	}
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return "⚠  Invalid cron expression"
	}
	return buildCronDesc(parts[0], parts[1], parts[2], parts[3], parts[4])
}

func buildCronDesc(min, hour, dom, month, dow string) string {
	if min == "*" && hour == "*" && dom == "*" && month == "*" && dow == "*" {
		return "Every minute"
	}
	timeStr := buildCronTimeStr(min, hour)
	whenStr := buildCronWhenStr(dom, month, dow)
	if whenStr == "" {
		return cronCapitalize(timeStr)
	}
	return cronCapitalize(timeStr) + ", " + whenStr
}

func buildCronTimeStr(min, hour string) string {
	minStep, minIsStep := cronParseStep(min)
	hourStep, hourIsStep := cronParseStep(hour)
	minNum, minIsNum := cronParseNum(min)
	hourNum, hourIsNum := cronParseNum(hour)
	minFrom, minTo, minIsRange := cronParseRange(min)
	hourFrom, hourTo, hourIsRange := cronParseRange(hour)
	minAny := min == "*"
	hourAny := hour == "*"

	// Range cases
	switch {
	case minIsRange && hourAny:
		return fmt.Sprintf("at minutes %d\u2013%d of every hour", minFrom, minTo)
	case minAny && hourIsRange:
		return fmt.Sprintf("every minute, from %02d:xx to %02d:xx", hourFrom, hourTo)
	case minIsNum && hourIsRange:
		if minNum == 0 {
			return fmt.Sprintf("from %02d:00 to %02d:00", hourFrom, hourTo)
		}
		return fmt.Sprintf("at :%02d, from %02d:xx to %02d:xx", minNum, hourFrom, hourTo)
	case minIsRange && hourIsNum:
		return fmt.Sprintf("at minutes %d\u2013%d during %02d:xx", minFrom, minTo, hourNum)
	}

	switch {
	case minAny && hourAny:
		return "every minute"
	case minIsStep && hourAny:
		if minStep == 1 {
			return "every minute"
		}
		return fmt.Sprintf("every %d minutes", minStep)
	case minIsNum && hourAny:
		if minNum == 0 {
			return "every hour at :00"
		}
		return fmt.Sprintf("every hour at :%02d", minNum)
	case minAny && hourIsStep:
		return fmt.Sprintf("every %d hours", hourStep)
	case minIsNum && hourIsStep:
		if minNum == 0 {
			return fmt.Sprintf("every %d hours", hourStep)
		}
		return fmt.Sprintf("every %d hours at :%02d", hourStep, minNum)
	case minIsStep && hourIsNum:
		return fmt.Sprintf("every %d minutes during %02d:xx", minStep, hourNum)
	case minAny && hourIsNum:
		return fmt.Sprintf("every minute during %02d:xx", hourNum)
	case minIsNum && hourIsNum:
		if minNum == 0 && hourNum == 0 {
			return "at midnight (00:00)"
		}
		if minNum == 0 && hourNum == 12 {
			return "at noon (12:00)"
		}
		return fmt.Sprintf("at %02d:%02d", hourNum, minNum)
	default:
		return fmt.Sprintf("at %s:%s", hour, min)
	}
}

func buildCronWhenStr(dom, month, dow string) string {
	domAny := dom == "*"
	monthAny := month == "*"
	dowAny := dow == "*"
	if domAny && monthAny && dowAny {
		return ""
	}
	var parts []string
	if !monthAny {
		if n, ok := cronParseNum(month); ok && n >= 1 && n <= 12 {
			parts = append(parts, "in "+descMonthFull[n])
		} else if from, to, ok := cronParseRange(month); ok && from >= 1 && to <= 12 {
			parts = append(parts, "from "+descMonthFull[from]+" to "+descMonthFull[to])
		} else if n, ok := cronParseStep(month); ok {
			parts = append(parts, fmt.Sprintf("every %d months", n))
		} else {
			parts = append(parts, "in month "+month)
		}
	}
	if !domAny {
		if n, ok := cronParseNum(dom); ok {
			parts = append(parts, "on the "+cronOrdinal(n))
		} else if from, to, ok := cronParseRange(dom); ok {
			parts = append(parts, fmt.Sprintf("on the %s to %s", cronOrdinal(from), cronOrdinal(to)))
		} else if n, ok := cronParseStep(dom); ok {
			parts = append(parts, fmt.Sprintf("every %d days", n))
		} else {
			parts = append(parts, "on day "+dom)
		}
	}
	if !dowAny {
		if n, ok := cronParseNum(dow); ok && n >= 0 && n <= 7 {
			idx := n % 7
			parts = append(parts, "on "+descWeekday[idx]+"s")
		} else if from, to, ok := cronParseRange(dow); ok {
			switch {
			case from == 1 && to == 5:
				parts = append(parts, "on weekdays (Mon\u2013Fri)")
			case from == 0 && to == 6:
				parts = append(parts, "every day of the week")
			default:
				parts = append(parts, fmt.Sprintf("on %s to %s", descWeekday[from%7], descWeekday[to%7]))
			}
		} else {
			parts = append(parts, "on weekday "+dow)
		}
	}
	return strings.Join(parts, ", ")
}

func cronParseRange(s string) (int, int, bool) {
	idx := strings.Index(s, "-")
	if idx <= 0 {
		return 0, 0, false
	}
	from, e1 := strconv.Atoi(s[:idx])
	to, e2 := strconv.Atoi(s[idx+1:])
	if e1 != nil || e2 != nil {
		return 0, 0, false
	}
	return from, to, true
}

func cronParseStep(s string) (int, bool) {
	if !strings.HasPrefix(s, "*/") {
		return 0, false
	}
	n, err := strconv.Atoi(s[2:])
	return n, err == nil
}

func cronParseNum(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	return n, err == nil
}

func cronOrdinal(n int) string {
	switch n % 100 {
	case 11, 12, 13:
		return fmt.Sprintf("%dth", n)
	}
	switch n % 10 {
	case 1:
		return fmt.Sprintf("%dst", n)
	case 2:
		return fmt.Sprintf("%dnd", n)
	case 3:
		return fmt.Sprintf("%drd", n)
	default:
		return fmt.Sprintf("%dth", n)
	}
}

func cronCapitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
