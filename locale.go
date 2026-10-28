package main

import "github.com/egoist/mygo"

func localeTag(language string) string {
	if language == "en" {
		return "en"
	}
	return "zh-Hans"
}

func nativeText(language, english, chinese string) string {
	if language == "en" {
		return english
	}
	return chinese
}

func (d *Desktop) trayOptions(language string) mygo.TrayOptions {
	return mygo.TrayOptions{Icon: trayIcon, IconIsTemplate: true, Menu: d.trayMenu(language)}
}

func (d *Desktop) trayMenu(language string) *mygo.Menu {
	return mygo.NewMenu([]*mygo.MenuItem{
		{Label: nativeText(language, "Explain word under pointer", "解释鼠标下的单词"), Click: func(*mygo.MenuItem, *mygo.Window) { d.startLookup() }},
		{Label: nativeText(language, "Settings…", "设置…"), Click: func(*mygo.MenuItem, *mygo.Window) { d.OpenSettings() }},
		mygo.Separator(),
		{Label: nativeText(language, "Quit InlineExplain", "退出 InlineExplain"), Click: func(*mygo.MenuItem, *mygo.Window) { mygo.App.Quit() }},
	})
}

func (d *Desktop) applyLocale(language string) {
	mygo.App.SetLocale(localeTag(language))
	if d.preferences != nil {
		d.preferences.SetTitle(nativeText(language, "InlineExplain Settings", "InlineExplain 设置"))
	}
	if d.tray != nil {
		d.tray.SetMenu(d.trayMenu(language))
		d.tray.SetToolTip(nativeText(language, "InlineExplain · Point at a word and press Option+S", "InlineExplain · 鼠标指向单词，按 Option+S 取词"))
	}
}
