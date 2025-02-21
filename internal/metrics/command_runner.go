package metrics

import (
	"os/exec"
)

// CommandRunner интерфейс для выполнения команд.
type CommandRunner interface {
	CombinedOutput() ([]byte, error)
}

// RealCommand реализует CommandRunner для реальных команд.
type RealCommand struct {
	*exec.Cmd
}

// CombinedOutput выполняет команду и возвращает вывод.
func (r *RealCommand) CombinedOutput() ([]byte, error) {
	return r.Cmd.CombinedOutput()
}
