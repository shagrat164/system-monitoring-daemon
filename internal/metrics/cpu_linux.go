//go:build linux

package metrics

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shagrat164/system-monitoring-daemon/internal/model"
)

// GetCPUStats - получает CPU статистику с помощью команды sar.
func GetCPUStats(cmd Commander) (model.CPUStats, error) {
	output, err := cmd.Run("sar", "-u", "1", "1")
	if err != nil {
		return model.CPUStats{}, fmt.Errorf("sar command failed: %w", err)
	}

	lines := strings.Split(strings.ReplaceAll(string(output), ",", "."), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 8 && strings.Contains(line, "all") {
			user, err := strconv.ParseFloat(fields[2], 64) // %user
			if err != nil {
				return model.CPUStats{}, fmt.Errorf("failed to parse user: %w", err)
			}
			system, err := strconv.ParseFloat(fields[4], 64) // %system
			if err != nil {
				return model.CPUStats{}, fmt.Errorf("failed to parse system: %w", err)
			}
			idle, err := strconv.ParseFloat(fields[7], 64) // %idle
			if err != nil {
				return model.CPUStats{}, fmt.Errorf("failed to parse idle: %w", err)
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
