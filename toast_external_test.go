package toast_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevin-rieck/go-bubble-toast"
)

func TestAppCanQueryVisibleToastByID(t *testing.T) {
	model := toast.New()
	model, _, _ = model.Push(toast.Info("connected", toast.WithID("network")))

	got, ok := model.VisibleByID("network")
	if !ok {
		t.Fatal("expected visible Toast to be found by Toast ID")
	}
	if got.ID != "network" || got.Message != "connected" {
		t.Fatalf("unexpected Toast returned: %#v", got)
	}
	if !model.IsVisible("network") {
		t.Fatal("expected Toast ID to be reported visible")
	}
}

func TestAppCanQueryQueuedToastByID(t *testing.T) {
	model := toast.New(toast.WithMaxVisible(1))
	model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible")))
	model, _, _ = model.Push(toast.Warning("waiting", toast.WithID("queued")))

	got, ok := model.QueuedByID("queued")
	if !ok {
		t.Fatal("expected queued Toast to be found by Toast ID")
	}
	if got.ID != "queued" || got.Message != "waiting" {
		t.Fatalf("unexpected Toast returned: %#v", got)
	}
	if !model.IsQueued("queued") {
		t.Fatal("expected Toast ID to be reported queued")
	}
	if model.IsVisible("queued") {
		t.Fatal("queued Toast ID should not be reported visible")
	}
}

func TestToastIDQueriesDoNotFindMissingOrDismissedToasts(t *testing.T) {
	model := toast.New(toast.WithMaxVisible(1))
	model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible")))
	model, _, _ = model.Push(toast.NewToast("queued", toast.WithID("queued")))

	if _, ok := model.VisibleByID("missing"); ok {
		t.Fatal("missing Toast ID should not be found in visible Toast Stack")
	}
	if _, ok := model.QueuedByID("missing"); ok {
		t.Fatal("missing Toast ID should not be found in queue")
	}

	model, _ = model.Dismiss("visible")
	model, _ = model.Dismiss("queued")

	if model.IsVisible("visible") || model.IsQueued("queued") {
		t.Fatal("dismissed Toast IDs should not be reported visible or queued")
	}
}

func TestToastIDQueriesReturnCopies(t *testing.T) {
	model := toast.New(toast.WithMaxVisible(1))
	model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible")))
	model, _, _ = model.Push(toast.NewToast("queued", toast.WithID("queued")))

	visible, ok := model.VisibleByID("visible")
	if !ok {
		t.Fatal("expected visible Toast")
	}
	visible.Message = "mutated"
	if got, _ := model.VisibleByID("visible"); got.Message == "mutated" {
		t.Fatal("VisibleByID should return a copy")
	}

	queued, ok := model.QueuedByID("queued")
	if !ok {
		t.Fatal("expected queued Toast")
	}
	queued.Message = "mutated"
	if got, _ := model.QueuedByID("queued"); got.Message == "mutated" {
		t.Fatal("QueuedByID should return a copy")
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
