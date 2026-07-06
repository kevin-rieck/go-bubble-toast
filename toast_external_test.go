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

func TestPublicReplaceHelperUpdatesVisibleToastByID(t *testing.T) {
	model := toast.New()
	model, _, _ = model.Push(toast.Info("syncing", toast.WithID("sync-status")))

	var cmd tea.Cmd
	model, cmd = model.Replace("sync-status", toast.Success("synced"))

	if cmd == nil {
		t.Fatal("replacing visible Toast should restart its Dismissal timer")
	}
	visible := model.Visible()
	if len(visible) != 1 {
		t.Fatalf("replacement should preserve one Toast per stable Toast ID, visible=%#v", visible)
	}
	if visible[0].ID != "sync-status" || visible[0].Kind != toast.KindSuccess || visible[0].Message != "synced" {
		t.Fatalf("visible Toast was not replaced by Toast ID: %#v", visible[0])
	}
}

func TestPublicReplaceCommandRoutesThroughBubbleTeaUpdate(t *testing.T) {
	model := toast.New()
	model, _, _ = model.Push(toast.Info("syncing", toast.WithID("sync-status")))

	var cmd tea.Cmd
	model, cmd = model.Update(toast.Replace("sync-status", toast.Success("synced"))())

	if cmd == nil {
		t.Fatal("command-based Replace should restart visible Toast Dismissal")
	}
	visible := model.Visible()
	if len(visible) != 1 || visible[0].ID != "sync-status" || visible[0].Message != "synced" {
		t.Fatalf("command-based Replace did not update stable Toast ID: %#v", visible)
	}
}

func TestPublicReplaceHelperUpdatesQueuedToastInPlace(t *testing.T) {
	model := toast.New(toast.WithMaxVisible(1))
	model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible")))
	model, _, _ = model.Push(toast.Info("first queued", toast.WithID("first")))
	model, _, _ = model.Push(toast.Warning("second queued", toast.WithID("second")))

	var cmd tea.Cmd
	model, cmd = model.Replace("first", toast.Success("updated first queued"))

	if cmd != nil {
		t.Fatal("replacing queued Toast should not start a timer until it becomes visible")
	}
	queued := model.Queued()
	if len(queued) != 2 || queued[0].ID != "first" || queued[1].ID != "second" {
		t.Fatalf("queued replacement should preserve queued position, queued=%#v", queued)
	}
	if queued[0].Kind != toast.KindSuccess || queued[0].Message != "updated first queued" {
		t.Fatalf("queued Toast was not replaced in place: %#v", queued[0])
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
