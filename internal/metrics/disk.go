package metrics

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shagrat164/system-monitoring-daemon/internal/config"
	"github.com/shagrat164/system-monitoring-daemon/internal/logger"
	"github.com/shagrat164/system-monitoring-daemon/internal/model"
	pb "github.com/shagrat164/system-monitoring-daemon/proto"
)

// CollectDiskStats - собирает статистику дисков и отправляет усреднённые данные в канал.
func CollectDiskStats(cfg *config.Config,
	log *logger.Logger,
	statsChan chan *pb.StatsResponse,
	interval, duration int32,
	cmd Commander,
) {
	if !cfg.Enabled.Disk {
		log.Info("Load disk collection disabled")
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

	maxHistory := int(m / n) // Максимальное количество записей в истории
	historyMap := make(map[string][]model.DiskStats)

	ticker := time.NewTicker(n)
	defer ticker.Stop()

	for range ticker.C {
		diskStats, err := GetDiskStats(cmd)
		if err != nil {
			log.Error(fmt.Sprintf("failed to collect disk stats: %v", err))
			continue
		}

		for _, stat := range diskStats {
			h := historyMap[stat.Device]
			h = append(h, stat)
			if len(h) > maxHistory {
				h = h[1:] // Обрезаем первую запись
			}
			historyMap[stat.Device] = h
		}

		// Формируем усреднённые данные
		var pbDiskStats []*pb.DiskStats
		allDevicesReady := true
		for device, h := range historyMap {
			// Проверяем, накоплено ли достаточно записей для каждого устройства
			if len(h) < maxHistory {
				allDevicesReady = false
				break
			}

			var sumTps, sumKBs float64
			for _, stat := range h {
				sumTps += stat.Tps
				sumKBs += stat.KBs
			}
			count := float64(len(h))
			pbDiskStats = append(pbDiskStats, &pb.DiskStats{
				Device:  device,
				Tps:     round(sumTps / count),
				KbTotal: round(sumKBs / count),
			})
		}

		// Условие "молчания": отправляем данные только если все устройства имеют полную историю
		if !allDevicesReady || len(pbDiskStats) == 0 {
			continue
		}

		stats := &pb.StatsResponse{
			DiskStats: pbDiskStats,
		}

		log.Debug(stats.String())

		statsChan <- stats
	}
}

// GetDiskStats - получает статистику дисков с помощью команды iostat.
func GetDiskStats(cmd Commander) ([]model.DiskStats, error) {
	output, err := cmd.Run("iostat", "-d", "-k", "1", "1")
	if err != nil {
		return nil, fmt.Errorf("iostat command failed: %w", err)
	}

	lines := strings.Split(strings.ReplaceAll(string(output), ",", "."), "\n")
	var stats []model.DiskStats
	for i, line := range lines {
		if i < 3 {
			continue // Пропускаем первые три строки (системная информация и заголовок)
		}

		fields := strings.Fields(line)
		// Проверяем строку с данными (начинается с "Device" — заголовок, пропускаем)
		if len(fields) >= 6 && fields[0] != "" {
			tps, err := strconv.ParseFloat(fields[1], 64) // tps
			if err != nil {
				return nil, fmt.Errorf("failed to parse tps: %w", err)
			}
			kbRead, err := strconv.ParseFloat(fields[2], 64) // kB_read/s
			if err != nil {
				return nil, fmt.Errorf("failed to parse kB_read/s: %w", err)
			}
			kbWrite, err := strconv.ParseFloat(fields[3], 64) // kB_wrtn/s
			if err != nil {
				return nil, fmt.Errorf("failed to parse kB_wrtn/s: %w", err)
			}

			stats = append(stats, model.DiskStats{
				Device: fields[0],
				Tps:    tps,
				KBs:    kbRead + kbWrite,
			})
		}
	}

	if len(stats) == 0 {
		return nil, fmt.Errorf("no valid disk stats found in iostat output")
	}

	return stats, nil
}
