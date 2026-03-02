// Copyright (c) 2025 Code Together
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
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// NotificationSeverity represents the severity level of a notification
type NotificationSeverity int

const (
	SeverityInfo NotificationSeverity = iota
	SeveritySuccess
	SeverityWarning
	SeverityError
)

// ShowNotification displays a toast notification with the given message and severity
func (a *App) ShowNotification(message string, severity NotificationSeverity) {
	// Stop any existing notification timer
	if a.NotificationTimer != nil {
		a.NotificationTimer.Stop()
	}

	// Create or get notification view
	if a.NotificationView == nil {
		a.NotificationView = tview.NewTextView()
		a.NotificationView.SetDynamicColors(true)
		a.NotificationView.SetTextAlign(tview.AlignCenter)
	}

	// Set color based on severity
	var bgColor tcell.Color
	var icon string

	switch severity {
	case SeveritySuccess:
		bgColor = tcell.GetColor("#98C379") // Green
		icon = "✓"
	case SeverityWarning:
		bgColor = tcell.GetColor("#E5C07B") // Yellow
		icon = "⚠"
	case SeverityError:
		bgColor = tcell.GetColor("#E06C75") // Red
		icon = "✕"
	default: // Info
		bgColor = tcell.GetColor("#61AFEF") // Blue
		icon = "ℹ"
	}

	// Format the notification
	text := fmt.Sprintf("[%s::b]%s [::-]%s", colorToHex(bgColor), icon, message)
	a.NotificationView.SetText(text)
	a.NotificationView.SetBackgroundColor(bgColor)
	a.NotificationView.SetTextColor(tcell.ColorBlack)

	// Add notification to the top of the main layout
	// For simplicity, we'll just update status bar with the notification
	a.StatusBar.SetText(fmt.Sprintf("[%s::b]%s %s[-]", colorToHex(bgColor), icon, message))
	a.StatusBar.SetBackgroundColor(bgColor)

	// Mark notification as active
	a.NotificationActive = true

	// Auto-dismiss after 5 seconds
	a.NotificationTimer = time.AfterFunc(5*time.Second, func() {
		a.Application.QueueUpdateDraw(func() {
			a.DismissNotification()
		})
	})
}

// DismissNotification removes the current notification
func (a *App) DismissNotification() {
	if !a.NotificationActive {
		return
	}

	// Reset status bar to normal
	a.StatusBar.SetBackgroundColor(a.Styles.StatusBarBg)
	a.StatusBar.SetTextColor(a.Styles.StatusBarText)
	a.UpdateStatusBar()

	a.NotificationActive = false

	if a.NotificationTimer != nil {
		a.NotificationTimer.Stop()
		a.NotificationTimer = nil
	}
}
