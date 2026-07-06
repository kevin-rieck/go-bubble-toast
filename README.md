# Bubble Toast

Bubble Toast is a Bubble Tea/Lip Gloss component for transient, non-blocking Toasts in terminal UIs.

## Demo

![Bubble Toast feature showcase](assets/showcase.gif)

### More examples

| Startup Toast | Toast stack | Progress bar |
| --- | --- | --- |
| ![Startup Toast](assets/startup.gif) | ![Multiple Toasts](assets/burst.gif) | ![Progress bar Toasts](assets/progress.gif) |

These GIFs were created with [VHS](https://github.com/charmbracelet/vhs).

## Run the example

```sh
go run ./examples/basic
```

Press `t` to show Toasts and `q` to quit.

## Basic usage

```go
type model struct {
    toasts toast.Model
}

func (m model) Init() tea.Cmd {
    return toast.Show(toast.Info("ready", toast.WithTitle("App")))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.toasts, cmd = m.toasts.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return m.toasts.Overlay("your app view")
}
```

Small apps can also push directly:

```go
m := toast.New()
m, id, cmd := m.Push(toast.Success("saved"))
_, _ = id, cmd
```

## Updating a long-running status Toast

Use a stable Toast ID when a status may change over time. `Replace` reuses the
same Toast ID, so the Toast Stack or queue keeps one entry for that status.

```go
const syncStatus = "sync-status"

cmd := toast.Replace(syncStatus, toast.Info("syncing", toast.WithPersistent()))

// Later, when the work finishes:
cmd = toast.Replace(syncStatus, toast.Success("synced"))
```

Direct model updates use the same replacement behavior:

```go
m, cmd = m.Replace(syncStatus, toast.Error("sync failed"))
```

## Release

See [CHANGELOG.md](CHANGELOG.md) for release notes.
