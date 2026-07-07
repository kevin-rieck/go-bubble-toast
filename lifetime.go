package toast

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var expirationTick = tea.Tick

func (m Model) timer(e entry) tea.Cmd {
	if e.toast.Persistent {
		return nil
	}
	epoch := m.epoch
	cmds := []tea.Cmd{
		expirationTick(m.toastDuration(e.toast), func(time.Time) tea.Msg {
			return expirationMsg{id: e.toast.ID, generation: e.generation, epoch: epoch}
		}),
	}
	if m.progressModel != nil {
		cmds = append(cmds, tickProgress())
	}
	return tea.Batch(cmds...)
}

func (m Model) toastDuration(t Toast) time.Duration {
	if t.Duration != 0 {
		return t.Duration
	}
	if d := m.kindDurations[t.Kind]; d != 0 {
		return d
	}
	return m.defaultDuration
}

func (m Model) isStaleExpiration(msg expirationMsg) bool {
	if msg.epoch != m.epoch {
		return true
	}
	for _, e := range m.visible {
		if e.toast.ID == msg.id {
			return e.generation != msg.generation
		}
	}
	return false
}

func (m Model) progressFraction(e entry) float64 {
	percent := 1.0 - float64(time.Since(e.renderedAt))/float64(m.toastDuration(e.toast))
	if percent < 0 {
		return 0
	}
	if percent > 1 {
		return 1
	}
	return percent
}

func tickProgress() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return progressTickMsg(t)
	})
}
