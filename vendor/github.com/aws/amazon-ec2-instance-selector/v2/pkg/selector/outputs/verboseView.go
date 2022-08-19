// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package outputs

import (
	"fmt"
	"math"
	"strings"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	// verbose view formatting
	outlinePadding = 8

	// controls
	verboseControls = "Controls: ↑/↓ - up/down • esc - return to table • q - quit"
)

// verboseModel represents the current state of the verbose view
type verboseModel struct {
	// model for verbose output viewport
	viewport viewport.Model

	// the instance which the verbose output is focused on
	focusedInstanceName *string
}

// styling for viewport
var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()
)

// initVerboseModel initializes and returns a new verboseModel based on the given
// instance type details
func initVerboseModel(instanceTypes []*instancetypes.Details) *verboseModel {
	viewportModel := viewport.New(initialDimensionVal, initialDimensionVal)
	viewportModel.MouseWheelEnabled = true

	return &verboseModel{
		viewport: viewportModel,
	}
}

// resizeView will change the dimensions of the verbose viewport in order to accommodate
// the new window dimensions represented by the given tea.WindowSizeMsg
func (m verboseModel) resizeView(msg tea.WindowSizeMsg) verboseModel {
	// handle width changes
	m.viewport.Width = msg.Width

	// handle height changes
	if outlinePadding >= msg.Height {
		// height too short to fit viewport
		m.viewport.Height = 0
	} else {
		newHeight := msg.Height - outlinePadding
		m.viewport.Height = newHeight
	}

	return m
}

// update updates the state of the verboseModel
func (m verboseModel) update(msg tea.Msg) (verboseModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m verboseModel) view() string {
	outputStr := strings.Builder{}

	// format header for viewport
	instanceName := titleStyle.Render(*m.focusedInstanceName)
	line := strings.Repeat("─", int(math.Max(0, float64(m.viewport.Width-lipgloss.Width(instanceName)))))
	outputStr.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, instanceName, line))
	outputStr.WriteString("\n")

	outputStr.WriteString(m.viewport.View())
	outputStr.WriteString("\n")

	// format footer for viewport
	pagePercentage := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line = strings.Repeat("─", int(math.Max(0, float64(m.viewport.Width-lipgloss.Width(pagePercentage)))))
	outputStr.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, line, pagePercentage))
	outputStr.WriteString("\n")

	// controls
	outputStr.WriteString(controlsStyle.Render(verboseControls))
	outputStr.WriteString("\n")

	return outputStr.String()
}
