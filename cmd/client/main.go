package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os/signal"
	"strconv"
	"syscall"

	pb "github.com/shagrat164/system-monitoring-daemon/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr     string // Адрес сервера
	interval string // Интервал выдачи данных
	duration string // Диапазон усреднения
)

func init() {
	flag.StringVar(&addr, "addr", "localhost:50051", "the address to connect to")
	flag.StringVar(&interval, "i", "5", "information release interval [s]")
	flag.StringVar(&duration, "d", "15", "range of information averaging [s]")
}

func main() {
	flag.Parse()

	intv, err := strconv.Atoi(interval)
	if err != nil {
		log.Printf("Convert param i error: %v\n", err)
		return
	}

	dur, err := strconv.Atoi(duration)
	if err != nil {
		log.Printf("Convert param d error: %v\n", err)
		return
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v\n", err)
		return
	}
	defer conn.Close()

	c := pb.NewMonitoringClient(conn)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	r, err := c.GetStats(ctx, &pb.StatsRequest{Interval: int32(intv), Duration: int32(dur)}) //nolint:gosec
	if err != nil {
		log.Printf("could not great: %v\n", err)
		return
	}

	for {
		stats, err := r.Recv()
		if errors.Is(err, io.EOF) {
			log.Println("Recive EOF. Close client.")
			break
		}
		if err != nil {
			log.Printf("err from Recv(): %v\n", err)
			break
		}

		// Очистка экрана
		clearTerminal()
		fmt.Printf("Address server: %s\n", addr)
		fmt.Printf("Internal = %s[s] Duration = %s[s]\n\n", interval, duration)

		// Вывод информации
		printLoadAvgTable(stats)
		printCPUTable(stats)
		printDiskTable(stats)
		printFiileSystemTable(stats)
	}
}

// Очистка экрана.
func clearTerminal() {
	fmt.Print("\033[H\033[2J") // ANSI-код для очистки терминала
}

// Таблица статистики load average.
func printLoadAvgTable(stats *pb.StatsResponse) {
	fmt.Println("Load Average:")
	fmt.Printf("  %-8s %-8s %-8s\n", "1 min", "5 min", "15 min")
	fmt.Printf("  %-8.2f %-8.2f %-8.2f\n",
		stats.GetLoadAverage_1Min(), stats.GetLoadAverage_5Min(), stats.GetLoadAverage_15Min())
	fmt.Println()
}

// Таблица статистики CPU.
func printCPUTable(stats *pb.StatsResponse) {
	fmt.Println("CPU Usage:")
	fmt.Printf("  %-10s %-10s %-10s\n", "User %", "System %", "Idle %")
	fmt.Printf("  %-10.2f %-10.2f %-10.2f\n",
		stats.GetCpuUser(), stats.GetCpuSystem(), stats.GetCpuIdle())
	fmt.Println()
}

// Таблица статистики дисков.
func printDiskTable(stats *pb.StatsResponse) {
	fmt.Println("Disk Usage:")
	fmt.Printf("  %-10s %-8s %-8s\n", "Device", "TPS", "KB/s")
	for _, disk := range stats.DiskStats {
		fmt.Printf("  %-10s %-8.2f %-8.2f\n",
			disk.GetDevice(), disk.GetTps(), disk.GetKbTotal())
	}
	fmt.Println()
}

// Таблица статистики файолвых систем.
func printFiileSystemTable(stats *pb.StatsResponse) {
	fmt.Println("Filesystem Usage:")
	fmt.Printf("  %-15s %-15s %-12s %-8s %-12s %-8s\n",
		"Filesystem", "Mount Point", "Used MB", "Used %", "Inodes Used", "Inodes %")
	for _, fs := range stats.FilesystemStats {
		fmt.Printf("  %-15s %-15s %-12.2f %-8.2f %-12.0f %-8.2f\n",
			fs.GetFilesystem(), fs.GetMountpoint(),
			fs.GetUsedMb(), fs.GetUsedPercent(),
			fs.GetInodesUsed(), fs.GetInodesPercent())
	}
	fmt.Println()
}
