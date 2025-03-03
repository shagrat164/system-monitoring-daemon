package metrics

import (
	"os"
	"os/exec"
)

// Commander - интерфейс для выполнения команд.
type Commander interface {
	Run(cmd string, args ...string) ([]byte, error)
}

// RealCommander - реальная реализация для запуска команд.
type RealCommander struct{}

func (r RealCommander) Run(cmd string, args ...string) ([]byte, error) {
	return exec.Command(cmd, args...).Output()
}

// FileReader - интерфейс для чтения файлов.
type FileReader interface {
	ReadFile(filename string) ([]byte, error)
}

// RealFileReader - реальная реализация интерфейса FileReader.
type RealFileReader struct{}

func (r RealFileReader) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}
