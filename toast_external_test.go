package toast_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevin-rieck/go-bubble-toast"
)

func TestPublicDismissNewestRemovesNewestVisibleToastAndDrainsQueue(t *testing.T) {
	model := toast.New(toast.WithPlacement(toast.TopRight), toast.WithMaxVisible(2), toast.WithDefaultDuration(0))
	model, _, _ = model.Push(toast.NewToast("old", toast.WithID("old")))
	model, _, _ = model.Push(toast.NewToast("new", toast.WithID("new")))
	model, _, _ = model.Push(toast.NewToast("queued", toast.WithID("queued")))

	var cmd tea.Cmd
	model, cmd = model.Update(toast.DismissNewest()())

	if cmd == nil {
		t.Fatal("dismissing newest visible Toast should drain queued Toast and schedule its timer")
	}
	visible := model.Visible()
	if len(visible) != 2 || visible[0].ID != "queued" || visible[1].ID != "old" {
		t.Fatalf("newest Dismissal should remove newest top Toast and drain queue, visible=%#v", visible)
	}
}

func TestPublicDismissOldestRemovesOldestVisibleBottomToastAndDrainsQueue(t *testing.T) {
	model := toast.New(toast.WithPlacement(toast.BottomRight), toast.WithMaxVisible(2))
	model, _, _ = model.Push(toast.NewToast("old", toast.WithID("old")))
	model, _, _ = model.Push(toast.NewToast("new", toast.WithID("new")))
	model, _, _ = model.Push(toast.NewToast("queued", toast.WithID("queued")))

	var cmd tea.Cmd
	model, cmd = model.Update(toast.DismissOldest()())

	if cmd == nil {
		t.Fatal("dismissing oldest visible Toast should drain queued Toast and schedule its timer")
	}
	visible := model.Visible()
	if len(visible) != 2 || visible[0].ID != "new" || visible[1].ID != "queued" {
		t.Fatalf("oldest Dismissal should remove oldest bottom Toast and drain queue, visible=%#v", visible)
	}
}

func TestPublicAPIErgonomicsForBubbleTeaApps(t *testing.T) {
	model := toast.New()
	msg := toast.Show(toast.Success("saved", toast.WithTitle("Done"), toast.WithID("save-status")))()

	var cmd tea.Cmd
	model, cmd = model.Update(msg)

	if cmd == nil {
		t.Fatal("message-based Show should schedule visible Toast Dismissal")
	}
	if model.Len() != 1 || model.Visible()[0].ID != "save-status" {
		t.Fatalf("Toast model did not route public Show message: %#v", model.Visible())
	}

	model, _ = model.Update(toast.Dismiss("save-status")())
	if model.Len() != 0 {
		t.Fatalf("public Dismiss command did not remove Toast, len=%d", model.Len())
	}
}
