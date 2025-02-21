package metrics

import (
	"fmt"
	"strconv"
	"strings"
)

// CPUStats представляет собой статистику использования CPU.
type CPUStats struct {
	User   float64 // % времени в user mode
	System float64 // % времени в system mode
	Idle   float64 // % времени в idle mode
}

// GetCPUStats возвращает статистику использования CPU.
func GetCPUStats(runner CommandRunner) (*CPUStats, error) {
	// Выполняем команду sar
	output, err := runner.CombinedOutput()
	if err != nil {
		return nil, err
	}

	// Парсим вывод sar
	/* samle
	Linux 6.11.0-17-generic (7US5DQ) 	21.02.2025 	_x86_64_	(4 CPU)

	14:17:24        CPU     %user     %nice   %system   %iowait    %steal     %idle
	14:17:25        all      1,52      0,00      0,25      0,00      0,00     98,22
	Среднее:     all      1,52      0,00      0,25      0,00      0,00     98,22
	*/
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "all") {
			fields := strings.Fields(line)
			if len(fields) < 8 {
				return nil, fmt.Errorf("invalid sar output format")
			}

			user, _ := strconv.ParseFloat(strings.ReplaceAll(fields[2], ",", "."), 64)
			system, _ := strconv.ParseFloat(strings.ReplaceAll(fields[4], ",", "."), 64)
			idle, _ := strconv.ParseFloat(strings.ReplaceAll(fields[7], ",", "."), 64)

			return &CPUStats{
				User:   user,
				System: system,
				Idle:   idle,
			}, nil
		}
	}

	return nil, fmt.Errorf("cpu stats not found in top output")
}
