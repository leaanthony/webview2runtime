# WebView2Runtime

This package provides the following capabilities:
  
  - Retrieve version of installed WebView2 runtime, if installed
  - Determine if the installed version is older than the required version
  - Trigger an installation of the runtime using the Microsoft Bootstrapper 
  - Utility methods for user notifications and confirmations 

## Usage

The general workflow should be:

  - Check if there's a version installed using `GetInstalledVersion()`
  - If so, check it's new enough to support your application using `IsOlderThan()` 
  - If not, prompt the user if they wish to install/update the runtime using `Confirm()`
  - If the user wishes to proceed, Install the runtime using `InstallUsingBootstrapper()`

### Example

```go
package mypackage

import "github.com/leaanthony/webview2runtime"

func BootstrapRuntime() error {
	var err error
	shouldInstall := true
	message := "The WebView2 runtime is required. Press Ok to install."
	installedVersion := webview2runtime.GetInstalledVersion()
	if installedVersion != nil {
		shouldInstall, err = installedVersion.IsOlderThan("90.0.818.66")
		if err != nil {
			_ = webview2runtime.Error(err.Error(), "Error")
			return err
		}
		if shouldInstall {
			message = "The WebView2 runtime needs updating. Press Ok to install."
		}
	}
	if shouldInstall {
		confirmed, err := webview2runtime.Confirm(message, "Missing Requirements")
		if err != nil {
			return err
		}
		if confirmed {
			installedCorrectly, err := webview2runtime.InstallUsingBootstrapper()
			if err != nil {
				_ = webview2runtime.Error(err.Error(), "Error")
				return err
			}
			if !installedCorrectly {
				err = webview2runtime.Error("The runtime failed to install correctly. Please try again.", "Error")
				return err
			}
		}
	}
	return nil
}
```