package server

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"sync"
	"time"

	"github.com/shagrat164/system-monitoring-daemon/internal/metrics"
	"github.com/shagrat164/system-monitoring-daemon/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// go: generate protoc --go_out=. --go-grpc_out=.
// --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/monitoring.proto

// MonitoringServer реализует интерфейс MonitoringServer.
type MonitoringServer struct {
	proto.UnimplementedMonitoringServer
}

func (s *MonitoringServer) GetStats(req *proto.StatsRequest, stream proto.Monitoring_GetStatsServer) error {
	interval := time.Duration(req.Interval) * time.Second
	duration := time.Duration(req.Duration) * time.Second

	// Кольцевой буфер для хранения значений load average
	type loadAvgSnapshot struct {
		timestamp time.Time
		value     float64
	}

	var mu sync.Mutex
	var loadAvgData []loadAvgSnapshot //nolint:prealloc

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		// Получаем текущее значение load average
		cmd := exec.Command("uptime")
		runner := &metrics.RealCommand{Cmd: cmd}
		loadAvg, err := metrics.GetLoadAverage(runner)
		if err != nil {
			log.Printf("Failed to get load average: %v", err)
			continue
		}

		log.Printf("GetLoadAverage(): %v", loadAvg)

		// Получаем текущее значение CPU
		cmd = exec.Command("sar", "-u", "1", "1")
		runner = &metrics.RealCommand{Cmd: cmd}
		cpuStats, err := metrics.GetCPUStats(runner)
		if err != nil {
			log.Printf("Failed to get CPU stats: %v", err)
			continue
		}

		log.Printf("GetCPUStats(): %v", cpuStats)

		// Получаем текущее значение дисков
		cmd = exec.Command("iostat", "-d", "-k")
		runner = &metrics.RealCommand{Cmd: cmd}
		diskStats, err := metrics.GetDiskStats(runner)
		if err != nil {
			log.Printf("Failed to get disk stats: %v", err)
			continue
		}

		log.Printf("GetDiskStats(): %v", diskStats)

		// Получаем текущее значение файловых систем
		cmd = exec.Command("df", "-BM", "--output=source,used,pcent,iused,ipcent,target")
		runner = &metrics.RealCommand{Cmd: cmd}
		filesystemStats, err := metrics.GetFilesystemStats(runner)
		if err != nil {
			log.Printf("Failed to get filesystem stats: %v", err)
			continue
		}

		log.Printf("GetFilesystemStats(): %v", filesystemStats)
		log.Println()

		// Добавляем текущее значение load average в буфер
		mu.Lock()
		loadAvgData = append(loadAvgData, loadAvgSnapshot{
			timestamp: time.Now(),
			value:     loadAvg.OneMinute,
		})

		// Удаляем старые значения, выходящие за пределы duration
		cutoff := time.Now().Add(-duration)
		for len(loadAvgData) > 0 && loadAvgData[0].timestamp.Before(cutoff) {
			loadAvgData = loadAvgData[1:]
		}
		mu.Unlock()

		// Рассчитываем среднее значение load average за последние M секунд
		var sum float64
		for _, snapshot := range loadAvgData {
			sum += snapshot.value
		}
		avgLoad := sum / float64(len(loadAvgData))

		// Преобразуем diskStats и filesystemStats в protobuf-формат
		var diskStatsProto []*proto.DiskStats
		for device, stats := range diskStats {
			diskStatsProto = append(diskStatsProto, &proto.DiskStats{
				Device:  device,
				Tps:     stats.TPS,
				KbRead:  stats.KBRead,
				KbWrite: stats.KBWrite,
			})
		}

		var filesystemStatsProto []*proto.FilesystemStats
		for mountPoint, stats := range filesystemStats {
			filesystemStatsProto = append(filesystemStatsProto, &proto.FilesystemStats{
				MountPoint:    mountPoint,
				UsedMb:        stats.UsedMB,
				UsedPercent:   stats.UsedPercent,
				UsedInodes:    stats.UsedInodes,
				InodesPercent: stats.InodesPercent,
			})
		}

		// Формируем StatsResponse
		stats := &proto.StatsResponse{
			LoadAverage_1Min:  avgLoad,
			LoadAverage_5Min:  loadAvg.FiveMinutes,
			LoadAverage_15Min: loadAvg.FifteenMinutes,
			CpuUser:           cpuStats.User,
			CpuSystem:         cpuStats.System,
			CpuIdle:           cpuStats.Idle,
			DiskStats:         diskStatsProto,
			FilesystemStats:   filesystemStatsProto,
		}

		// Отправляем данные клиенту
		if err := stream.Send(stats); err != nil {
			log.Printf("Failed to send stats: %v", err)
			return err
		}
	}

	return nil
}

// Run запускает GRPC-сервер.
func Run(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterMonitoringServer(grpcServer, &MonitoringServer{})

	// Включаем reflection для удобства отладки с grpcurl
	reflection.Register(grpcServer)

	fmt.Printf("GRPC server listening on port %d\n", port)
	return grpcServer.Serve(listener)
}
