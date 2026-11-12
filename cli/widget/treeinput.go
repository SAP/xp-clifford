package widget

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/SAP/xp-clifford/erratt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/x/ansi"
)

const (
	treeVisibleDepth  = 5
	treeIndentWidth   = 4
	treeDefaultWidth  = 80
	treeDefaultHeight = 10

	treeFilterPrompt         = "/"
	treeFilterPlaceholder    = "Filter options"
	treeOmittedLevels        = "..."
	treeTruncationMarker     = "…"
	treeNoMatchesText        = "No matching options"
	treeSelectionCountFormat = "%d/%d selected"
)

// TreeOption is a selectable node.
type TreeOption = TreeOptionOfType[string]

// TreeOptionOfType is a selectable node whose label is converted to display
// text by the TypeConverter passed to TreeInputOfType.
type TreeOptionOfType[T any] struct {
	Label    T
	Children []TreeOptionOfType[T]
}

// NewTreeOption creates a string node with optional children.
func NewTreeOption(label string, children ...TreeOption) TreeOption {
	return NewTreeOptionOfType(label, children...)
}

// NewTreeOptionOfType creates a typed node with optional children.
func NewTreeOptionOfType[T any](label T, children ...TreeOptionOfType[T]) TreeOptionOfType[T] {
	return TreeOptionOfType[T]{Label: label, Children: children}
}

// TreeInput presents an interactive multi-select tree and returns selected labels.
//
// Selecting or clearing a parent applies to its entire subtree. Toggling a
// child leaves its ancestors unchanged. The counter includes all nodes, even
// hidden descendants. Selected labels are returned in depth-first input order,
// once per selected node, including repeated labels.
func TreeInput(ctx context.Context, title string, options []TreeOption) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(options) == 0 {
		return []string{}, nil
	}
	treeSelect := newTreeSelect(title, options)
	g := huh.NewGroup(treeSelect).WithShowHelp(true)
	f := huh.NewForm(g).WithTheme(huh.ThemeCatppuccin()).WithHeight(treeDefaultHeight)
	if err := f.RunWithContext(ctx); err != nil {
		return nil, err
	}
	return treeSelect.values(), nil
}

// TreeInputOfType works like TreeInput for labels of type T. Like
// MultiInputOfType, it uses converter.FromType to display options and
// converter.ToType to convert selected labels back to T.
// The converter should round-trip labels without losing information.
func TreeInputOfType[T any](ctx context.Context, title string, options []TreeOptionOfType[T], converter TypeConverter[T]) ([]T, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	in, err := treeOptionsToStrings(options, converter)
	if err != nil {
		return nil, err
	}
	outStrings, err := TreeInput(ctx, title, in)
	if err != nil {
		return nil, err
	}
	out := make([]T, len(outStrings))
	for i, label := range outStrings {
		out[i], err = converter.ToType(label)
		if err != nil {
			return nil, erratt.Errorf("error converting from string: %w", err).With("value", label)
		}
	}
	return out, nil
}

func treeOptionsToStrings[T any](options []TreeOptionOfType[T], converter TypeConverter[T]) ([]TreeOption, error) {
	out := make([]TreeOption, len(options))
	for i, option := range options {
		label, err := converter.FromType(option.Label)
		if err != nil {
			return nil, erratt.Errorf("error converting to string: %w", err).With("value", option.Label)
		}
		children, err := treeOptionsToStrings(option.Children, converter)
		if err != nil {
			return nil, err
		}
		out[i] = NewTreeOption(label, children...)
	}
	return out, nil
}

// Tree state is separate from the caller's options. Each node has its own
// identity, so repeated labels remain independent.
type treeNode struct {
	label    string
	parent   *treeNode
	children []*treeNode
	depth    int
	selected bool
	visible  bool
}

func newTreeNode(option TreeOption, parent *treeNode) *treeNode {
	node := &treeNode{label: option.Label, parent: parent}
	if parent != nil {
		node.depth = parent.depth + 1
	}
	for _, child := range option.Children {
		node.children = append(node.children, newTreeNode(child, node))
	}
	return node
}

func walkTree(nodes []*treeNode, visit func(*treeNode)) {
	for _, node := range nodes {
		visit(node)
		walkTree(node.children, visit)
	}
}

func (n *treeNode) setSelection(selected bool) {
	n.selected = selected
	for _, child := range n.children {
		child.setSelection(selected)
	}
}

// treeSelect implements huh.Field. nodes holds all nodes; filtered contains
// matching nodes and their ancestors for navigation. displayRows limits depth.
type treeSelect struct {
	title          string
	roots          []*treeNode
	nodes          []*treeNode
	filtered       []*treeNode
	filter         textinput.Model
	filtering      bool
	cursor, top    int
	width, height  int
	terminalHeight int
	focused        bool
	accessible     bool
	theme          *huh.Theme
	keys           huh.MultiSelectKeyMap
	position       huh.FieldPosition
}

var _ huh.Field = (*treeSelect)(nil)

func newTreeSelect(title string, options []TreeOption) *treeSelect {
	filter := textinput.New()
	filter.Prompt = treeFilterPrompt
	filter.Placeholder = treeFilterPlaceholder
	m := &treeSelect{
		title: title, width: treeDefaultWidth, height: treeDefaultHeight,
		theme:  huh.ThemeCatppuccin(),
		filter: filter,
	}
	for _, option := range options {
		m.roots = append(m.roots, newTreeNode(option, nil))
	}
	walkTree(m.roots, func(node *treeNode) { m.nodes = append(m.nodes, node) })
	m.WithKeyMap(huh.NewDefaultKeyMap())
	m.WithTheme(m.theme)
	m.applyFilter()
	return m
}

func (m *treeSelect) applyFilter() {
	query := strings.ToLower(m.filter.Value())
	for _, node := range m.nodes {
		node.visible = false
	}
	var firstMatch *treeNode
	for _, node := range m.nodes {
		if strings.Contains(strings.ToLower(node.label), query) {
			if firstMatch == nil {
				firstMatch = node
			}
			for ancestor := node; ancestor != nil && !ancestor.visible; ancestor = ancestor.parent {
				ancestor.visible = true
			}
		}
	}
	m.filtered = m.filtered[:0]
	m.cursor, m.top = 0, 0
	for _, node := range m.nodes {
		if node.visible {
			if node == firstMatch {
				m.cursor = len(m.filtered)
			}
			m.filtered = append(m.filtered, node)
		}
	}
	m.WithPosition(m.position)
	m.ensureVisible()
}

func (m *treeSelect) setFiltering(filtering bool) {
	m.filtering = filtering
	if !filtering {
		m.filter.Blur()
	}
	m.WithPosition(m.position)
	m.ensureVisible()
}

func (m *treeSelect) values() []string {
	values := []string{}
	for _, node := range m.nodes {
		if node.selected {
			values = append(values, node.label)
		}
	}
	return values
}

type treeRow struct {
	node  *treeNode // nil represents omitted levels ("...").
	depth int
}

func (m *treeSelect) displayRows() []treeRow {
	if len(m.filtered) == 0 {
		return nil
	}
	// Keep the focused node and its nearest ancestors in the depth window.
	anchor := m.filtered[m.cursor]
	for level := 1; level < treeVisibleDepth && anchor.parent != nil; level++ {
		anchor = anchor.parent
	}
	root := anchor
	for root.parent != nil {
		root = root.parent
	}
	var rows []treeRow
	var appendBranch func(*treeNode, int, int)
	appendBranch = func(node *treeNode, depth, levels int) {
		if !node.visible {
			return
		}
		rows = append(rows, treeRow{node: node, depth: depth})
		for _, child := range node.children {
			if !child.visible {
				continue
			}
			if levels == 1 {
				rows = append(rows, treeRow{depth: depth + 1})
				return
			}
			appendBranch(child, depth+1, levels-1)
		}
	}
	for _, candidate := range m.roots {
		if candidate == root && anchor != root {
			rows = append(rows, treeRow{})
			appendBranch(anchor, 1, treeVisibleDepth)
		} else {
			appendBranch(candidate, 0, treeVisibleDepth)
		}
	}
	return rows
}

func (m *treeSelect) Init() tea.Cmd { return nil }

func (m *treeSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.terminalHeight = max(1, size.Height)
	}
	if m.focused && m.filtering {
		if input, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(input, m.keys.SetFilter):
				// Match MultiInput: finishing an empty result restores all options.
				if len(m.filtered) == 0 {
					m.filter.SetValue("")
					m.applyFilter()
				}
				m.setFiltering(false)
				return m, nil
			case input.String() == "up", input.String() == "down":
				if input.String() == "up" {
					m.cursor = max(0, m.cursor-1)
				} else {
					m.cursor = max(0, min(len(m.filtered)-1, m.cursor+1))
				}
				m.ensureVisible()
				return m, nil
			}
		}
		previous := m.filter.Value()
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msg)
		if previous != m.filter.Value() {
			m.applyFilter()
		}
		m.ensureVisible()
		return m, cmd
	}
	if msg, ok := msg.(tea.KeyMsg); ok && m.focused {
		switch {
		case key.Matches(msg, m.keys.Filter):
			m.setFiltering(true)
			return m, m.filter.Focus()
		case key.Matches(msg, m.keys.ClearFilter):
			m.filter.SetValue("")
			m.applyFilter()
			return m, nil
		case msg.String() == "esc":
			return m, tea.Interrupt
		case key.Matches(msg, m.keys.Next, m.keys.Submit):
			return m, huh.NextField
		case key.Matches(msg, m.keys.Prev):
			return m, huh.PrevField
		}
		if len(m.filtered) == 0 {
			return m, nil
		}
		node := m.filtered[m.cursor]
		switch {
		case key.Matches(msg, m.keys.Up):
			m.cursor = max(0, m.cursor-1)
		case key.Matches(msg, m.keys.Down):
			m.cursor = min(len(m.filtered)-1, m.cursor+1)
		case key.Matches(msg, m.keys.GotoTop):
			m.cursor = 0
		case key.Matches(msg, m.keys.GotoBottom):
			m.cursor = len(m.filtered) - 1
		case msg.String() == "pgup":
			m.cursor = max(0, m.cursor-m.rows())
		case msg.String() == "pgdown":
			m.cursor = min(len(m.filtered)-1, m.cursor+m.rows())
		case key.Matches(msg, m.keys.HalfPageUp):
			m.cursor = max(0, m.cursor-max(1, m.rows()/2))
		case key.Matches(msg, m.keys.HalfPageDown):
			m.cursor = min(len(m.filtered)-1, m.cursor+max(1, m.rows()/2))
		case key.Matches(msg, m.keys.Toggle):
			node.setSelection(!node.selected)
		}
	}
	m.ensureVisible()
	return m, nil
}

func (m *treeSelect) rows() int {
	height := m.height
	if m.terminalHeight > 0 {
		height = min(height, m.terminalHeight-2) // Reserve the group's help footer.
	}
	if m.filtering || m.filter.Value() != "" {
		height--
	}
	return max(1, height-2) // Reserve title and counter.
}

func (m *treeSelect) ensureVisible() {
	rows := m.displayRows()
	m.top = min(m.top, max(0, len(rows)-m.rows()))
	if len(m.filtered) == 0 {
		return
	}
	for pos, row := range rows {
		if row.node == m.filtered[m.cursor] {
			if pos < m.top {
				m.top = pos
			}
			if pos >= m.top+m.rows() {
				m.top = pos - m.rows() + 1
			}
			break
		}
	}
}

func treeInputText(text string, width int) string {
	text = strings.NewReplacer("\n", " ", "\r", " ", "\t", " ").Replace(text)
	return ansi.Truncate(text, max(0, width), treeTruncationMarker)
}

func (m *treeSelect) View() string {
	styles := m.theme.Focused
	if !m.focused {
		styles = m.theme.Blurred
	}
	width := max(1, m.width-styles.Base.GetHorizontalFrameSize())
	lines := []string{
		styles.Title.Render(treeInputText(m.title, width)),
		styles.Description.Render(treeInputText(fmt.Sprintf(treeSelectionCountFormat, len(m.values()), len(m.nodes)), width)),
	}
	if m.filtering {
		m.filter.Width = max(1, width-ansi.StringWidth(m.filter.Prompt))
		lines = append(lines, ansi.Truncate(m.filter.View(), width, treeTruncationMarker))
	} else if m.filter.Value() != "" {
		lines = append(lines, styles.Description.Render(treeInputText(treeFilterPrompt+m.filter.Value(), width)))
	}
	if len(m.filtered) == 0 {
		lines = append(lines, styles.Description.Render(treeInputText(treeNoMatchesText, width)))
	}
	rows := m.displayRows()
	end := min(len(rows), m.top+m.rows())
	for pos := m.top; pos < end; pos++ {
		row := rows[pos]
		indent := strings.Repeat(" ", row.depth*treeIndentWidth)
		if row.node == nil {
			lines = append(lines, styles.Description.Render(treeInputText("  "+indent+treeOmittedLevels, width)))
			continue
		}
		node := row.node
		cursor := "  "
		if node == m.filtered[m.cursor] {
			cursor = styles.MultiSelectSelector.String()
		}
		style := styles.UnselectedOption
		prefix := styles.UnselectedPrefix.String()
		if node.selected {
			style = styles.SelectedOption
			prefix = styles.SelectedPrefix.String()
		}
		label := treeInputText(node.label, width-ansi.StringWidth(cursor+prefix)-len(indent))
		line := cursor + indent + prefix + style.Render(label)
		lines = append(lines, ansi.Truncate(line, width, treeTruncationMarker))
	}
	return styles.Base.Render(strings.Join(lines, "\n"))
}

// huh.Field integration: focus, layout, help, and accessible prompting.
func (m *treeSelect) Focus() tea.Cmd {
	m.focused = true
	if m.filtering {
		return m.filter.Focus()
	}
	return nil
}

func (m *treeSelect) Blur() tea.Cmd {
	m.focused = false
	m.filter.Blur()
	return nil
}

func (m *treeSelect) Error() error   { return nil }
func (m *treeSelect) Skip() bool     { return len(m.roots) == 0 }
func (m *treeSelect) Zoom() bool     { return false }
func (m *treeSelect) GetKey() string { return "" }
func (m *treeSelect) GetValue() any  { return m.values() }

func (m *treeSelect) KeyBinds() []key.Binding {
	return []key.Binding{
		m.keys.Up, m.keys.Down, m.keys.Toggle,
		m.keys.Filter, m.keys.SetFilter, m.keys.ClearFilter,
		m.keys.Prev, m.keys.Next, m.keys.Submit,
	}
}

func (m *treeSelect) WithTheme(theme *huh.Theme) huh.Field {
	m.theme = theme
	m.filter.PromptStyle = theme.Focused.Description
	m.filter.TextStyle = theme.Focused.TextInput.Text
	m.filter.PlaceholderStyle = theme.Focused.TextInput.Placeholder
	m.filter.Cursor.Style = theme.Focused.TextInput.Cursor
	m.filter.Cursor.TextStyle = theme.Focused.TextInput.CursorText
	return m
}

func (m *treeSelect) WithKeyMap(keys *huh.KeyMap) huh.Field {
	m.keys = keys.MultiSelect
	m.WithPosition(m.position)
	return m
}

func (m *treeSelect) WithAccessible(accessible bool) huh.Field {
	m.accessible = accessible
	return m
}

func (m *treeSelect) WithWidth(width int) huh.Field {
	m.width = max(1, width)
	return m
}

func (m *treeSelect) WithHeight(height int) huh.Field {
	m.height = max(1, height)
	m.ensureVisible()
	return m
}

func (m *treeSelect) WithPosition(position huh.FieldPosition) huh.Field {
	m.position = position
	m.keys.Prev.SetEnabled(!m.filtering && !position.IsFirst())
	m.keys.Next.SetEnabled(!m.filtering && !position.IsLast())
	m.keys.Submit.SetEnabled(!m.filtering && position.IsLast())
	m.keys.Toggle.SetEnabled(!m.filtering)
	m.keys.Filter.SetEnabled(!m.filtering)
	m.keys.SetFilter.SetEnabled(m.filtering)
	m.keys.ClearFilter.SetEnabled(!m.filtering && m.filter.Value() != "")
	return m
}

func (m *treeSelect) Run() error {
	return huh.NewForm(huh.NewGroup(m)).WithAccessible(m.accessible).Run()
}

func (m *treeSelect) RunAccessible(w io.Writer, r io.Reader) error {
	// Numbered choices are only prompt controls; selected labels are returned.
	scanner := bufio.NewScanner(r)
	for {
		var prompt strings.Builder
		fmt.Fprintf(&prompt, "%s\n"+treeSelectionCountFormat+"\n", m.title, len(m.values()), len(m.nodes))
		if m.filter.Value() != "" {
			fmt.Fprintf(&prompt, "Filter: %s\n", m.filter.Value())
		}
		if len(m.filtered) == 0 {
			prompt.WriteString(treeNoMatchesText + "\n")
		}
		for i, node := range m.filtered {
			checkbox := "[ ]"
			if node.selected {
				checkbox = "[•]"
			}
			fmt.Fprintf(&prompt, "%d. %s%s %s\n", i+1,
				strings.Repeat(" ", node.depth*treeIndentWidth), checkbox, node.label)
		}
		prompt.WriteString("Enter a number to toggle a subtree, /text to filter (/ to clear), 0 to confirm, or q to cancel: ")
		if _, err := io.WriteString(w, prompt.String()); err != nil {
			return err
		}
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return err
			}
			return io.EOF
		}
		input := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(input, treeFilterPrompt) {
			m.filter.SetValue(strings.TrimPrefix(input, treeFilterPrompt))
			m.applyFilter()
			continue
		}
		if input == "q" {
			return huh.ErrUserAborted
		}
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 0 || choice > len(m.filtered) {
			continue
		}
		if choice == 0 {
			return nil
		}
		node := m.filtered[choice-1]
		node.setSelection(!node.selected)
	}
}
