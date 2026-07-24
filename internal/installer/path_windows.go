//go:build windows

package installer

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

// addToPath prepends/appends dir to the current user's persistent PATH
// (HKCU\Environment) and notifies running processes of the change.
func addToPath(dir string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("opening HKCU\\Environment: %w", err)
	}
	defer k.Close()

	existing, _, err := k.GetStringValue("Path")
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("reading current user Path: %w", err)
	}

	for _, p := range strings.Split(existing, ";") {
		if p == "" {
			continue
		}
		if strings.EqualFold(strings.TrimRight(p, `\`), strings.TrimRight(dir, `\`)) {
			return nil // already present
		}
	}

	newPath := existing
	if newPath != "" && !strings.HasSuffix(newPath, ";") {
		newPath += ";"
	}
	newPath += dir

	if err := k.SetExpandStringValue("Path", newPath); err != nil {
		return fmt.Errorf("writing user Path: %w", err)
	}

	broadcastEnvChange()
	return nil
}

// broadcastEnvChange notifies other running processes (e.g. Explorer) that
// the environment changed, so newly launched programs pick up the new PATH
// without a logoff/logon. Best-effort only.
func broadcastEnvChange() {
	const (
		hwndBroadcast    = 0xffff
		wmSettingChange  = 0x001A
		smtoAbortIfHung  = 0x0002
		broadcastTimeout = 5000
	)
	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessageTimeout := user32.NewProc("SendMessageTimeoutW")
	env, err := syscall.UTF16PtrFromString("Environment")
	if err != nil {
		return
	}
	sendMessageTimeout.Call(
		uintptr(hwndBroadcast),
		uintptr(wmSettingChange),
		0,
		uintptr(unsafe.Pointer(env)),
		uintptr(smtoAbortIfHung),
		uintptr(broadcastTimeout),
		0,
	)
}
