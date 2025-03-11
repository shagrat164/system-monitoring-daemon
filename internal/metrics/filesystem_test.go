package metrics

import (
	"math"
	"testing"
	"time"

	"github.com/shagrat164/system-monitoring-daemon/internal/config"
	"github.com/shagrat164/system-monitoring-daemon/internal/logger"
	"github.com/shagrat164/system-monitoring-daemon/internal/model"
	pb "github.com/shagrat164/system-monitoring-daemon/proto"
)

func TestGetFilesystemStats(t *testing.T) {
	tests := []struct {
		name      string
		cmd       Commander
		wantStats []model.FilesystemStats
		wantErr   bool
	}{
		{
			name: "multiple mount points",
			cmd: &MockCommander{
				Outputs: [][]byte{
					[]byte(`Filesystem      Size  Used Avail Use% Mounted on
					/dev/sda1       50G   20G   30G  40% /
					/dev/sda1       50G   20G   30G  40% /mnt/extra`),
					[]byte(`Filesystem      Inodes  IUsed IFree IUse% Mounted on
					/dev/sda1      100000  40000 60000  40% /
					/dev/sda1      100000  40000 60000  40% /mnt/extra`),
				},
			},
			wantStats: []model.FilesystemStats{
				{
					Filesystem:    "/dev/sda1",
					UsedMB:        20 * 1024,
					UsedPercent:   40,
					InodesUsed:    40000,
					InodesPercent: 40,
				},
				{
					Filesystem:    "/dev/sda1",
					UsedMB:        20 * 1024,
					UsedPercent:   40,
					InodesUsed:    40000,
					InodesPercent: 40,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, err := GetFilesystemStats(tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFilesystemStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(stats) != len(tt.wantStats) {
				t.Errorf("GetFilesystemStats() got %d stats, want %d", len(stats), len(tt.wantStats))
			}
			for i, got := range stats {
				want := tt.wantStats[i]
				if got.Filesystem != want.Filesystem ||
					math.Abs(got.UsedMB-want.UsedMB) > 0.01 ||
					math.Abs(got.UsedPercent-want.UsedPercent) > 0.01 ||
					math.Abs(got.InodesUsed-want.InodesUsed) > 0.01 ||
					math.Abs(got.InodesPercent-want.InodesPercent) > 0.01 {
					t.Errorf("GetFilesystemStats() got = %+v, want %+v", got, want)
				}
			}
		})
	}
}

func TestCollectFilesystemStats(t *testing.T) {
	tests := []struct {
		name      string
		interval  int32
		duration  int32
		cmd       Commander
		wantCount int
		wantStats []model.FilesystemStats
	}{
		{
			name:     "basic averaging with silence",
			interval: 1,
			duration: 3,
			cmd: &MockCommander{
				Outputs: [][]byte{
					[]byte(`Filesystem      Size  Used Avail Use% Mounted on
					/dev/sda1       50G   20G   30G  40% /`),
					[]byte(`Filesystem      Inodes  IUsed IFree IUse% Mounted on
					/dev/sda1      100000  40000 60000  40% /`),
					[]byte(`Filesystem      Size  Used Avail Use% Mounted on
					/dev/sda1       50G   21G   29G  42% /`),
					[]byte(`Filesystem      Inodes  IUsed IFree IUse% Mounted on
					/dev/sda1      100000  41000 59000  41% /`),
					[]byte(`Filesystem      Size  Used Avail Use% Mounted on
					/dev/sda1       50G   22G   28G  44% /`),
					[]byte(`Filesystem      Inodes  IUsed IFree IUse% Mounted on
					/dev/sda1      100000  42000 58000  42% /`),
					[]byte(`Filesystem      Size  Used Avail Use% Mounted on
					/dev/sda1       50G   23G   27G  46% /`),
					[]byte(`Filesystem      Inodes  IUsed IFree IUse% Mounted on
					/dev/sda1      100000  43000 57000  43% /`),
				},
			},
			wantCount: 2, // Отправка начинается с t=2 (3-я итерация)
			wantStats: []model.FilesystemStats{
				{
					Filesystem:    "/dev/sda1",
					UsedMB:        21 * 1024, // Среднее за t=0, t=1, t=2
					UsedPercent:   42,
					InodesUsed:    41000,
					InodesPercent: 41,
				},
				{
					Filesystem:    "/dev/sda1",
					UsedMB:        22 * 1024, // Среднее за t=1, t=2, t=3
					UsedPercent:   44,
					InodesUsed:    42000,
					InodesPercent: 42,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewConfig()
			cfg.Enabled.Filesystem = true
			log, _ := logger.New(cfg.Logger)
			statsChan := make(chan *pb.StatsResponse, tt.wantCount+1)

			go CollectFilesystemStats(t.Context(), cfg, log, statsChan, tt.interval, tt.duration, tt.cmd)

			type fsStats struct {
				Filesystem    string
				UsedMB        float64
				UsedPercent   float64
				InodesUsed    float64
				InodesPercent float64
			}
			var gotStats []fsStats
			timeout := time.After(time.Duration((int(tt.duration)+tt.wantCount)*2) * time.Second) // Увеличенный таймаут
			for i := 0; i < tt.wantCount; i++ {
				select {
				case stats := <-statsChan:
					for _, fs := range stats.FilesystemStats {
						gotStats = append(gotStats, fsStats{
							Filesystem:    fs.Filesystem,
							UsedMB:        fs.UsedMb,
							UsedPercent:   fs.UsedPercent,
							InodesUsed:    fs.InodesUsed,
							InodesPercent: fs.InodesPercent,
						})
						t.Logf("Iteration %d: got %+v", i, fs)
					}
				case <-timeout:
					t.Fatalf("Timed out waiting for stats after %d iterations, got %d stats", tt.wantCount, len(gotStats))
				}
			}

			if len(gotStats) != tt.wantCount {
				t.Errorf("CollectFilesystemStats sent %d stats, want %d", len(gotStats), tt.wantCount)
			}

			const epsilon = 0.01
			for i, got := range gotStats {
				want := tt.wantStats[i]
				if got.Filesystem != want.Filesystem ||
					math.Abs(got.UsedMB-want.UsedMB) > epsilon ||
					math.Abs(got.UsedPercent-want.UsedPercent) > epsilon ||
					math.Abs(got.InodesUsed-want.InodesUsed) > epsilon ||
					math.Abs(got.InodesPercent-want.InodesPercent) > epsilon {
					t.Errorf("CollectFilesystemStats stats #%d = %+v, want %+v", i, got, want)
				}
			}
		})
	}
}
