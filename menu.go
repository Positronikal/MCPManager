package main

import (
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// createApplicationMenu builds the native OS menu bar.
// Callbacks close over a.ctx, which is populated by startup() before any
// menu item can be clicked.
func (a *App) createApplicationMenu() *menu.Menu {
	appMenu := menu.NewMenu()

	// ── File ──────────────────────────────────────────────────────────────
	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Refresh Discovery", keys.Key("f5"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:action", "refresh-discovery")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Exit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		runtime.Quit(a.ctx)
	})

	// ── Edit (built-in role: Cut / Copy / Paste / Select All) ─────────────
	appMenu.Append(menu.EditMenu())

	// ── View ──────────────────────────────────────────────────────────────
	viewMenu := appMenu.AddSubmenu("View")
	viewMenu.AddText("Servers", keys.CmdOrCtrl("1"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:action", "view-servers")
	})
	viewMenu.AddText("Netstat", keys.CmdOrCtrl("2"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:action", "view-netstat")
	})
	viewMenu.AddText("Shell", keys.CmdOrCtrl("3"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:action", "view-shell")
	})
	viewMenu.AddText("Explorer", keys.CmdOrCtrl("4"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:action", "view-explorer")
	})
	viewMenu.AddSeparator()
	viewMenu.AddText("Help View", keys.CmdOrCtrl("5"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:action", "view-help")
	})
	viewMenu.AddText("Toggle Log Panel", keys.CmdOrCtrl("l"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:action", "toggle-log-panel")
	})

	// ── Server ────────────────────────────────────────────────────────────
	serverMenu := appMenu.AddSubmenu("Server")
	serverMenu.AddText("Start Server", keys.Combo("s", keys.CmdOrCtrlKey, keys.ShiftKey), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:action", "server-start")
	})
	serverMenu.AddText("Stop Server", keys.Combo("x", keys.CmdOrCtrlKey, keys.ShiftKey), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:action", "server-stop")
	})
	serverMenu.AddText("Restart Server", keys.Combo("r", keys.CmdOrCtrlKey, keys.ShiftKey), func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:action", "server-restart")
	})
	serverMenu.AddSeparator()
	serverMenu.AddText("View Configuration", nil, func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:action", "server-view-config")
	})
	serverMenu.AddText("Open Server Directory", nil, func(_ *menu.CallbackData) {
		runtime.EventsEmit(a.ctx, "menu:action", "server-open-directory")
	})

	// ── Window ────────────────────────────────────────────────────────────
	windowMenu := appMenu.AddSubmenu("Window")
	windowMenu.AddText("Minimize", nil, func(_ *menu.CallbackData) {
		runtime.WindowMinimise(a.ctx)
	})
	windowMenu.AddText("Maximize/Restore", nil, func(_ *menu.CallbackData) {
		runtime.WindowToggleMaximise(a.ctx)
	})
	windowMenu.AddSeparator()
	windowMenu.AddText("Full Screen", keys.Key("f11"), func(_ *menu.CallbackData) {
		runtime.WindowFullscreen(a.ctx)
	})

	// ── Help ──────────────────────────────────────────────────────────────
	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("MCP Documentation", nil, func(_ *menu.CallbackData) {
		runtime.BrowserOpenURL(a.ctx, "https://modelcontextprotocol.io/introduction")
	})
	helpMenu.AddText("View on GitHub", nil, func(_ *menu.CallbackData) {
		runtime.BrowserOpenURL(a.ctx, "https://github.com/Positronikal/MCPManager")
	})
	helpMenu.AddSeparator()
	helpMenu.AddText("About MCP Manager", nil, func(_ *menu.CallbackData) {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:  runtime.InfoDialog,
			Title: "About MCP Manager",
			Message: fmt.Sprintf(
				"MCP Manager v%s\n\nA desktop application for managing Model Context Protocol servers.\n\n© 2025 MCP Manager Contributors\nLicensed under the GNU General Public License v3.0",
				appVersion,
			),
		})
	})

	return appMenu
}
