//go:build windows

package metrics

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shagrat164/system-monitoring-daemon/internal/executor"
	"github.com/shagrat164/system-monitoring-daemon/internal/model"
)

// GetCPUStats - получает CPU статистику с помощью команды PowerShell
func GetCPUStats(cmd Commander) (model.CPUStats, error) {
	command := "Get-CimInstance -ClassName Win32_PerfFormattedData_PerfOS_Processor -Filter \"Name='_Total'\" " +
		"| Select-Object PercentUserTime, PercentPrivilegedTime, PercentIdleTime"
	output, err := executor.ExecutePowerShell(command)
	if err != nil {
		return model.CPUStats{}, fmt.Errorf("Get-CimInstance command failed: %w", err)
	}

	lines := strings.Split(output, "\n")
	for n, line := range lines {
		if n != 3 { // Пропустить строки заголовков
			continue // Обрабатывать только строку с данными
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			user, err := strconv.ParseFloat(fields[0], 64) // PercentUserTime
			if err != nil {
				return model.CPUStats{}, fmt.Errorf("failed to parse PercentUserTime: %w", err)
			}
			system, err := strconv.ParseFloat(fields[1], 64) // PercentPrivilegedTime
			if err != nil {
				return model.CPUStats{}, fmt.Errorf("failed to parse PercentPrivilegedTime: %w", err)
			}
			idle, err := strconv.ParseFloat(fields[2], 64) // PercentIdleTime
			if err != nil {
				return model.CPUStats{}, fmt.Errorf("failed to parse PercentIdleTime: %w", err)
			}

			return model.CPUStats{
				User:   user,
				System: system,
				Idle:   idle,
			}, nil
		}
	}

	return model.CPUStats{}, fmt.Errorf("no valid CPU stats found in sar output")
}
