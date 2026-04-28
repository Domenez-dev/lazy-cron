package styles

import "github.com/charmbracelet/lipgloss"

var (
	ColorBg     = lipgloss.AdaptiveColor{Light: "#f5f5f5", Dark: "#1a1b26"}
	ColorFg     = lipgloss.AdaptiveColor{Light: "#24283b", Dark: "#c0caf5"}
	ColorMuted  = lipgloss.AdaptiveColor{Light: "#9699b0", Dark: "#565f89"}
	ColorAccent = lipgloss.AdaptiveColor{Light: "#7c3aed", Dark: "#bb9af7"}
	ColorGreen  = lipgloss.AdaptiveColor{Light: "#16a34a", Dark: "#9ece6a"}
	ColorRed    = lipgloss.AdaptiveColor{Light: "#dc2626", Dark: "#f7768e"}
	ColorYellow = lipgloss.AdaptiveColor{Light: "#ca8a04", Dark: "#e0af68"}
	ColorBlue   = lipgloss.AdaptiveColor{Light: "#2563eb", Dark: "#7aa2f7"}
	ColorCyan   = lipgloss.AdaptiveColor{Light: "#0891b2", Dark: "#7dcfff"}
	ColorBorder = lipgloss.AdaptiveColor{Light: "#d1d5db", Dark: "#3b4261"}
	ColorSelect = lipgloss.AdaptiveColor{Light: "#ede9fe", Dark: "#2d2f4e"}

	TitleStyle        = lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).PaddingLeft(1)
	SubtitleStyle     = lipgloss.NewStyle().Foreground(ColorMuted).PaddingLeft(1)
	HeaderStyle       = lipgloss.NewStyle().Bold(true).Foreground(ColorFg).Background(ColorBorder).Padding(0, 1)
	SelectedRowStyle  = lipgloss.NewStyle().Background(ColorSelect).Foreground(ColorAccent).Bold(true)
	NormalRowStyle    = lipgloss.NewStyle().Foreground(ColorFg)
	MutedStyle        = lipgloss.NewStyle().Foreground(ColorMuted)
	EnabledStyle      = lipgloss.NewStyle().Foreground(ColorGreen).Bold(true)
	DisabledStyle     = lipgloss.NewStyle().Foreground(ColorRed).Strikethrough(true)
	SystemStyle       = lipgloss.NewStyle().Foreground(ColorYellow)
	UserStyle         = lipgloss.NewStyle().Foreground(ColorBlue)
	CrondStyle        = lipgloss.NewStyle().Foreground(ColorCyan)
	StatusBarStyle    = lipgloss.NewStyle().Foreground(ColorMuted).PaddingLeft(1)
	KeyStyle          = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	KeyDescStyle      = lipgloss.NewStyle().Foreground(ColorMuted)
	ErrorStyle        = lipgloss.NewStyle().Foreground(ColorRed).Bold(true)
	SuccessStyle      = lipgloss.NewStyle().Foreground(ColorGreen).Bold(true)
	BorderStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder)
	InputStyle        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorBorder).Padding(0, 1)
	ActiveInputStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ColorAccent).Padding(0, 1)
	SectionTitleStyle = lipgloss.NewStyle().Foreground(ColorFg).Bold(true)
	CronExprStyle     = lipgloss.NewStyle().Foreground(ColorCyan).Bold(true)
	DescriptionStyle  = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	HelpBarStyle      = lipgloss.NewStyle().Foreground(ColorMuted).BorderTop(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(ColorBorder).PaddingLeft(1)
)

const AppName = "lazy-chrony"
const AppVersion = "0.1.0"
