package metrics

import (
	"context"
	"time"

	"github.com/shagrat164/system-monitoring-daemon/internal/config"
	"github.com/shagrat164/system-monitoring-daemon/internal/logger"
	pb "github.com/shagrat164/system-monitoring-daemon/proto"
)

// CollectMetrics - запускает сбор всех метрик.
func CollectMetrics(ctx context.Context,
	cfg *config.Config,
	log *logger.Logger,
	statsChan chan *pb.StatsResponse,
	interval, duration int32,
	reader FileReader,
	cmd Commander,
) {
	loadChan := make(chan *pb.StatsResponse)
	cpuChan := make(chan *pb.StatsResponse)
	diskChan := make(chan *pb.StatsResponse)
	filesystemChan := make(chan *pb.StatsResponse)

	// Запускаем сбор load average в отдельной горутине
	go CollectLoadAvg(ctx, cfg, log, loadChan, interval, duration, reader)

	// Запускаем сбор CPU статистики в отдельной горутине
	go CollectCPUStats(ctx, cfg, log, cpuChan, interval, duration, cmd)

	// Запускаем сбор статистики по ФС в отдельной горутине
	go CollectDiskStats(ctx, cfg, log, diskChan, interval, duration, cmd)
	go CollectFilesystemStats(ctx, cfg, log, filesystemChan, interval, duration, cmd)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := &pb.StatsResponse{}

		if cfg.Enabled.LoadAvg {
			loadStats := <-loadChan
			stats.LoadAverage_1Min = loadStats.GetLoadAverage_1Min()
			stats.LoadAverage_5Min = loadStats.GetLoadAverage_5Min()
			stats.LoadAverage_15Min = loadStats.GetLoadAverage_15Min()
		}
		if cfg.Enabled.CPU {
			cpuStats := <-cpuChan
			stats.CpuUser = cpuStats.GetCpuUser()
			stats.CpuSystem = cpuStats.GetCpuSystem()
			stats.CpuIdle = cpuStats.GetCpuIdle()
		}
		if cfg.Enabled.Disk {
			diskStats := <-diskChan
			stats.DiskStats = diskStats.GetDiskStats()
		}
		if cfg.Enabled.Filesystem {
			filesystemStats := <-filesystemChan
			stats.FilesystemStats = filesystemStats.GetFilesystemStats()
		}

		select {
		case <-ctx.Done():
			log.Debug("Gorutine CollectMetrics is done.")
			return
		case statsChan <- stats:
		}
	}
}
