package ui

import tea "charm.land/bubbletea/v2"

func isCtrlKey(msg tea.KeyPressMsg, key rune) bool {
	return msg.Code == key && msg.Mod.Contains(tea.ModCtrl)
}
