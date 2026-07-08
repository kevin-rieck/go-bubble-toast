package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	gutterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	emptyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)
)

// Columns renders example instructions beside Toast output so Toasts never cover
// the information the example is trying to teach.
func Columns(text, toasts string) string {
	if strings.TrimSpace(toasts) == "" {
		toasts = emptyStyle.Render("Toasts appear here")
	}
	gutter := gutterStyle.Render("  │  ")
	return lipgloss.JoinHorizontal(lipgloss.Top, text, gutter, toasts)
}
