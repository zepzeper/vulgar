package cli

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Modern, cohesive color palette - easier on the eyes
	ColorPrimary   = lipgloss.Color("#5B8DEF") // Soft blue - primary actions
	ColorSecondary = lipgloss.Color("#10B981") // Green - success
	ColorMuted     = lipgloss.Color("#9CA3AF") // Lighter gray - better contrast
	ColorError     = lipgloss.Color("#F87171") // Softer red
	ColorWarning   = lipgloss.Color("#FBBF24") // Warmer amber
	ColorInfo      = lipgloss.Color("#60A5FA") // Bright blue
	ColorCode      = lipgloss.Color("#C4B5FD") // Soft purple
	ColorCodeBg    = lipgloss.Color("#1F2937") // Dark gray
	ColorAccent    = lipgloss.Color("#8B5CF6") // Accent purple
)

var (
	StyleBold = lipgloss.NewStyle().Bold(true)

	StylePrimary = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	StyleSecondary = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	StyleMuted = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleError = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	StyleWarning = lipgloss.NewStyle().
			Foreground(ColorWarning)

	StyleInfo = lipgloss.NewStyle().
			Foreground(ColorInfo)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	StyleCode = lipgloss.NewStyle().
			Foreground(ColorCode).
			Background(ColorCodeBg).
			Padding(0, 1)

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	StyleSubtitle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	StyleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	StyleBoxSuccess = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Padding(0, 1)

	StyleBoxError = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(0, 1)

	StyleBoxWarning = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorWarning).
			Padding(0, 1)
)

// Symbols - text-based
var (
	SymbolSuccess = "+"
	SymbolError   = "x"
	SymbolWarning = "!"
	SymbolInfo    = "*"
	SymbolArrow   = ">"
	SymbolBullet  = "-"
	SymbolCheck   = "[x]"
	SymbolUncheck = "[ ]"
	SymbolLoading = "..."
	SymbolPipe    = "|"
	SymbolDash    = "-"
	SymbolDot     = "."
	SymbolFolder  = "[D]"
	SymbolFile    = "[F]"
	SymbolLock    = "[L]"
	SymbolUnlock  = "[U]"
	SymbolUp      = "^"
	SymbolDown    = "v"
	SymbolLeft    = "<"
	SymbolRight   = ">"
)

func Primary(s string) string   { return StylePrimary.Render(s) }
func Secondary(s string) string { return StyleSecondary.Render(s) }
func Muted(s string) string     { return StyleMuted.Render(s) }
func Error(s string) string     { return StyleError.Render(s) }
func Warning(s string) string   { return StyleWarning.Render(s) }
func Info(s string) string      { return StyleInfo.Render(s) }
func Success(s string) string   { return StyleSuccess.Render(s) }
func Code(s string) string      { return StyleCode.Render(s) }
func Title(s string) string     { return StyleTitle.Render(s) }
func Subtitle(s string) string  { return StyleSubtitle.Render(s) }
func Bold(s string) string      { return StyleBold.Render(s) }

// CustomHuhTheme creates a custom theme for huh components with improved colors
func CustomHuhTheme() *huh.Theme {
	return &huh.Theme{
		Form: huh.FormStyles{
			Base: lipgloss.NewStyle().
				Padding(1, 0),
		},
		Group: huh.GroupStyles{
			Base: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorAccent).
				Padding(1, 2),
			Title: lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true),
			Description: lipgloss.NewStyle().
				Foreground(ColorMuted),
		},
		FieldSeparator: lipgloss.NewStyle().
			Foreground(ColorMuted).
			SetString(" • "),
		Blurred: huh.FieldStyles{
			Base: lipgloss.NewStyle().
				Foreground(ColorMuted),
			Title: lipgloss.NewStyle().
				Foreground(ColorMuted),
			Description: lipgloss.NewStyle().
				Foreground(ColorMuted),
			Option: lipgloss.NewStyle().
				Foreground(ColorMuted),
			UnselectedOption: lipgloss.NewStyle().
				Foreground(ColorMuted),
		},
		Focused: huh.FieldStyles{
			Base: lipgloss.NewStyle().
				Foreground(ColorPrimary),
			Title: lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true),
			Description: lipgloss.NewStyle().
				Foreground(ColorMuted),
			SelectSelector: lipgloss.NewStyle().
				Foreground(ColorPrimary).
				SetString("▸"),
			Option: lipgloss.NewStyle().
				Foreground(ColorPrimary),
			SelectedOption: lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true),
			SelectedPrefix: lipgloss.NewStyle().
				Foreground(ColorSecondary).
				SetString("✓ "),
			UnselectedOption: lipgloss.NewStyle().
				Foreground(ColorMuted),
			UnselectedPrefix: lipgloss.NewStyle().
				Foreground(ColorMuted).
				SetString("  "),
			FocusedButton: lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true),
			BlurredButton: lipgloss.NewStyle().
				Foreground(ColorMuted),
			ErrorIndicator: lipgloss.NewStyle().
				Foreground(ColorError),
			ErrorMessage: lipgloss.NewStyle().
				Foreground(ColorError),
			TextInput: huh.TextInputStyles{
				Text: lipgloss.NewStyle().
					Foreground(ColorPrimary),
				Cursor: lipgloss.NewStyle().
					Foreground(ColorPrimary),
				CursorText: lipgloss.NewStyle().
					Foreground(ColorPrimary),
				Placeholder: lipgloss.NewStyle().
					Foreground(ColorMuted),
				Prompt: lipgloss.NewStyle().
					Foreground(ColorPrimary).
					Bold(true),
			},
		},
		Help: help.Styles{
			Ellipsis: lipgloss.NewStyle().
				Foreground(ColorMuted),
			ShortKey: lipgloss.NewStyle().
				Foreground(ColorPrimary),
			ShortDesc: lipgloss.NewStyle().
				Foreground(ColorMuted),
			ShortSeparator: lipgloss.NewStyle().
				Foreground(ColorMuted),
			FullKey: lipgloss.NewStyle().
				Foreground(ColorPrimary),
			FullDesc: lipgloss.NewStyle().
				Foreground(ColorMuted),
			FullSeparator: lipgloss.NewStyle().
				Foreground(ColorMuted),
		},
	}
}
