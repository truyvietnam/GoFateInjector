package main

import (
	"GoFateInjector/injector"
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
	"github.com/rodrigocfd/windigo/win/com/com"
	"github.com/rodrigocfd/windigo/win/com/com/comco"
	"github.com/rodrigocfd/windigo/win/com/shell"
	"github.com/rodrigocfd/windigo/win/com/shell/shellco"
)

var (
	InjectPopupId  = 1
	OpenPopupId    = 2
	ClosePopupId   = 3
	WMUserTrayIcon = co.WM_USER + 1
)

func main() {
	runtime.LockOSThread()

	myWindow := NewMyWindow() // instantiate
	myWindow.wnd.RunAsMain()  // ...and run
}

// This struct represents our main window.
type MyWindow struct {
	wnd         ui.WindowMain
	statusBar   ui.StatusBar
	procName    ui.Edit
	dllPathName ui.Edit
	injectBtn   ui.Button
	selectBtn   ui.Button
	hideBtn     ui.Button
	customCheck ui.CheckBox
	autoCheck   ui.CheckBox
	delayTxt    ui.Edit
}

func ReadAndApplyConfig(windows *MyWindow) {
	var procName, dllPath, delay string
	var customProc bool

	f, e := os.Open("config.txt")
	if e != nil {
		panic(e)
	}
	defer f.Close()
	s := bufio.NewScanner(f)

	for s.Scan() {
		fmt.Sscanf(s.Text(), "customProcName=%t", &customProc)
		fmt.Sscanf(s.Text(), "delaystr=%s", &delay)
		fmt.Sscanf(s.Text(), "procName=%s", &procName)

		if strings.Contains(s.Text(), "dllPath=") { //because double quote so I have to do this homeless method
			dllPath = s.Text()
		}
	}

	windows.procName.SetText(procName)
	windows.delayTxt.SetText(delay)
	windows.dllPathName.SetText(strings.TrimLeft(dllPath, "dllPath="))
	if customProc {
		windows.customCheck.SetCheckStateAndTrigger(co.BST_CHECKED)
	}
}

func WriteConfig(windows *MyWindow) {
	fileName := "config.txt"

	file, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	delay, _ := strconv.Atoi(windows.delayTxt.Text())
	data := fmt.Sprintf("#Go Fate Client injector config file\ncustomProcName=%t\ndelaystr=%d\ndllPath=%s\nprocName=%s", windows.customCheck.IsChecked(), delay, windows.dllPathName.Text(), windows.procName.Text())

	_, err = writer.WriteString(data)
	if err != nil {
		return
	}

	err = writer.Flush()
	if err != nil {
		return
	}
}

func GetChosenFilePath(owner win.HWND) string {
	var path string

	fod := shell.NewIFileOpenDialog(
		com.CoCreateInstance(
			shellco.CLSID_FileOpenDialog, nil,
			comco.CLSCTX_INPROC_SERVER,
			shellco.IID_IFileOpenDialog),
	)
	defer fod.Release()
	fod.SetFileTypes([]shell.FilterSpec{
		{Name: "Dynamic link library", Spec: "*.dll"},
	})
	fod.SetTitle("Select the .dll file")
	if fod.Show(owner) {
		chosenFile := fod.GetResult()
		path = chosenFile.GetDisplayName(shellco.SIGDN_FILESYSPATH)
	}
	return path
}

func SysTrayIcon(hwnd win.HWND) {
	nid := win.NOTIFYICONDATA{}
	nid.SetCbSize()

	nid.Hwnd = hwnd // the owner HWND goes here...
	nid.UID = 20    // any ID number
	nid.UFlags = co.NIF_ICON | co.NIF_MESSAGE | co.NIF_TIP
	nid.HIcon = hwnd.Hinstance().LoadIcon(win.IconResInt(2))
	nid.UCallbackMessage = WMUserTrayIcon
	nid.SetSzTip("Double-Click to show Go Fate Injector")
	win.ShellNotifyIcon(co.NIM_ADD, &nid)
}

func Inject(windows *MyWindow) {
	procId, err := injector.FindProcessByName(windows.procName.Text())
	if err != nil && err != injector.FoundProcess {
		windows.statusBar.Parts().SetAllTexts(fmt.Sprintf("%s | %d", err.Error(), procId)) //SetTextAndResize(fmt.Sprintf("%s | %d", err.Error(), procId))
		return
	}
	err2 := injector.Inject(procId, windows.dllPathName.Text())
	if err2 != nil {
		windows.statusBar.Hwnd().SetWindowPos(win.HWND(0), 0, 0,
			20, 100, co.SWP_NOZORDER|co.SWP_NOMOVE) //sketchy

		windows.statusBar.Parts().SetAllTexts(fmt.Sprintf("%s | %d | %s", err.Error(), procId, err2.Error())) //SetTextAndResize(fmt.Sprintf("%s | %d | %s", err.Error(), procId, err2.Error()))

		if err2 == injector.Injected {
			WriteConfig(windows)
		}
	}
}

func AutoInject(windows *MyWindow, delay int) {
	windows.statusBar.Hwnd().SetWindowPos(win.HWND(0), 0, 0,
		20, 100, co.SWP_NOZORDER|co.SWP_NOMOVE) //sketchy (resize again)
	if delay != 1 {
		windows.statusBar.Parts().SetAllTexts(fmt.Sprintf("AutoInject: Enabled | trying every %d seconds", delay)) //SetTextAndResize(fmt.Sprintf("AutoInject: Enabled | trying every %d seconds", delay))
	} else if delay <= 1 {
		windows.delayTxt.SetText("1")
		delay = 1
		windows.statusBar.Parts().SetAllTexts("AutoInject: Enabled | trying every second") //SetTextAndResize("AutoInject: Enabled | trying every second")
	}

	for {
		if !windows.autoCheck.IsChecked() {
			break
		}
		time.Sleep(time.Duration(delay) * time.Second)
		Inject(windows)
	}
}

// Creates a new instance of our main window.
func NewMyWindow() *MyWindow {
	wnd := ui.NewWindowMain(
		ui.WindowMainOpts().
			Title("Go Fate Injector").
			ClientArea(win.SIZE{Cx: 281, Cy: 124}).
			IconId(2),
	)

	com.CoInitializeEx(comco.COINIT_APARTMENTTHREADED)

	me := &MyWindow{
		wnd:       wnd,
		statusBar: ui.NewStatusBar(wnd),
		procName: ui.NewEdit(wnd,
			ui.EditOpts().
				Position(win.POINT{X: 110, Y: 5}).
				Size(win.SIZE{Cx: 165, Cy: 20}).
				Text("minecraft.windows.exe"),
		),
		injectBtn: ui.NewButton(wnd,
			ui.ButtonOpts().
				Text("Inject").
				Position(win.POINT{X: 5, Y: 5}).
				Size(win.SIZE{Cx: 100, Cy: 40}),
		),
		selectBtn: ui.NewButton(wnd,
			ui.ButtonOpts().
				Text("Select").
				Position(win.POINT{X: 5, Y: 75}).
				Size(win.SIZE{Cx: 60, Cy: 20}),
		),
		dllPathName: ui.NewEdit(wnd,
			ui.EditOpts().
				Position(win.POINT{X: 70, Y: 75}).
				Size(win.SIZE{Cx: 205, Cy: 20}).
				Text("Click \"Select\" to select the dll file"),
		),
		hideBtn: ui.NewButton(wnd,
			ui.ButtonOpts().
				Text("Hide Menu").
				Position(win.POINT{X: 5, Y: 50}).
				Size(win.SIZE{Cx: 100, Cy: 20}),
		),
		customCheck: ui.NewCheckBox(wnd,
			ui.CheckBoxOpts().
				Text("Custom Target").
				Position(win.POINT{X: 110, Y: 30}).
				Size(win.SIZE{Cx: 165, Cy: 20}),
		),
		autoCheck: ui.NewCheckBox(wnd,
			ui.CheckBoxOpts().
				Text("Auto Inject").
				Position(win.POINT{X: 110, Y: 50}).
				Size(win.SIZE{Cx: 130, Cy: 20}),
		),
		delayTxt: ui.NewEdit(wnd,
			ui.EditOpts().
				Text("5").
				Position(win.POINT{X: 245, Y: 50}).
				Size(win.SIZE{Cx: 30, Cy: 20}).
				CtrlStyles(co.ES_CENTER|co.ES_NUMBER),
		),
	}

	me.customCheck.On().BnClicked(func() {
		if !me.customCheck.IsChecked() {
			me.procName.SetText("minecraft.windows.exe")
			me.procName.Hwnd().EnableWindow(false)
		} else {
			me.procName.Hwnd().EnableWindow(true)
		}
	})

	me.autoCheck.On().BnClicked(func() {
		if me.autoCheck.IsChecked() {
			me.procName.Hwnd().EnableWindow(false)
			me.dllPathName.Hwnd().EnableWindow(false)
			me.delayTxt.Hwnd().EnableWindow(false)
			me.selectBtn.Hwnd().EnableWindow(false)
			me.customCheck.Hwnd().EnableWindow(false)
			delay, _ := strconv.Atoi(me.delayTxt.Text())
			go AutoInject(me, delay)
		} else {
			me.procName.Hwnd().EnableWindow(true)
			me.dllPathName.Hwnd().EnableWindow(true)
			me.delayTxt.Hwnd().EnableWindow(true)
			me.selectBtn.Hwnd().EnableWindow(true)
			me.customCheck.Hwnd().EnableWindow(true)
		}
	})

	me.injectBtn.On().BnClicked(func() {
		Inject(me)
	})

	me.selectBtn.On().BnClicked(func() {
		s := GetChosenFilePath(wnd.Hwnd())
		if s != "" {
			me.dllPathName.SetText(s)
		}
	})

	me.hideBtn.On().BnClicked(func() {
		wnd.Hwnd().ShowWindow(co.SW_HIDE)
	})

	wnd.On().WmRButtonDown(func(p wm.Mouse) {
		mousepos := win.GetCursorPos()
		wnd.Hwnd().ScreenToClientPt(&mousepos)

		popup := win.CreatePopupMenu()
		defer popup.DestroyMenu()
		popup.AddItem(InjectPopupId, "Inject")
		popup.AddItem(OpenPopupId, "Open")
		popup.AddItem(ClosePopupId, "Close")

		popup.ShowAtPoint(mousepos, wnd.Hwnd(), 0)
	})

	wnd.On().AddUserCustom(WMUserTrayIcon, co.CMD(co.WM_RBUTTONDOWN), func(p wm.Any) {
		mousepos := win.GetCursorPos()
		wnd.Hwnd().ScreenToClientPt(&mousepos)

		popup := win.CreatePopupMenu()
		defer popup.DestroyMenu()
		popup.AddItem(InjectPopupId, "Inject")
		popup.AddItem(OpenPopupId, "Open")
		popup.AddItem(ClosePopupId, "Close")

		popup.ShowAtPoint(mousepos, wnd.Hwnd(), 0)
	})

	wnd.On().AddUserCustom(WMUserTrayIcon, co.CMD(co.WM_LBUTTONDBLCLK), func(p wm.Any) {
		wnd.Hwnd().ShowWindow(co.SW_SHOW)
	})

	wnd.On().WmCommandMenu(InjectPopupId, func(p wm.Command) {
		Inject(me)
	})

	wnd.On().WmCommandMenu(OpenPopupId, func(p wm.Command) {
		wnd.Hwnd().ShowWindow(co.SW_SHOW)
	})

	wnd.On().WmCommandMenu(ClosePopupId, func(p wm.Command) {
		wnd.Hwnd().DestroyWindow()
	})

	wnd.On().WmCreate(func(p wm.Create) int {
		SysTrayIcon(wnd.Hwnd())
		me.procName.Hwnd().EnableWindow(false)
		me.customCheck.SetCheckStateAndTrigger(co.BST_UNCHECKED)

		if _, err := os.Stat("config.txt"); os.IsNotExist(err) {
			WriteConfig(me)
		}
		ReadAndApplyConfig(me)
		me.statusBar.Parts().SetAllTexts("Version 1.1 | Made by bigrat.monster")

		return 0
	})

	wnd.On().WmCtlColorStatic(func(p wm.CtlColor) win.HBRUSH {
		return win.CreateSysColorBrush(co.COLOR_WINDOW)
	})

	wnd.On().WmCtlColorBtn(func(p wm.CtlColor) win.HBRUSH {
		return win.CreateSysColorBrush(co.COLOR_WINDOW)
	})

	return me
}
