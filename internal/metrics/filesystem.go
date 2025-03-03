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

// CollectFilesystemStats - собирает статистику файловых систем и отправляет усреднённые данные в канал.
func CollectFilesystemStats(cfg *config.Config,
	log *logger.Logger,
	statsChan chan *pb.StatsResponse,
	interval, duration int32,
	cmd Commander,
) {
	if !cfg.Enabled.Filesystem {
		log.Info("Load filesystem collection disabled")
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

	maxHistory := int(m / n)
	historyMap := make(map[string][]model.FilesystemStats)

	ticker := time.NewTicker(n)
	defer ticker.Stop()

	for range ticker.C {
		fsStats, err := GetFilesystemStats(cmd)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to collect filesystem stats: %v", err))
			continue
		}

		// Обновляем историю для каждой файловой системы
		for _, stat := range fsStats {
			h := historyMap[stat.MountPoint]
			h = append(h, stat)
			if len(h) > maxHistory {
				h = h[1:] // Обрезаем старую запись
			}
			historyMap[stat.MountPoint] = h
		}

		// "Молчим", пока не накопим maxHistory записей для всех файловых систем
		allReady := true
		for _, h := range historyMap {
			if len(h) < maxHistory {
				allReady = false
				break
			}
		}
		if !allReady {
			continue
		}

		// Формируем усреднённые данные
		var pbFsStats []*pb.FilesystemStats
		for mp, h := range historyMap {
			var sumUsedMB, sumUsedPercent, sumInodesUsed, sumInodesPercent float64
			for _, stat := range h {
				sumUsedMB += stat.UsedMB
				sumUsedPercent += stat.UsedPercent
				sumInodesUsed += stat.InodesUsed
				sumInodesPercent += stat.InodesPercent
			}
			count := float64(len(h))
			pbFsStats = append(pbFsStats, &pb.FilesystemStats{
				Filesystem:    h[0].Filesystem,
				Mountpoint:    mp,
				UsedMb:        round(sumUsedMB / count),
				UsedPercent:   round(sumUsedPercent / count),
				InodesUsed:    round(sumInodesUsed / count),
				InodesPercent: round(sumInodesPercent / count),
			})
		}

		if len(pbFsStats) > 0 {
			statsChan <- &pb.StatsResponse{
				FilesystemStats: pbFsStats,
			}
		}
	}
}

// GetFilesystemStats - получает статистику файловых систем с помощью df.
func GetFilesystemStats(cmd Commander) ([]model.FilesystemStats, error) {
	// Получаем данные об объёмах (df -h)
	outputH, err := cmd.Run("df", "-h")
	if err != nil {
		return nil, fmt.Errorf("df -h command failed: %w", err)
	}

	// Получаем данные об инодах (df -i)
	outputI, err := cmd.Run("df", "-i")
	if err != nil {
		return nil, fmt.Errorf("df -i command failed: %w", err)
	}

	// Парсим df -h
	linesH := strings.Split(strings.ReplaceAll(string(outputH), ",", "."), "\n")
	volumeRecords := make(map[string][]string) // Ключ - точка монтирования
	for i, line := range linesH {
		if i == 0 {
			continue // Пропускаем заголовок
		}
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			mountPoint := fields[5] // Точка монтирования (последнее поле)
			volumeRecords[mountPoint] = fields
		}
	}

	// Парсим df -i
	linesI := strings.Split(string(outputI), "\n")
	inodeRecords := make(map[string][]string) // Ключ - точка монтирования
	for i, line := range linesI {
		if i == 0 {
			continue // Пропускаем заголовок
		}
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			mountPoint := fields[5] // Точка монтирования (последнее поле)
			inodeRecords[mountPoint] = fields
		}
	}

	var stats []model.FilesystemStats
	for mountPoint, hFields := range volumeRecords {
		if iFields, ok := inodeRecords[mountPoint]; ok && len(hFields) >= 6 && len(iFields) >= 6 {
			usedMB, err := convSizeToMB(hFields[2]) // Used
			if err != nil {
				continue
			}
			usedPercent, err := strconv.ParseFloat(strings.TrimSuffix(hFields[4], "%"), 64) // Use%
			if err != nil {
				continue
			}
			inodesUsed, err := strconv.ParseFloat(iFields[2], 64) // IUsed
			if err != nil {
				continue
			}
			inodesPercent, err := strconv.ParseFloat(strings.TrimSuffix(iFields[4], "%"), 64) // IUse%
			if err != nil {
				continue
			}

			stats = append(stats, model.FilesystemStats{
				Filesystem:    hFields[0],
				MountPoint:    mountPoint,
				UsedMB:        usedMB,
				UsedPercent:   usedPercent,
				InodesUsed:    inodesUsed,
				InodesPercent: inodesPercent,
			})
		}
	}

	if len(stats) == 0 {
		return nil, fmt.Errorf("no valid filesystem stats found")
	}

	return stats, nil
}

// convSizeToMB - преобразует размер в мегабайты.
func convSizeToMB(size string) (float64, error) {
	size = strings.ToUpper(size)

	var multiplier float64
	var valStr string
	switch {
	case strings.HasSuffix(size, "G"):
		multiplier = 1024
		valStr = strings.TrimSuffix(size, "G")
	case strings.HasSuffix(size, "M"):
		multiplier = 1
		valStr = strings.TrimSuffix(size, "M")
	case strings.HasSuffix(size, "K"):
		multiplier = 1.0 / 1024
		valStr = strings.TrimSuffix(size, "K")
	default:
		multiplier = 1
		valStr = strings.TrimSuffix(size, "M")
	}

	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to conv size %s: %w", size, err)
	}

	return val * multiplier, nil
}
