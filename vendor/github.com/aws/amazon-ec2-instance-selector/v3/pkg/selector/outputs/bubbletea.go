// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package outputs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/instancetypes"
	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/sorter"
)

const (
	// can't get terminal dimensions on startup, so use this.
	initialDimensionVal = 30

	instanceTypeKey = "instance type"
	selectedKey     = "selected"
)

const (
	// table states.
	stateTable   = "table"
	stateVerbose = "verbose"
	stateSorting = "sorting"
)

var controlsStyle = lipgloss.NewStyle().Faint(true)

// BubbleTeaModel is used to hold the state of the bubble tea TUI.
type BubbleTeaModel struct {
	// holds the output currentState of the model
	currentState string

	// the model for the table view
	tableModel tableModel

	// holds state for the verbose view
	verboseModel verboseModel

	// holds the state for the sorting view
	sortingModel sortingModel
}

// NewBubbleTeaModel initializes a new bubble tea Model which represents
// a stylized table to display instance types.
func NewBubbleTeaModel(instanceTypes []*instancetypes.Details) BubbleTeaModel {
	return BubbleTeaModel{
		currentState: stateTable,
		tableModel:   *initTableModel(instanceTypes),
		verboseModel: *initVerboseModel(),
		sortingModel: *initSortingModel(instanceTypes),
	}
}

// Init is used by bubble tea to initialize a bubble tea table.
func (m BubbleTeaModel) Init() tea.Cmd {
	return nil
}

// Update is used by bubble tea to update the state of the bubble
// tea model based on user input.
func (m BubbleTeaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// don't listen for input if currently typing into text field
		if m.tableModel.filterTextInput.Focused() {
			break
		} else if m.sortingModel.sortTextInput.Focused() {
			// see if we should sort and switch states to table
			if m.currentState == stateSorting && msg.String() == "enter" {
				jsonPath := m.sortingModel.sortTextInput.Value()

				sortDirection := sorter.SortAscending
				if m.sortingModel.isDescending {
					sortDirection = sorter.SortDescending
				}

				var err error
				m.tableModel, err = m.tableModel.sortTable(jsonPath, sortDirection)
				if err != nil {
					m.sortingModel.sortTextInput.SetValue(jsonPathError)
					break
				}

				m.currentState = stateTable

				m.sortingModel.sortTextInput.Blur()
			}

			break
		}

		// check for quit or change in state
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "e":
			// switch from table state to verbose state
			if m.currentState == stateTable {
				// get focused instance type
				focusedRow := m.tableModel.table.HighlightedRow()
				focusedInstance, ok := focusedRow.Data[instanceTypeKey].(*instancetypes.Details)
				if !ok {
					break
				}

				// set content of view
				m.verboseModel.focusedInstanceName = focusedInstance.InstanceType
				m.verboseModel.viewport.SetContent(VerboseInstanceTypeOutput([]*instancetypes.Details{focusedInstance})[0])

				// move viewport to top of printout
				m.verboseModel.viewport.SetYOffset(0)

				// switch from table state to verbose state
				m.currentState = stateVerbose
			}
		case "s":
			// switch from table view to sorting view
			if m.currentState == stateTable {
				m.currentState = stateSorting
			}
		case "enter":
			// sort and switch states to table
			if m.currentState == stateSorting {
				sortFilter := string(m.sortingModel.shorthandList.SelectedItem().(item))

				sortDirection := sorter.SortAscending
				if m.sortingModel.isDescending {
					sortDirection = sorter.SortDescending
				}

				var err error
				m.tableModel, err = m.tableModel.sortTable(sortFilter, sortDirection)
				if err != nil {
					m.sortingModel.sortTextInput.SetValue("INVALID SHORTHAND VALUE")
					break
				}

				m.currentState = stateTable

				m.sortingModel.sortTextInput.Blur()
			}
		case "esc":
			// switch from sorting state or verbose state to table state
			if m.currentState == stateSorting || m.currentState == stateVerbose {
				m.currentState = stateTable
			}
		}
	case tea.WindowSizeMsg:
		// This is needed to handle a bug with bubble tea
		// where resizing causes misprints (https://github.com/Evertras/bubble-table/issues/121)
		termenv.ClearScreen() //nolint:staticcheck

		// handle screen resizing
		m.tableModel = m.tableModel.resizeView(msg)
		m.verboseModel = m.verboseModel.resizeView(msg)
		m.sortingModel = m.sortingModel.resizeView(msg)
	}

	var cmd tea.Cmd
	// update currently active state
	switch m.currentState {
	case stateTable:
		m.tableModel, cmd = m.tableModel.update(msg)
	case stateVerbose:
		m.verboseModel, cmd = m.verboseModel.update(msg)
	case stateSorting:
		m.sortingModel, cmd = m.sortingModel.update(msg)
	}

	return m, cmd
}

// View is used by bubble tea to render the bubble tea model.
func (m BubbleTeaModel) View() string {
	switch m.currentState {
	case stateTable:
		return m.tableModel.view()
	case stateVerbose:
		return m.verboseModel.view()
	case stateSorting:
		return m.sortingModel.view()
	}

	return ""
}
