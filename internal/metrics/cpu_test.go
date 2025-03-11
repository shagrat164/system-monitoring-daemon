package metrics

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/shagrat164/system-monitoring-daemon/internal/config"
	"github.com/shagrat164/system-monitoring-daemon/internal/logger"
	"github.com/shagrat164/system-monitoring-daemon/internal/model"
	pb "github.com/shagrat164/system-monitoring-daemon/proto"
)

// MockCommander - мок для выполнения команд.
type MockCommander struct {
	Outputs [][]byte // Срез выводов для последовательных вызовов
	Err     error
	CallNum int
}

func (m *MockCommander) Run(cmd string, args ...string) ([]byte, error) { //nolint:revive
	if m.Err != nil {
		return nil, m.Err
	}
	if m.CallNum >= len(m.Outputs) {
		return m.Outputs[len(m.Outputs)-1], nil // Повторяем последний вывод
	}
	output := m.Outputs[m.CallNum]
	m.CallNum++
	return output, nil
}

// TestGetCPUStats - проверяет функцию GetCPUStats.
func TestGetCPUStats(t *testing.T) {
	tests := []struct {
		name        string
		cmd         Commander
		wantStats   model.CPUStats
		wantErr     bool
		errContains string
	}{
		{
			name: "valid data",
			cmd: &MockCommander{
				Outputs: [][]byte{
					[]byte("12:00:01 CPU %user %nice %system %iowait %steal %idle\n12:00:02 all 5.00 0.00 10.00 0.00 0.00 85.00\n"),
				},
			},
			wantStats: model.CPUStats{
				User:   5.00,
				System: 10.00,
				Idle:   85.00,
			},
			wantErr: false,
		},
		{
			name:        "command error",
			cmd:         &MockCommander{Err: errors.New("command failed")},
			wantStats:   model.CPUStats{},
			wantErr:     true,
			errContains: "sar command failed",
		},
		{
			name:        "invalid output",
			cmd:         &MockCommander{Outputs: [][]byte{[]byte("invalid data")}},
			wantStats:   model.CPUStats{},
			wantErr:     true,
			errContains: "no valid CPU stats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, err := GetCPUStats(tt.cmd)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetCPUStats() error = nil, want error containing %q", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetCPUStats() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("GetCPUStats() unexpected error: %v", err)
				return
			}

			const epsilon = 0.01
			if math.Abs(stats.User-tt.wantStats.User) > epsilon ||
				math.Abs(stats.System-tt.wantStats.System) > epsilon ||
				math.Abs(stats.Idle-tt.wantStats.Idle) > epsilon {
				t.Errorf("GetCPUStats() got = %+v, want %+v", stats, tt.wantStats)
			}
		})
	}
}

// TestCollectCPUStats - проверяет функцию CollectCPUStats с усреднением.
func TestCollectCPUStats(t *testing.T) {
	tests := []struct {
		name      string
		interval  int32
		duration  int32
		cmd       Commander
		wantCount int
		wantStats []model.CPUStats
	}{
		{
			name:     "basic averaging with silence",
			interval: 1,
			duration: 3,
			cmd: &MockCommander{
				Outputs: [][]byte{
					[]byte("12:00:01 CPU %user %nice %system %iowait %steal %idle\n12:00:02 all 5.00 0.00 10.00 0.00 0.00 85.00\n"),
					[]byte("12:00:01 CPU %user %nice %system %iowait %steal %idle\n12:00:02 all 10.00 0.00 20.00 0.00 0.00 70.00\n"),
					[]byte("12:00:01 CPU %user %nice %system %iowait %steal %idle\n12:00:02 all 15.00 0.00 30.00 0.00 0.00 55.00\n"),
					[]byte("12:00:01 CPU %user %nice %system %iowait %steal %idle\n12:00:02 all 20.00 0.00 40.00 0.00 0.00 40.00\n"),
				},
			},
			wantCount: 2, // Отправка начинается с t=2 (3-я итерация), всего 2 отправки за 4 итерации
			wantStats: []model.CPUStats{
				{User: 10.00, System: 20.00, Idle: 70.00}, // t=2: среднее за t=0, t=1, t=2
				{User: 15.00, System: 30.00, Idle: 55.00}, // t=3: среднее за t=1, t=2, t=3
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewConfig()
			cfg.Enabled.CPU = true
			log, _ := logger.New(cfg.Logger)
			statsChan := make(chan *pb.StatsResponse, tt.wantCount+1)

			go CollectCPUStats(t.Context(), cfg, log, statsChan, tt.interval, tt.duration, tt.cmd)

			type cpuStats struct {
				User   float64
				System float64
				Idle   float64
			}
			var gotStats []cpuStats
			timeout := time.After(time.Duration((int(tt.duration)+tt.wantCount)*2) * time.Second) // Увеличенный таймаут
			for i := 0; i < tt.wantCount; i++ {
				select {
				case stats := <-statsChan:
					got := cpuStats{
						User:   stats.CpuUser,
						System: stats.CpuSystem,
						Idle:   stats.CpuIdle,
					}
					gotStats = append(gotStats, got)
					t.Logf("Iteration %d: got %+v", i, got)
				case <-timeout:
					t.Fatalf("Timed out waiting for stats")
				}
			}

			if len(gotStats) != tt.wantCount {
				t.Errorf("CollectCPUStats sent %d stats, want %d", len(gotStats), tt.wantCount)
			}

			const epsilon = 0.01
			for i, got := range gotStats {
				want := tt.wantStats[i]
				if math.Abs(got.User-want.User) > epsilon ||
					math.Abs(got.System-want.System) > epsilon ||
					math.Abs(got.Idle-want.Idle) > epsilon {
					t.Errorf("CollectCPUStats stats #%d = %+v, want %+v", i, got, want)
				}
			}
		})
	}
}
