package sys

import "syscall"

var (
	shell32       = syscall.NewLazyDLL("shell32.dll")
	isUserAnAdmin = shell32.NewProc("IsUserAnAdmin")
)

// IsAdmin returns true if the current process is running with elevated Administrator privileges.
func IsAdmin() bool {
	ret, _, _ := isUserAnAdmin.Call()
	return ret != 0
}
