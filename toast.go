package toast

import (
	"crypto/rand"
	"encoding/binary"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultDuration   = 5 * time.Second
	defaultMaxVisible = 3
	defaultMaxQueued  = 20
	defaultWidth      = 40
	defaultMaxHeight  = 4
	defaultGap        = 1
	defaultMargin     = 1
)

type ID string

type Kind string

const (
	KindNone    Kind = ""
	KindInfo    Kind = "info"
	KindSuccess Kind = "success"
	KindWarning Kind = "warning"
	KindError   Kind = "error"
)

type Placement int

const (
	TopLeft Placement = iota
	TopRight
	TopCenter
	BottomLeft
	BottomRight
	BottomCenter
)

// QueueOverflowPolicy controls what happens when a new Toast arrives while the queue is full.
type QueueOverflowPolicy int

const (
	// DropOldestQueuedToast drops the oldest queued Toast and queues the new Toast.
	DropOldestQueuedToast QueueOverflowPolicy = iota
	// DropNewestToast drops the new incoming Toast and preserves the existing queue.
	DropNewestToast
)

// RendererPreset configures a built-in Toast presentation.
type RendererPreset int

const (
	// PresetDefault uses Bubble Toast's default bordered presentation.
	PresetDefault RendererPreset = iota
	// PresetCompact renders a bordered Toast with reduced padding.
	PresetCompact
	// PresetMinimal renders Toast Content without borders or padding.
	PresetMinimal
	// PresetIcon renders built-in Toast Kind icons before Toast Content.
	PresetIcon
)

type ToastAction struct {
	Key   string
	Label string
	Cmd   tea.Cmd
}

type Toast struct {
	ID          ID
	Kind        Kind
	Title       string
	Message     string
	Content     string
	Actions     []ToastAction
	Occurrences int
	Duration    time.Duration
	Persistent  bool
}

type ShowMsg struct{ Toast Toast }
type DismissMsg struct{ ID ID }
type DismissNewestMsg struct{}
type DismissOldestMsg struct{}
type DismissAllMsg struct{}

type expirationMsg struct {
	id         ID
	generation uint64
	epoch      uint64
}

type progressTickMsg time.Time

type RenderContext struct {
	Width     int
	MaxHeight int
	Style     lipgloss.Style
	Placement Placement
	Index     int
	Total     int
}

type Renderer func(Toast, RenderContext) string

type QueueIndicatorContext struct {
	Count     int
	Placement Placement
}

type QueueIndicatorRenderer func(QueueIndicatorContext) string

type Theme struct {
	None    lipgloss.Style
	Info    lipgloss.Style
	Success lipgloss.Style
	Warning lipgloss.Style
	Error   lipgloss.Style
}

type Margin struct{ Top, Right, Bottom, Left int }

type entry struct {
	toast      Toast
	generation uint64
	createdAt  time.Time
	renderedAt time.Time
}

type Model struct {
	defaultDuration        time.Duration
	kindDurations          map[Kind]time.Duration
	maxVisible             int
	maxQueued              int
	placement              Placement
	width                  int
	maxHeight              int
	gap                    int
	margin                 Margin
	theme                  Theme
	renderer               Renderer
	queueIndicatorRenderer QueueIndicatorRenderer
	queueIndicatorEnabled  bool
	queueOverflowPolicy    QueueOverflowPolicy
	progressModel          *progress.Model
	iconsEnabled           bool
	kindIcons              map[Kind]string
	iconOverrides          map[Kind]bool
	asciiOnly              bool
	noColor                bool
	coalesceDuplicates     bool

	visible []entry
	queued  []entry
	nextID  uint64
	nextGen uint64
	epoch   uint64
	winW    int
	winH    int
}

type Option func(*Model)
type ToastOption func(*Toast)

func New(options ...Option) Model {
	m := Model{
		defaultDuration:       defaultDuration,
		maxVisible:            defaultMaxVisible,
		maxQueued:             defaultMaxQueued,
		placement:             TopRight,
		width:                 defaultWidth,
		maxHeight:             defaultMaxHeight,
		gap:                   defaultGap,
		margin:                Margin{defaultMargin, defaultMargin, defaultMargin, defaultMargin},
		theme:                 DefaultTheme(),
		queueIndicatorEnabled: true,
		nextID:                1,
		nextGen:               1,
		epoch:                 newEpoch(),
	}
	for _, opt := range options {
		opt(&m)
	}
	m.normalize()
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ShowMsg:
		updated, _, cmd := m.Push(msg.Toast)
		return updated, cmd
	case DismissMsg:
		return m.Dismiss(string(msg.ID))
	case DismissNewestMsg:
		return m.DismissNewest()
	case DismissOldestMsg:
		return m.DismissOldest()
	case DismissAllMsg:
		return m.DismissAll()
	case tea.KeyMsg:
		return m.invokeAction(msg)
	case expirationMsg:
		return m.expire(msg)
	case tea.WindowSizeMsg:
		m.winW, m.winH = msg.Width, msg.Height
		return m, nil
	case progress.FrameMsg:
		if m.progressModel != nil {
			progressModel, cmd := m.progressModel.Update(msg)
			p := progressModel.(progress.Model)
			m.progressModel = &p
			return m, cmd
		}
		return m, nil
	case progressTickMsg:
		if m.progressModel == nil || len(m.visible) == 0 {
			return m, nil
		}
		return m, tickProgress()
	default:
		return m, nil
	}
}

func Show(t Toast) tea.Cmd { return func() tea.Msg { return ShowMsg{Toast: t} } }

// Replace returns a command that updates or replaces the Toast with the given Toast ID.
func Replace(id string, t Toast) tea.Cmd {
	t.ID = ID(id)
	return Show(t)
}
func Dismiss(id string) tea.Cmd { return func() tea.Msg { return DismissMsg{ID: ID(id)} } }
func DismissNewest() tea.Cmd    { return func() tea.Msg { return DismissNewestMsg{} } }
func DismissOldest() tea.Cmd    { return func() tea.Msg { return DismissOldestMsg{} } }
func DismissAll() tea.Cmd       { return func() tea.Msg { return DismissAllMsg{} } }

func NewToast(message string, options ...ToastOption) Toast {
	return buildToast(message, KindNone, options...)
}
func Info(message string, options ...ToastOption) Toast {
	return buildToast(message, KindInfo, options...)
}
func Success(message string, options ...ToastOption) Toast {
	return buildToast(message, KindSuccess, options...)
}
func Warning(message string, options ...ToastOption) Toast {
	return buildToast(message, KindWarning, options...)
}
func Error(message string, options ...ToastOption) Toast {
	return buildToast(message, KindError, options...)
}

func buildToast(message string, kind Kind, options ...ToastOption) Toast {
	t := Toast{Kind: kind, Message: message}
	for _, opt := range options {
		opt(&t)
	}
	return t
}

// Replace updates or replaces the Toast with the given Toast ID.
func (m Model) Replace(id string, t Toast) (Model, tea.Cmd) {
	t.ID = ID(id)
	updated, _, cmd := m.Push(t)
	return updated, cmd
}

func (m Model) Push(t Toast) (Model, ID, tea.Cmd) {
	if t.ID == "" && m.coalesceDuplicates {
		if i := indexOfDuplicate(m.visible, t); i >= 0 {
			m.nextGen++
			m.visible[i].toast.Occurrences = nextOccurrenceCount(m.visible[i].toast.Occurrences)
			m.visible[i].generation = m.nextGen
			m.visible[i].renderedAt = time.Now()
			return m, m.visible[i].toast.ID, m.timer(m.visible[i])
		}
		if i := indexOfDuplicate(m.queued, t); i >= 0 {
			m.queued[i].toast.Occurrences = nextOccurrenceCount(m.queued[i].toast.Occurrences)
			return m, m.queued[i].toast.ID, nil
		}
	}
	if t.ID == "" {
		t.ID = m.generateID()
	}
	if i := indexOf(m.visible, t.ID); i >= 0 {
		m.nextGen++
		m.visible[i] = entry{toast: t, generation: m.nextGen, createdAt: m.visible[i].createdAt, renderedAt: time.Now()}
		return m, t.ID, m.timer(m.visible[i])
	}
	if i := indexOf(m.queued, t.ID); i >= 0 {
		m.nextGen++
		m.queued[i] = entry{toast: t, generation: m.nextGen, createdAt: m.queued[i].createdAt}
		return m, t.ID, nil
	}
	m.nextGen++
	e := entry{toast: t, generation: m.nextGen, createdAt: time.Now()}
	if len(m.visible) < m.maxVisible {
		e.renderedAt = time.Now()
		m.visible = append(m.visible, e)
		return m, t.ID, m.timer(e)
	}
	if m.maxQueued == 0 {
		return m, t.ID, nil
	}
	if len(m.queued) >= m.maxQueued {
		if m.queueOverflowPolicy == DropNewestToast {
			return m, t.ID, nil
		}
		m.queued = m.queued[1:]
	}
	m.queued = append(m.queued, e)
	return m, t.ID, nil
}

func (m Model) Dismiss(id string) (Model, tea.Cmd) {
	m.visible = removeID(m.visible, ID(id))
	m.queued = removeID(m.queued, ID(id))
	return m.drain()
}

func (m Model) DismissNewest() (Model, tea.Cmd) {
	if len(m.visible) == 0 {
		return m, nil
	}
	m.visible = m.visible[:len(m.visible)-1]
	return m.drain()
}

func (m Model) DismissOldest() (Model, tea.Cmd) {
	if len(m.visible) == 0 {
		return m, nil
	}
	m.visible = m.visible[1:]
	return m.drain()
}

func (m Model) DismissAll() (Model, tea.Cmd) {
	m.visible = nil
	m.queued = nil
	return m, nil
}

func (m Model) invokeAction(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()
	for _, e := range m.renderEntries() {
		for _, action := range e.toast.Actions {
			if action.Key == key {
				updated, drainCmd := m.Dismiss(string(e.toast.ID))
				return updated, tea.Batch(action.Cmd, drainCmd)
			}
		}
	}
	return m, nil
}

func (m Model) Visible() []Toast {
	entries := m.renderEntries()
	out := make([]Toast, len(entries))
	for i, e := range entries {
		out[i] = e.toast
	}
	return out
}

func (m Model) Queued() []Toast {
	out := make([]Toast, len(m.queued))
	for i, e := range m.queued {
		out[i] = e.toast
	}
	return out
}

// VisibleByID returns the visible Toast with the given Toast ID, if present.
func (m Model) VisibleByID(id string) (Toast, bool) {
	return findToast(m.visible, ID(id))
}

// IsVisible reports whether the given Toast ID is currently visible.
func (m Model) IsVisible(id string) bool {
	_, ok := m.VisibleByID(id)
	return ok
}

// QueuedByID returns the queued Toast with the given Toast ID, if present.
func (m Model) QueuedByID(id string) (Toast, bool) {
	return findToast(m.queued, ID(id))
}

// IsQueued reports whether the given Toast ID is currently queued.
func (m Model) IsQueued(id string) bool {
	_, ok := m.QueuedByID(id)
	return ok
}

func (m Model) Len() int { return len(m.visible) + len(m.queued) }

func (m Model) expire(msg expirationMsg) (Model, tea.Cmd) {
	if m.isStaleExpiration(msg) {
		return m, nil
	}
	m.visible = removeMatching(m.visible, msg.id, msg.generation)
	return m.drain()
}

func (m Model) drain() (Model, tea.Cmd) {
	var cmds []tea.Cmd
	for len(m.visible) < m.maxVisible && len(m.queued) > 0 {
		e := m.queued[0]
		m.queued = m.queued[1:]
		e.renderedAt = time.Now()
		m.visible = append(m.visible, e)
		if cmd := m.timer(e); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) generateID() ID {
	id := ID("toast-" + strconv64(m.nextID))
	m.nextID++
	return id
}

func indexOf(entries []entry, id ID) int {
	for i, e := range entries {
		if e.toast.ID == id {
			return i
		}
	}
	return -1
}

func indexOfDuplicate(entries []entry, t Toast) int {
	for i, e := range entries {
		if e.toast.Kind == t.Kind && e.toast.Message == t.Message {
			return i
		}
	}
	return -1
}

func nextOccurrenceCount(current int) int {
	if current < 2 {
		return 2
	}
	return current + 1
}

func findToast(entries []entry, id ID) (Toast, bool) {
	for _, e := range entries {
		if e.toast.ID == id {
			return e.toast, true
		}
	}
	return Toast{}, false
}

func removeID(entries []entry, id ID) []entry {
	out := entries[:0]
	for _, e := range entries {
		if e.toast.ID != id {
			out = append(out, e)
		}
	}
	return out
}

func removeMatching(entries []entry, id ID, gen uint64) []entry {
	out := entries[:0]
	for _, e := range entries {
		if !(e.toast.ID == id && e.generation == gen) {
			out = append(out, e)
		}
	}
	return out
}

func (m *Model) normalize() {
	if m.defaultDuration <= 0 {
		m.defaultDuration = defaultDuration
	}
	if m.maxVisible <= 0 {
		m.maxVisible = defaultMaxVisible
	}
	if m.maxQueued < 0 {
		m.maxQueued = defaultMaxQueued
	}
	if m.width <= 0 {
		m.width = defaultWidth
	}
	if m.maxHeight < 0 {
		m.maxHeight = defaultMaxHeight
	}
	if m.gap < 0 {
		m.gap = defaultGap
	}
	if m.margin.Top < 0 {
		m.margin.Top = 0
	}
	if m.margin.Right < 0 {
		m.margin.Right = 0
	}
	if m.margin.Bottom < 0 {
		m.margin.Bottom = 0
	}
	if m.margin.Left < 0 {
		m.margin.Left = 0
	}
}

func WithDefaultDuration(d time.Duration) Option { return func(m *Model) { m.defaultDuration = d } }

// WithKindDuration configures the default Toast Lifetime for a Toast Kind.
func WithKindDuration(kind Kind, d time.Duration) Option {
	return func(m *Model) {
		if d <= 0 {
			return
		}
		if m.kindDurations == nil {
			m.kindDurations = make(map[Kind]time.Duration)
		}
		m.kindDurations[kind] = d
	}
}
func WithMaxVisible(n int) Option { return func(m *Model) { m.maxVisible = n } }
func WithMaxQueued(n int) Option  { return func(m *Model) { m.maxQueued = n } }

// WithDuplicateCoalescing coalesces duplicate Toasts with matching Toast Kind and message.
func WithDuplicateCoalescing() Option {
	return func(m *Model) { m.coalesceDuplicates = true }
}

// WithoutDuplicateCoalescing disables duplicate Toast coalescing.
func WithoutDuplicateCoalescing() Option {
	return func(m *Model) { m.coalesceDuplicates = false }
}

// WithQueueOverflowPolicy configures which Toast is dropped when the queue is full.
func WithQueueOverflowPolicy(policy QueueOverflowPolicy) Option {
	return func(m *Model) { m.queueOverflowPolicy = policy }
}

func WithPlacement(p Placement) Option { return func(m *Model) { m.placement = p } }
func WithWidth(w int) Option           { return func(m *Model) { m.width = w } }
func WithMaxHeight(h int) Option       { return func(m *Model) { m.maxHeight = h } }
func WithGap(g int) Option             { return func(m *Model) { m.gap = g } }
func WithOverlayMargin(top, right, bottom, left int) Option {
	return func(m *Model) { m.margin = Margin{top, right, bottom, left} }
}
func WithTheme(t Theme) Option { return func(m *Model) { m.theme = t } }
func WithStyle(kind Kind, style lipgloss.Style) Option {
	return func(m *Model) { setStyle(&m.theme, kind, style) }
}
func WithRenderer(r Renderer) Option { return func(m *Model) { m.renderer = r } }

// WithRendererPreset selects a built-in Toast presentation preset.
func WithRendererPreset(p RendererPreset) Option {
	return func(m *Model) { applyRendererPreset(m, p) }
}

// WithQueueIndicator customizes the affordance shown when Toasts are queued.
func WithQueueIndicator(r QueueIndicatorRenderer) Option {
	return func(m *Model) {
		m.queueIndicatorEnabled = true
		m.queueIndicatorRenderer = r
	}
}

// WithoutQueueIndicator disables the affordance shown when Toasts are queued.
func WithoutQueueIndicator() Option {
	return func(m *Model) { m.queueIndicatorEnabled = false }
}

func WithProgress(enabled bool) Option {
	return func(m *Model) {
		if enabled {
			p := progress.New(
				progress.WithDefaultGradient(),
				progress.WithoutPercentage(),
				progress.WithFillCharacters('─', ' '),
			)
			m.progressModel = &p
		} else {
			m.progressModel = nil
		}
	}
}

// WithKindIcons enables default accessible icons for built-in Toast Kinds.
func WithKindIcons() Option {
	return func(m *Model) {
		m.iconsEnabled = true
		if m.asciiOnly {
			m.kindIcons = asciiKindIcons()
		} else {
			m.kindIcons = defaultKindIcons()
		}
	}
}

// WithNoColor configures built-in rendering affordances to avoid ANSI colors.
func WithNoColor() Option {
	return func(m *Model) { m.noColor = true }
}

// WithASCIIOnly configures built-in rendering affordances to avoid Unicode.
func WithASCIIOnly() Option {
	return func(m *Model) {
		m.asciiOnly = true
		m.theme = asciiTheme(m.theme)
		if m.iconsEnabled {
			icons := asciiKindIcons()
			for kind := range m.iconOverrides {
				icons[kind] = m.kindIcons[kind]
			}
			m.kindIcons = icons
		}
	}
}

// WithIcon overrides the icon rendered for a Toast Kind and enables icons.
func WithIcon(kind Kind, icon string) Option {
	return func(m *Model) {
		m.iconsEnabled = true
		if m.kindIcons == nil {
			if m.asciiOnly {
				m.kindIcons = asciiKindIcons()
			} else {
				m.kindIcons = defaultKindIcons()
			}
		}
		if m.iconOverrides == nil {
			m.iconOverrides = make(map[Kind]bool)
		}
		m.iconOverrides[kind] = true
		m.kindIcons[kind] = icon
	}
}

// WithoutIcons disables Toast Kind icon rendering.
func WithoutIcons() Option {
	return func(m *Model) { m.iconsEnabled = false }
}

func WithID(id string) ToastOption             { return func(t *Toast) { t.ID = ID(id) } }
func WithTitle(title string) ToastOption       { return func(t *Toast) { t.Title = title } }
func WithKind(kind Kind) ToastOption           { return func(t *Toast) { t.Kind = kind } }
func WithDuration(d time.Duration) ToastOption { return func(t *Toast) { t.Duration = d } }
func WithPersistent() ToastOption              { return func(t *Toast) { t.Persistent = true } }
func WithContent(content string) ToastOption   { return func(t *Toast) { t.Content = content } }

// WithAction adds a keyboard action hint and command to a Toast.
func WithAction(key, label string, cmd tea.Cmd) ToastOption {
	return func(t *Toast) {
		t.Actions = append(t.Actions, ToastAction{Key: key, Label: label, Cmd: cmd})
	}
}

func newEpoch() uint64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err == nil {
		return binary.LittleEndian.Uint64(b[:])
	}
	return uint64(time.Now().UnixNano())
}

func strconv64(n uint64) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
