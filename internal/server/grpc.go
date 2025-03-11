package server

import (
	"fmt"
	"net"

	"github.com/shagrat164/system-monitoring-daemon/internal/config"
	"github.com/shagrat164/system-monitoring-daemon/internal/logger"
	"github.com/shagrat164/system-monitoring-daemon/internal/metrics"
	pb "github.com/shagrat164/system-monitoring-daemon/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Run - запускает gRPC-сервер.
func Run(cfg *config.Config, log *logger.Logger) error {
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Создаем буферизированный канал на 10 элементов
	loadChan := make(chan *pb.StatsResponse, 10)

	srv := grpc.NewServer()
	pb.RegisterMonitoringServer(srv, &monitoringServer{
		cfg:         cfg,
		log:         log,
		metricsChan: loadChan,
	})

	// Включаем reflection для удобства отладки с grpcurl
	reflection.Register(srv)

	log.Info(fmt.Sprintf("gRPC server started on port %s", cfg.GRPCPort))
	return srv.Serve(lis)
}

// monitoringServer - реализует интерфейс MonitoringServer.
type monitoringServer struct {
	pb.UnimplementedMonitoringServer
	cfg         *config.Config
	log         *logger.Logger
	metricsChan chan *pb.StatsResponse
}

// GetStats - реализует поток статистики.
func (s *monitoringServer) GetStats(req *pb.StatsRequest, stream pb.Monitoring_GetStatsServer) error {
	s.log.Info("New client connected to GetStats stream")

	// Передаём RealFileReader для реального чтения файла
	reader := metrics.RealFileReader{}
	// RealCommander для реального выполнения команд
	cmd := metrics.RealCommander{}

	// Запускаем сбор данных с учетом N и M из запроса клиента
	go metrics.CollectMetrics(stream.Context(), s.cfg, s.log, s.metricsChan, req.Interval, req.Duration, reader, cmd)

	// Передаем данные из канала в поток
	for {
		select {
		case stats := <-s.metricsChan:
			if err := stream.Send(stats); err != nil {
				s.log.Error(fmt.Sprintf("Failed to send stats: %v", err))
				return err
			}
		case <-stream.Context().Done():
			s.log.Info("Client disconnected")
			return nil
		}
	}
}
