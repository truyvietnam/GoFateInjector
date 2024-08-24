package injector

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	ErrNoProcess                    = errors.New("Can't find process!")
	ErrProcess32Fail                = errors.New("Process32Next failed")
	ErrCreateToolhelp32SnapshotFail = errors.New("CreateToolhelp32Snapshot failed")
	FoundProcess                    = errors.New("Process found!") //not error
	InvalidPath                     = errors.New("invalid file path")
	Injected                        = errors.New("valid file path | Injected!")

	kernel32Dll        = windows.NewLazyDLL("kernel32.dll")
	virtualAllocEx     = kernel32Dll.NewProc("VirtualAllocEx")
	createRemoteThread = kernel32Dll.NewProc("CreateRemoteThread")
)

func FindProcessByName(name string) (int, error) {
	ProcList := syscall.ProcessEntry32{}
	ProcList.Size = uint32(unsafe.Sizeof(ProcList))

	handleProcList, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, ErrCreateToolhelp32SnapshotFail
	}

	err = syscall.Process32Next(handleProcList, &ProcList)
	if err != nil {
		return 0, ErrProcess32Fail
	}

	for {
		err := syscall.Process32Next(handleProcList, &ProcList)
		if err != nil {
			return 0, ErrNoProcess
		}

		currName := windows.UTF16PtrToString(&ProcList.ExeFile[0])
		if name == strings.ToLower(currName) {
			return int(ProcList.ProcessID), FoundProcess
		}
	}
}

func Inject(procId int, path string) error {
	procHandler, err := windows.OpenProcess(windows.PROCESS_CREATE_PROCESS|windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_OPERATION|windows.PROCESS_VM_WRITE|windows.PROCESS_VM_READ, false, uint32(procId))
	if err != nil {
		log.Println(err)
	}
	loc, _, _ := virtualAllocEx.Call(uintptr(procHandler), 0, windows.MAX_PATH, windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_READWRITE)

	if _, err := os.Stat(path); os.IsNotExist(err) || strings.ToLower(filepath.Ext(path)) != ".dll" {
		return InvalidPath
	}

	ptrPath, err := windows.BytePtrFromString(path)
	if err != nil {
		log.Println(err)
	}

	zero := uintptr(0)
	err = windows.WriteProcessMemory(procHandler, loc, ptrPath, uintptr(len(path)+1), &zero)
	if err != nil {
		log.Println(err)
	}

	loadLibraryAddr, err := syscall.GetProcAddress(syscall.Handle(kernel32Dll.Handle()), "LoadLibraryA")
	if err != nil {
		log.Println(err)
	}

	remoteThread, _, _ := createRemoteThread.Call(uintptr(procHandler), 0, 0, loadLibraryAddr, loc, 0, 0)
	defer windows.CloseHandle(windows.Handle(remoteThread))

	return Injected
}
