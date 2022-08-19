package textinput

import (
	"context"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	rw "github.com/mattn/go-runewidth"
)

const defaultBlinkSpeed = time.Millisecond * 530

// Internal ID management for text inputs. Necessary for blink integrity when
// multiple text inputs are involved.
var (
	lastID int
	idMtx  sync.Mutex
)

// Return the next ID we should use on the Model.
func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}

// initialBlinkMsg initializes cursor blinking.
type initialBlinkMsg struct{}

// blinkMsg signals that the cursor should blink. It contains metadata that
// allows us to tell if the blink message is the one we're expecting.
type blinkMsg struct {
	id  int
	tag int
}

// blinkCanceled is sent when a blink operation is canceled.
type blinkCanceled struct{}

// Internal messages for clipboard operations.
type pasteMsg string
type pasteErrMsg struct{ error }

// EchoMode sets the input behavior of the text input field.
type EchoMode int

const (
	// EchoNormal displays text as is. This is the default behavior.
	EchoNormal EchoMode = iota

	// EchoPassword displays the EchoCharacter mask instead of actual
	// characters.  This is commonly used for password fields.
	EchoPassword

	// EchoNone displays nothing as characters are entered. This is commonly
	// seen for password fields on the command line.
	EchoNone

	// EchoOnEdit.
)

// blinkCtx manages cursor blinking.
type blinkCtx struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// CursorMode describes the behavior of the cursor.
type CursorMode int

// Available cursor modes.
const (
	CursorBlink CursorMode = iota
	CursorStatic
	CursorHide
)

// String returns a the cursor mode in a human-readable format. This method is
// provisional and for informational purposes only.
func (c CursorMode) String() string {
	return [...]string{
		"blink",
		"static",
		"hidden",
	}[c]
}

// ValidateFunc is a function that returns an error if the input is invalid.
type ValidateFunc func(string) error

// Model is the Bubble Tea model for this text input element.
type Model struct {
	Err error

	// General settings.
	Prompt        string
	Placeholder   string
	BlinkSpeed    time.Duration
	EchoMode      EchoMode
	EchoCharacter rune

	// Styles. These will be applied as inline styles.
	//
	// For an introduction to styling with Lip Gloss see:
	// https://github.com/charmbracelet/lipgloss
	PromptStyle      lipgloss.Style
	TextStyle        lipgloss.Style
	BackgroundStyle  lipgloss.Style
	PlaceholderStyle lipgloss.Style
	CursorStyle      lipgloss.Style

	// CharLimit is the maximum amount of characters this input element will
	// accept. If 0 or less, there's no limit.
	CharLimit int

	// Width is the maximum number of characters that can be displayed at once.
	// It essentially treats the text field like a horizontally scrolling
	// viewport. If 0 or less this setting is ignored.
	Width int

	// The ID of this Model as it relates to other textinput Models.
	id int

	// The ID of the blink message we're expecting to receive.
	blinkTag int

	// Underlying text value.
	value []rune

	// focus indicates whether user input focus should be on this input
	// component. When false, ignore keyboard input and hide the cursor.
	focus bool

	// Cursor blink state.
	blink bool

	// Cursor position.
	pos int

	// Used to emulate a viewport when width is set and the content is
	// overflowing.
	offset      int
	offsetRight int

	// Used to manage cursor blink
	blinkCtx *blinkCtx

	// cursorMode determines the behavior of the cursor
	cursorMode CursorMode

	// Validate is a function that checks whether or not the text within the
	// input is valid. If it is not valid, the `Err` field will be set to the
	// error returned by the function. If the function is not defined, all
	// input is considered valid.
	Validate ValidateFunc
}

// New creates a new model with default settings.
func New() Model {
	return Model{
		Prompt:           "> ",
		BlinkSpeed:       defaultBlinkSpeed,
		EchoCharacter:    '*',
		CharLimit:        0,
		PlaceholderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),

		id:         nextID(),
		value:      nil,
		focus:      false,
		blink:      true,
		pos:        0,
		cursorMode: CursorBlink,

		blinkCtx: &blinkCtx{
			ctx: context.Background(),
		},
	}
}

// NewModel creates a new model with default settings.
//
// Deprecated. Use New instead.
var NewModel = New

// SetValue sets the value of the text input.
func (m *Model) SetValue(s string) {
	if m.Validate != nil {
		if err := m.Validate(s); err != nil {
			m.Err = err
			return
		}
	}

	m.Err = nil

	runes := []rune(s)
	if m.CharLimit > 0 && len(runes) > m.CharLimit {
		m.value = runes[:m.CharLimit]
	} else {
		m.value = runes
	}
	if m.pos == 0 || m.pos > len(m.value) {
		m.setCursor(len(m.value))
	}
	m.handleOverflow()
}

// Value returns the value of the text input.
func (m Model) Value() string {
	return string(m.value)
}

// Cursor returns the cursor position.
func (m Model) Cursor() int {
	return m.pos
}

// Blink returns whether or not to draw the cursor.
func (m Model) Blink() bool {
	return m.blink
}

// SetCursor moves the cursor to the given position. If the position is
// out of bounds the cursor will be moved to the start or end accordingly.
func (m *Model) SetCursor(pos int) {
	m.setCursor(pos)
}

// setCursor moves the cursor to the given position and returns whether or not
// the cursor blink should be reset. If the position is out of bounds the
// cursor will be moved to the start or end accordingly.
func (m *Model) setCursor(pos int) bool {
	m.pos = clamp(pos, 0, len(m.value))
	m.handleOverflow()

	// Show the cursor unless it's been explicitly hidden
	m.blink = m.cursorMode == CursorHide

	// Reset cursor blink if necessary
	return m.cursorMode == CursorBlink
}

// CursorStart moves the cursor to the start of the input field.
func (m *Model) CursorStart() {
	m.cursorStart()
}

// cursorStart moves the cursor to the start of the input field and returns
// whether or not the curosr blink should be reset.
func (m *Model) cursorStart() bool {
	return m.setCursor(0)
}

// CursorEnd moves the cursor to the end of the input field.
func (m *Model) CursorEnd() {
	m.cursorEnd()
}

// CursorMode returns the model's cursor mode. For available cursor modes, see
// type CursorMode.
func (m Model) CursorMode() CursorMode {
	return m.cursorMode
}

// SetCursorMode sets the model's cursor mode. This method returns a command.
//
// For available cursor modes, see type CursorMode.
func (m *Model) SetCursorMode(mode CursorMode) tea.Cmd {
	m.cursorMode = mode
	m.blink = m.cursorMode == CursorHide || !m.focus
	if mode == CursorBlink {
		return Blink
	}
	return nil
}

// cursorEnd moves the cursor to the end of the input field and returns whether
// the cursor should blink should reset.
func (m *Model) cursorEnd() bool {
	return m.setCursor(len(m.value))
}

// Focused returns the focus state on the model.
func (m Model) Focused() bool {
	return m.focus
}

// Focus sets the focus state on the model. When the model is in focus it can
// receive keyboard input and the cursor will be hidden.
func (m *Model) Focus() tea.Cmd {
	m.focus = true
	m.blink = m.cursorMode == CursorHide // show the cursor unless we've explicitly hidden it

	if m.cursorMode == CursorBlink && m.focus {
		return m.blinkCmd()
	}
	return nil
}

// Blur removes the focus state on the model.  When the model is blurred it can
// not receive keyboard input and the cursor will be hidden.
func (m *Model) Blur() {
	m.focus = false
	m.blink = true
}

// Reset sets the input to its default state with no input. Returns whether
// or not the cursor blink should reset.
func (m *Model) Reset() bool {
	m.value = nil
	return m.setCursor(0)
}

// handle a clipboard paste event, if supported. Returns whether or not the
// cursor blink should reset.
func (m *Model) handlePaste(v string) bool {
	paste := []rune(v)

	var availSpace int
	if m.CharLimit > 0 {
		availSpace = m.CharLimit - len(m.value)
	}

	// If the char limit's been reached cancel
	if m.CharLimit > 0 && availSpace <= 0 {
		return false
	}

	// If there's not enough space to paste the whole thing cut the pasted
	// runes down so they'll fit
	if m.CharLimit > 0 && availSpace < len(paste) {
		paste = paste[:len(paste)-availSpace]
	}

	// Stuff before and after the cursor
	head := m.value[:m.pos]
	tailSrc := m.value[m.pos:]
	tail := make([]rune, len(tailSrc))
	copy(tail, tailSrc)

	oldPos := m.pos

	// Insert pasted runes
	for _, r := range paste {
		head = append(head, r)
		m.pos++
		if m.CharLimit > 0 {
			availSpace--
			if availSpace <= 0 {
				break
			}
		}
	}

	// Put it all back together
	value := append(head, tail...)
	m.SetValue(string(value))

	if m.Err != nil {
		m.pos = oldPos
	}

	// Reset blink state if necessary and run overflow checks
	return m.setCursor(m.pos)
}

// If a max width is defined, perform some logic to treat the visible area
// as a horizontally scrolling viewport.
func (m *Model) handleOverflow() {
	if m.Width <= 0 || rw.StringWidth(string(m.value)) <= m.Width {
		m.offset = 0
		m.offsetRight = len(m.value)
		return
	}

	// Correct right offset if we've deleted characters
	m.offsetRight = min(m.offsetRight, len(m.value))

	if m.pos < m.offset {
		m.offset = m.pos

		w := 0
		i := 0
		runes := m.value[m.offset:]

		for i < len(runes) && w <= m.Width {
			w += rw.RuneWidth(runes[i])
			if w <= m.Width+1 {
				i++
			}
		}

		m.offsetRight = m.offset + i
	} else if m.pos >= m.offsetRight {
		m.offsetRight = m.pos

		w := 0
		runes := m.value[:m.offsetRight]
		i := len(runes) - 1

		for i > 0 && w < m.Width {
			w += rw.RuneWidth(runes[i])
			if w <= m.Width {
				i--
			}
		}

		m.offset = m.offsetRight - (len(runes) - 1 - i)
	}
}

// deleteBeforeCursor deletes all text before the cursor. Returns whether or
// not the cursor blink should be reset.
func (m *Model) deleteBeforeCursor() bool {
	m.value = m.value[m.pos:]
	m.offset = 0
	return m.setCursor(0)
}

// deleteAfterCursor deletes all text after the cursor. Returns whether or not
// the cursor blink should be reset. If input is masked delete everything after
// the cursor so as not to reveal word breaks in the masked input.
func (m *Model) deleteAfterCursor() bool {
	m.value = m.value[:m.pos]
	return m.setCursor(len(m.value))
}

// deleteWordLeft deletes the word left to the cursor. Returns whether or not
// the cursor blink should be reset.
func (m *Model) deleteWordLeft() bool {
	if m.pos == 0 || len(m.value) == 0 {
		return false
	}

	if m.EchoMode != EchoNormal {
		return m.deleteBeforeCursor()
	}

	// Linter note: it's critical that we acquire the initial cursor position
	// here prior to altering it via SetCursor() below. As such, moving this
	// call into the corresponding if clause does not apply here.
	oldPos := m.pos //nolint:ifshort

	blink := m.setCursor(m.pos - 1)
	for unicode.IsSpace(m.value[m.pos]) {
		if m.pos <= 0 {
			break
		}
		// ignore series of whitespace before cursor
		blink = m.setCursor(m.pos - 1)
	}

	for m.pos > 0 {
		if !unicode.IsSpace(m.value[m.pos]) {
			blink = m.setCursor(m.pos - 1)
		} else {
			if m.pos > 0 {
				// keep the previous space
				blink = m.setCursor(m.pos + 1)
			}
			break
		}
	}

	if oldPos > len(m.value) {
		m.value = m.value[:m.pos]
	} else {
		m.value = append(m.value[:m.pos], m.value[oldPos:]...)
	}

	return blink
}

// deleteWordRight deletes the word right to the cursor. Returns whether or not
// the cursor blink should be reset. If input is masked delete everything after
// the cursor so as not to reveal word breaks in the masked input.
func (m *Model) deleteWordRight() bool {
	if m.pos >= len(m.value) || len(m.value) == 0 {
		return false
	}

	if m.EchoMode != EchoNormal {
		return m.deleteAfterCursor()
	}

	oldPos := m.pos
	m.setCursor(m.pos + 1)
	for unicode.IsSpace(m.value[m.pos]) {
		// ignore series of whitespace after cursor
		m.setCursor(m.pos + 1)

		if m.pos >= len(m.value) {
			break
		}
	}

	for m.pos < len(m.value) {
		if !unicode.IsSpace(m.value[m.pos]) {
			m.setCursor(m.pos + 1)
		} else {
			break
		}
	}

	if m.pos > len(m.value) {
		m.value = m.value[:oldPos]
	} else {
		m.value = append(m.value[:oldPos], m.value[m.pos:]...)
	}

	return m.setCursor(oldPos)
}

// wordLeft moves the cursor one word to the left. Returns whether or not the
// cursor blink should be reset. If input is masked, move input to the start
// so as not to reveal word breaks in the masked input.
func (m *Model) wordLeft() bool {
	if m.pos == 0 || len(m.value) == 0 {
		return false
	}

	if m.EchoMode != EchoNormal {
		return m.cursorStart()
	}

	blink := false
	i := m.pos - 1
	for i >= 0 {
		if unicode.IsSpace(m.value[i]) {
			blink = m.setCursor(m.pos - 1)
			i--
		} else {
			break
		}
	}

	for i >= 0 {
		if !unicode.IsSpace(m.value[i]) {
			blink = m.setCursor(m.pos - 1)
			i--
		} else {
			break
		}
	}

	return blink
}

// wordRight moves the cursor one word to the right. Returns whether or not the
// cursor blink should be reset. If the input is masked, move input to the end
// so as not to reveal word breaks in the masked input.
func (m *Model) wordRight() bool {
	if m.pos >= len(m.value) || len(m.value) == 0 {
		return false
	}

	if m.EchoMode != EchoNormal {
		return m.cursorEnd()
	}

	blink := false
	i := m.pos
	for i < len(m.value) {
		if unicode.IsSpace(m.value[i]) {
			blink = m.setCursor(m.pos + 1)
			i++
		} else {
			break
		}
	}

	for i < len(m.value) {
		if !unicode.IsSpace(m.value[i]) {
			blink = m.setCursor(m.pos + 1)
			i++
		} else {
			break
		}
	}

	return blink
}

func (m Model) echoTransform(v string) string {
	switch m.EchoMode {
	case EchoPassword:
		return strings.Repeat(string(m.EchoCharacter), rw.StringWidth(v))
	case EchoNone:
		return ""

	default:
		return v
	}
}

// Update is the Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		m.blink = true
		return m, nil
	}

	var resetBlink bool

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyBackspace: // delete character before cursor
			m.Err = nil

			if msg.Alt {
				resetBlink = m.deleteWordLeft()
			} else {
				if len(m.value) > 0 {
					m.value = append(m.value[:max(0, m.pos-1)], m.value[m.pos:]...)
					if m.pos > 0 {
						resetBlink = m.setCursor(m.pos - 1)
					}
				}
			}
		case tea.KeyLeft, tea.KeyCtrlB:
			if msg.Alt { // alt+left arrow, back one word
				resetBlink = m.wordLeft()
				break
			}
			if m.pos > 0 { // left arrow, ^F, back one character
				resetBlink = m.setCursor(m.pos - 1)
			}
		case tea.KeyRight, tea.KeyCtrlF:
			if msg.Alt { // alt+right arrow, forward one word
				resetBlink = m.wordRight()
				break
			}
			if m.pos < len(m.value) { // right arrow, ^F, forward one character
				resetBlink = m.setCursor(m.pos + 1)
			}
		case tea.KeyCtrlW: // ^W, delete word left of cursor
			resetBlink = m.deleteWordLeft()
		case tea.KeyHome, tea.KeyCtrlA: // ^A, go to beginning
			resetBlink = m.cursorStart()
		case tea.KeyDelete, tea.KeyCtrlD: // ^D, delete char under cursor
			if len(m.value) > 0 && m.pos < len(m.value) {
				m.value = append(m.value[:m.pos], m.value[m.pos+1:]...)
			}
		case tea.KeyCtrlE, tea.KeyEnd: // ^E, go to end
			resetBlink = m.cursorEnd()
		case tea.KeyCtrlK: // ^K, kill text after cursor
			resetBlink = m.deleteAfterCursor()
		case tea.KeyCtrlU: // ^U, kill text before cursor
			resetBlink = m.deleteBeforeCursor()
		case tea.KeyCtrlV: // ^V paste
			return m, Paste
		case tea.KeyRunes, tea.KeySpace: // input regular characters
			if msg.Alt && len(msg.Runes) == 1 {
				if msg.Runes[0] == 'd' { // alt+d, delete word right of cursor
					resetBlink = m.deleteWordRight()
					break
				}
				if msg.Runes[0] == 'b' { // alt+b, back one word
					resetBlink = m.wordLeft()
					break
				}
				if msg.Runes[0] == 'f' { // alt+f, forward one word
					resetBlink = m.wordRight()
					break
				}
			}

			// Input a regular character
			if m.CharLimit <= 0 || len(m.value) < m.CharLimit {
				runes := msg.Runes

				value := make([]rune, len(m.value))
				copy(value, m.value)
				value = append(value[:m.pos], append(runes, value[m.pos:]...)...)
				m.SetValue(string(value))
				if m.Err == nil {
					resetBlink = m.setCursor(m.pos + len(runes))
				}
			}
		}

	case initialBlinkMsg:
		// We accept all initialBlinkMsgs genrated by the Blink command.

		if m.cursorMode != CursorBlink || !m.focus {
			return m, nil
		}

		cmd := m.blinkCmd()
		return m, cmd

	case blinkMsg:
		// We're choosy about whether to accept blinkMsgs so that our cursor
		// only exactly when it should.

		// Is this model blinkable?
		if m.cursorMode != CursorBlink || !m.focus {
			return m, nil
		}

		// Were we expecting this blink message?
		if msg.id != m.id || msg.tag != m.blinkTag {
			return m, nil
		}

		var cmd tea.Cmd
		if m.cursorMode == CursorBlink {
			m.blink = !m.blink
			cmd = m.blinkCmd()
		}
		return m, cmd

	case blinkCanceled: // no-op
		return m, nil

	case pasteMsg:
		resetBlink = m.handlePaste(string(msg))

	case pasteErrMsg:
		m.Err = msg
	}

	var cmd tea.Cmd
	if resetBlink {
		cmd = m.blinkCmd()
	}

	m.handleOverflow()
	return m, cmd
}

// View renders the textinput in its current state.
func (m Model) View() string {
	// Placeholder text
	if len(m.value) == 0 && m.Placeholder != "" {
		return m.placeholderView()
	}

	styleText := m.TextStyle.Inline(true).Render

	value := m.value[m.offset:m.offsetRight]
	pos := max(0, m.pos-m.offset)
	v := styleText(m.echoTransform(string(value[:pos])))

	if pos < len(value) {
		v += m.cursorView(m.echoTransform(string(value[pos]))) // cursor and text under it
		v += styleText(m.echoTransform(string(value[pos+1:]))) // text after cursor
	} else {
		v += m.cursorView(" ")
	}

	// If a max width and background color were set fill the empty spaces with
	// the background color.
	valWidth := rw.StringWidth(string(value))
	if m.Width > 0 && valWidth <= m.Width {
		padding := max(0, m.Width-valWidth)
		if valWidth+padding <= m.Width && pos < len(value) {
			padding++
		}
		v += styleText(strings.Repeat(" ", padding))
	}

	return m.PromptStyle.Render(m.Prompt) + v
}

// placeholderView returns the prompt and placeholder view, if any.
func (m Model) placeholderView() string {
	var (
		v     string
		p     = m.Placeholder
		style = m.PlaceholderStyle.Inline(true).Render
	)

	// Cursor
	if m.blink {
		v += m.cursorView(style(p[:1]))
	} else {
		v += m.cursorView(p[:1])
	}

	// The rest of the placeholder text
	v += style(p[1:])

	return m.PromptStyle.Render(m.Prompt) + v
}

// cursorView styles the cursor.
func (m Model) cursorView(v string) string {
	if m.blink {
		return m.TextStyle.Render(v)
	}
	return m.CursorStyle.Inline(true).Reverse(true).Render(v)
}

// blinkCmd is an internal command used to manage cursor blinking.
func (m *Model) blinkCmd() tea.Cmd {
	if m.cursorMode != CursorBlink {
		return nil
	}

	if m.blinkCtx != nil && m.blinkCtx.cancel != nil {
		m.blinkCtx.cancel()
	}

	ctx, cancel := context.WithTimeout(m.blinkCtx.ctx, m.BlinkSpeed)
	m.blinkCtx.cancel = cancel

	m.blinkTag++

	return func() tea.Msg {
		defer cancel()
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			return blinkMsg{id: m.id, tag: m.blinkTag}
		}
		return blinkCanceled{}
	}
}

// Blink is a command used to initialize cursor blinking.
func Blink() tea.Msg {
	return initialBlinkMsg{}
}

// Paste is a command for pasting from the clipboard into the text input.
func Paste() tea.Msg {
	str, err := clipboard.ReadAll()
	if err != nil {
		return pasteErrMsg{err}
	}
	return pasteMsg(str)
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
