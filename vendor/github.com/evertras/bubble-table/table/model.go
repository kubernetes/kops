package table

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	columnKeySelect = "___select___"
)

var (
	defaultHighlightStyle = lipgloss.NewStyle().Background(lipgloss.Color("#334"))
)

// Model is the main table model.  Create using New().
type Model struct {
	// Data
	columns []Column
	rows    []Row

	// Caches for optimizations
	visibleRowCacheUpdated bool
	visibleRowCache        []Row

	// Shown when data is missing from a row
	missingDataIndicator interface{}

	// Interaction
	focused bool
	keyMap  KeyMap

	// Taken from: 'Bubbles/List'
	// Additional key mappings for the short and full help views. This allows
	// you to add additional key mappings to the help menu without
	// re-implementing the help component. Of course, you can also disable the
	// list's help component and implement a new one if you need more
	// flexibility.
	// You have to supply a keybinding like this:
	// key.NewBinding( key.WithKeys("shift+left"), key.WithHelp("shift+←", "scroll left"))
	// It needs both 'WithKeys' and 'WithHelp'
	additionalShortHelpKeys func() []key.Binding
	additionalFullHelpKeys  func() []key.Binding

	selectableRows bool
	rowCursorIndex int

	// Events
	lastUpdateUserEvents []UserEvent

	// Styles
	baseStyle      lipgloss.Style
	highlightStyle lipgloss.Style
	headerStyle    lipgloss.Style
	rowStyleFunc   func(RowStyleFuncInput) lipgloss.Style
	border         Border
	selectedText   string
	unselectedText string

	// Header
	headerVisible bool

	// Footers
	footerVisible bool
	staticFooter  string

	// Pagination
	pageSize           int
	currentPage        int
	paginationWrapping bool

	// Sorting, where a stable sort is applied from first element to last so
	// that elements are grouped by the later elements.
	sortOrder []SortColumn

	// Filter
	filtered        bool
	filterTextInput textinput.Model

	// For flex columns
	targetTotalWidth int

	// The maximum total width for overflow/scrolling
	maxTotalWidth int

	// Internal cached calculations for reference, may be higher than
	// maxTotalWidth.  If this is the case, we need to adjust the view
	totalWidth int

	// How far to scroll to the right, in columns
	horizontalScrollOffsetCol int

	// How many columns to freeze when scrolling horizontally
	horizontalScrollFreezeColumnsCount int

	// Calculated maximum column we can scroll to before the last is displayed
	maxHorizontalColumnIndex int

	// Minimum total height of the table
	minimumHeight int

	// Internal cached calculation, the height of the header and footer
	// including borders. Used to determine how many padding rows to add.
	metaHeight int

	// If true, the table will be multiline
	multiline bool
}

// New creates a new table ready for further modifications.
func New(columns []Column) Model {
	filterInput := textinput.New()
	filterInput.Prompt = "/"
	model := Model{
		columns:        make([]Column, len(columns)),
		highlightStyle: defaultHighlightStyle.Copy(),
		border:         borderDefault,
		headerVisible:  true,
		footerVisible:  true,
		keyMap:         DefaultKeyMap(),

		selectedText:   "[x]",
		unselectedText: "[ ]",

		filterTextInput: filterInput,
		baseStyle:       lipgloss.NewStyle().Align(lipgloss.Right),

		paginationWrapping: true,
	}

	// Do a full deep copy to avoid unexpected edits
	copy(model.columns, columns)

	model.recalculateWidth()

	return model
}

// Init initializes the table per the Bubble Tea architecture.
func (m Model) Init() tea.Cmd {
	return nil
}
