package metrics

import (
	"fmt"
	"time"

	"github.com/shagrat164/system-monitoring-daemon/internal/config"
	"github.com/shagrat164/system-monitoring-daemon/internal/logger"
	"github.com/shagrat164/system-monitoring-daemon/internal/model"
	pb "github.com/shagrat164/system-monitoring-daemon/proto"
)

// CollectCPUStats - собирает CPU статистику и отправляет усреднённые данные в канал.
func CollectCPUStats(cfg *config.Config,
	log *logger.Logger,
	statsChan chan *pb.StatsResponse,
	interval, duration int32,
	cmd Commander,
) {
	if !cfg.Enabled.CPU {
		log.Info("Load CPU collection disabled")
		return
	}

	n := time.Duration(interval) * time.Second
	if n <= 0 {
		n = 5 * time.Second
	}
	m := time.Duration(duration) * time.Second
	if m <= 0 {
		m = 15 * time.Second
	}

	maxHistory := int(m / n)     // Максимальное количество записей в истории
	var history []model.CPUStats //nolint:prealloc

	ticker := time.NewTicker(n)
	defer ticker.Stop()

	for range ticker.C {
		cpuStats, err := GetCPUStats(cmd)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to collect CPU stats: %v", err))
			continue
		}

		history = append(history, cpuStats)
		if len(history) > maxHistory {
			history = history[1:] // Обрезаем первую запись
		}
		if len(history) == 0 {
			continue
		}

		// Условие "молчания": пропускаем отправку, пока не накопим maxHistory записей
		if len(history) < maxHistory {
			continue
		}

		// Теперь вычисляем средние значения и отправляем данные
		var sumUser, sumSystem, sumIdle float64
		for _, stat := range history {
			sumUser += stat.User
			sumSystem += stat.System
			sumIdle += stat.Idle
		}
		count := float64(len(history))
		if count == 0 {
			continue
		}

		stats := &pb.StatsResponse{
			CpuUser:   round(sumUser / count),
			CpuSystem: round(sumSystem / count),
			CpuIdle:   round(sumIdle / count),
		}

		log.Debug(stats.String())

		statsChan <- stats
	}
}
