package ui

import (
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

func newKeyPress(text string) tea.KeyPressMsg {
	if text == "" {
		return tea.KeyPressMsg{}
	}
	r, _ := utf8.DecodeRuneInString(text)
	return tea.KeyPressMsg{Code: r, Text: text}
}

func newSpecialKey(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}

