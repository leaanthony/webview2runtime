// +build windows

package webview2runtime

import (
	"golang.org/x/sys/windows/registry"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

// Info contains all the information about an installation of the webview2 runtime.
type Info struct {
	Location        string
	Name            string
	Version         string
	SilentUninstall string
}

// IsOlderThan returns true if the installed version is older than the given required version.
// Returns error if something goes wrong.
func (i *Info) IsOlderThan(requiredVersion string) (bool, error) {
	var mod = syscall.NewLazyDLL("WebView2Loader.dll")
	var CompareBrowserVersions = mod.NewProc("CompareBrowserVersions")
	v1, err := syscall.UTF16PtrFromString(i.Version)
	if err != nil {
		return false, err
	}
	v2, err := syscall.UTF16PtrFromString(requiredVersion)
	if err != nil {
		return false, err
	}
	var result int = 9
	_, _, err = CompareBrowserVersions.Call(uintptr(unsafe.Pointer(v1)), uintptr(unsafe.Pointer(v2)), uintptr(unsafe.Pointer(&result)))
	if result < -1 || result > 1 {
		return false, err
	}
	return result == -1, nil
}

// GetInstalledVersion returns the installed version of the webview2 runtime.
// If there is no version installed, a blank string is returned.
func GetInstalledVersion() *Info {
	var regkey = `SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`
	if runtime.GOARCH == "386" {
		regkey = `SOFTWARE\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`
	}

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, regkey, registry.QUERY_VALUE)
	if err != nil {
		// Cannot open key = not installed
		return nil
	}

	info := &Info{}
	info.Location = getKeyValue(k, "location")
	info.Name = getKeyValue(k, "name")
	info.Version = getKeyValue(k, "pv")
	info.SilentUninstall = getKeyValue(k, "SilentUninstall")

	return info
}

func getKeyValue(k registry.Key, name string) string {
	result, _, _ := k.GetStringValue(name)
	return result
}

// InstallUsingBootstrapper will download the bootstrapper from Microsoft and run it to install
// the latest version of the runtime.
// Returns true if the installer ran successfully.
// Returns an error if something goes wrong
func InstallUsingBootstrapper() (result bool, err error) {
	bootstrapperURL := `https://go.microsoft.com/fwlink/p/?LinkId=2124703`
	installer := filepath.Join(os.TempDir(), `MicrosoftEdgeWebview2Setup.exe`)

	// Download installer
	out, err := os.Create(installer)
	if err != nil {
		return false, err
	}
	defer func(out *os.File) {
		err = out.Close()
	}(out)
	resp, err := http.Get(bootstrapperURL)
	if err != nil {
		return false, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(resp.Body)
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return false, err
	}

	err = out.Close()
	if err != nil {
		return false, err
	}

	// Credit: https://stackoverflow.com/a/10385867
	cmd := exec.Command(installer)
	if err := cmd.Start(); err != nil {
		return false, err
	}
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus() == 0, nil
			}
		}
	}
	return true, nil
}

// Confirm will prompt the user with a message and OK / CANCEL buttons.
// Returns true if OK is selected by the user.
// Returns an error if something went wrong.
func Confirm(caption string, title string) (bool, error) {
	var flags uint = 0x00000001 // MB_OKCANCEL
	result, err := MessageBox(caption, title, flags)
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

// Error will an error message to the user.
// Returns an error if something went wrong.
func Error(caption string, title string) error {
	var flags uint = 0x00000010 // MB_ICONERROR
	_, err := MessageBox(caption, title, flags)
	return err
}

// MessageBox prompts the user with the given caption and title.
// Flags may be provided to customise the dialog.
// Returns an error if something went wrong.
func MessageBox(caption string, title string, flags uint) (int, error) {
	captionUTF16, err := syscall.UTF16PtrFromString(caption)
	if err != nil {
		return -1, err
	}
	titleUTF16, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return -1, err
	}
	ret, _, _ := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW").Call(
		uintptr(0),
		uintptr(unsafe.Pointer(captionUTF16)),
		uintptr(unsafe.Pointer(titleUTF16)),
		uintptr(flags))

	return int(ret), nil
}
