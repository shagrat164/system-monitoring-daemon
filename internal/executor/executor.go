//go:build windows

package executor

import (
	"os/exec"
	"syscall"
)

// ExecutePowerShell - выполняет команду PowerShell и передаёт результат.
func ExecutePowerShell(comm string) (string, error) {
	var out []byte
	cmd := exec.Command("powershell", comm)
	cmd.SysProcAttr = &syscall.SysProcAttr{ // Выставить атрибуты
		HideWindow: true, // Спрятать окошко
	}

	out, err := cmd.Output()
	return string(out), err
}
