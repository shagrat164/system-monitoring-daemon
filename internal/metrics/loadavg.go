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

// CollectLoadAvg - собирает статистику load average и отправляет усреднённые данные в канал.
func CollectLoadAvg(cfg *config.Config,
	log *logger.Logger,
	loadChan chan *pb.StatsResponse,
	interval, duration int32,
	reader FileReader,
) {
	if !cfg.Enabled.LoadAvg {
		log.Info("Load average collection disabled")
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

	maxHistory := int(m / n)          // Максимальное количество записей в истории
	var history []model.LoadAvgRecord //nolint:prealloc

	ticker := time.NewTicker(n)
	defer ticker.Stop()

	for range ticker.C {
		load1, load5, load15, err := GetLoadAvg(reader)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to collect load average: %v", err))
			continue
		}

		history = append(history, model.LoadAvgRecord{
			Load1min:  load1,
			Load5min:  load5,
			Load15min: load15,
		})

		if len(history) > maxHistory {
			history = history[1:] // Cut off first record
		}
		if len(history) == 0 {
			continue
		}

		// "Молчим", пока не накопим maxHistory записей
		if len(history) < maxHistory {
			continue
		}

		var sum1, sum5, sum15 float64
		for _, stat := range history {
			sum1 += stat.Load1min
			sum5 += stat.Load5min
			sum15 += stat.Load15min
		}
		count := float64(len(history))
		if count == 0 {
			continue
		}

		stats := &pb.StatsResponse{
			LoadAverage_1Min:  round(sum1 / count),
			LoadAverage_5Min:  round(sum5 / count),
			LoadAverage_15Min: round(sum15 / count),
		}

		loadChan <- stats
	}
}

// round - округляет число до заданного количества знаков после запятой.
func round(val float64) float64 {
	p := float64(1)
	for i := 0; i < 2; i++ {
		p *= 10
	}
	return float64(int64(val*p+0.5)) / p
}

// GetLoadAvg - читает load average из /proc/loadavg с использованием FileReader.
func GetLoadAvg(reader FileReader) (float64, float64, float64, error) {
	data, err := reader.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0, err
	}

	fields := strings.Fields(strings.ReplaceAll(string(data), ",", "."))
	if len(fields) < 3 {
		return 0, 0, 0, fmt.Errorf("invalid loadavg format")
	}

	load1, _ := strconv.ParseFloat(strings.ReplaceAll(fields[0], ",", "."), 64)
	load5, _ := strconv.ParseFloat(fields[1], 64)
	load15, _ := strconv.ParseFloat(fields[2], 64)

	return load1, load5, load15, nil
}
