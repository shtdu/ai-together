// Copyright (c) 2025 AI Together
//
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


package model

import (
	"github.com/gdamore/tcell/v2"
)

// Styles contains all tview styling configuration
type Styles struct {
	// Border color (#7D56F4 purple)
	BorderColor tcell.Color

	// Highlight color (#EE6FF8 pink)
	HighlightColor tcell.Color

	// Status bar background (#3E387D dark purple)
	StatusBarBg tcell.Color

	// Status bar text (#FAFAFA off-white)
	StatusBarText tcell.Color

	// Dim text (#6B6B8A gray)
	DimText tcell.Color

	// Normal text (#FAFAFA off-white)
	NormalText tcell.Color

	// Header color (#7D56F4 purple)
	HeaderColor tcell.Color
}

// DefaultStyles returns the default color scheme matching the bubbletea version
func DefaultStyles() *Styles {
	return &Styles{
		BorderColor:    tcell.GetColor("#7D56F4"), // #7D56F4 purple
		HighlightColor: tcell.GetColor("#EE6FF8"), // #EE6FF8 pink
		StatusBarBg:    tcell.GetColor("#3E387D"), // #3E387D dark purple
		StatusBarText:  tcell.GetColor("#FAFAFA"), // #FAFAFA off-white
		DimText:        tcell.GetColor("#6B6B8A"), // #6B6B8A gray
		NormalText:     tcell.GetColor("#FAFAFA"), // #FAFAFA off-white
		HeaderColor:    tcell.GetColor("#7D56F4"), // #7D56F4 purple
	}
}
