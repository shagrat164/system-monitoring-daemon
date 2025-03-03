package metrics

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/shagrat164/system-monitoring-daemon/internal/config"
	"github.com/shagrat164/system-monitoring-daemon/internal/logger"
	"github.com/shagrat164/system-monitoring-daemon/internal/model"
	pb "github.com/shagrat164/system-monitoring-daemon/proto"
)

func TestGetDiskStats(t *testing.T) {
	tests := []struct {
		name        string
		cmd         Commander
		wantStats   []model.DiskStats
		wantErr     bool
		errContains string
	}{
		{
			name: "valid data",
			cmd: &MockCommander{
				Outputs: [][]byte{
					[]byte(`Linux 5.4.0-42-generic (host)  02/23/2025  _x86_64_  (4 CPU)
					
					Device            tps    kB_read/s    kB_wrtn/s    kB_read    kB_wrtn
					sda             10.00       40.00       20.00      40000      20000`),
				},
			},
			wantStats: []model.DiskStats{
				{Device: "sda", Tps: 10.00, KBs: 60.00},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, err := GetDiskStats(tt.cmd)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetDiskStats() error = nil, want error containing %q", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetDiskStats() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("GetDiskStats() unexpected error: %v", err)
				return
			}

			if len(stats) != len(tt.wantStats) {
				t.Errorf("GetDiskStats() got %d stats, want %d", len(stats), len(tt.wantStats))
			}
			for i, got := range stats {
				want := tt.wantStats[i]
				if got.Device != want.Device || math.Abs(got.Tps-want.Tps) > 0.01 || math.Abs(got.KBs-want.KBs) > 0.01 {
					t.Errorf("GetDiskStats() got = %+v, want %+v", got, want)
				}
			}
		})
	}
}

func TestCollectDiskStats(t *testing.T) {
	tests := []struct {
		name      string
		interval  int32
		duration  int32
		cmd       Commander
		wantCount int
		wantStats []model.DiskStats
	}{
		{
			name:     "basic averaging with silence",
			interval: 1,
			duration: 3,
			cmd: &MockCommander{
				Outputs: [][]byte{
					[]byte(`Linux 5.4.0-42-generic (host)  02/23/2025  _x86_64_  (4 CPU)
					
					Device            tps    kB_read/s    kB_wrtn/s    kB_read    kB_wrtn
					sda             10.00       40.00       20.00      40000      20000`),
					[]byte(`Linux 5.4.0-42-generic (host)  02/23/2025  _x86_64_  (4 CPU)
					
					Device            tps    kB_read/s    kB_wrtn/s    kB_read    kB_wrtn
					sda             20.00       80.00       40.00      80000      40000`),
					[]byte(`Linux 5.4.0-42-generic (host)  02/23/2025  _x86_64_  (4 CPU)
					
					Device            tps    kB_read/s    kB_wrtn/s    kB_read    kB_wrtn
					sda             30.00      120.00       60.00     120000      60000`),
					[]byte(`Linux 5.4.0-42-generic (host)  02/23/2025  _x86_64_  (4 CPU)
					
					Device            tps    kB_read/s    kB_wrtn/s    kB_read    kB_wrtn
					sda             40.00      160.00       80.00     160000      80000`),
				},
			},
			wantCount: 2, // Отправка начинается с t=2 (3-я итерация), всего 2 отправки за 4 итерации
			wantStats: []model.DiskStats{
				{Device: "sda", Tps: 20.00, KBs: 120.00}, // t=2: среднее за t=0, t=1, t=2
				{Device: "sda", Tps: 30.00, KBs: 180.00}, // t=3: среднее за t=1, t=2, t=3
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewConfig()
			cfg.Enabled.Disk = true
			log, _ := logger.New(cfg.Logger)
			statsChan := make(chan *pb.StatsResponse, tt.wantCount+1)

			go CollectDiskStats(cfg, log, statsChan, tt.interval, tt.duration, tt.cmd)

			type diskStats struct {
				Device string
				Tps    float64
				KBs    float64
			}
			var gotStats []diskStats
			timeout := time.After(time.Duration((int(tt.duration)+tt.wantCount)*2) * time.Second) // Увеличенный таймаут
			for i := 0; i < tt.wantCount; i++ {
				select {
				case stats := <-statsChan:
					for _, d := range stats.DiskStats {
						gotStats = append(gotStats, diskStats{
							Device: d.Device,
							Tps:    d.Tps,
							KBs:    d.KbTotal,
						})
						t.Logf("Iteration %d: got %+v", i, d)
					}
				case <-timeout:
					t.Fatalf("Timed out waiting for stats after %d iterations, got %d stats", tt.wantCount, len(gotStats))
				}
			}

			if len(gotStats) != tt.wantCount {
				t.Errorf("CollectDiskStats sent %d stats, want %d", len(gotStats), tt.wantCount)
			}

			const epsilon = 0.01
			for i, got := range gotStats {
				want := tt.wantStats[i]
				if got.Device != want.Device ||
					math.Abs(got.Tps-want.Tps) > epsilon ||
					math.Abs(got.KBs-want.KBs) > epsilon {
					t.Errorf("CollectDiskStats stats #%d = %+v, want %+v", i, got, want)
				}
			}
		})
	}
}
