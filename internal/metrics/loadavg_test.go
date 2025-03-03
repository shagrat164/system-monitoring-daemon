package metrics

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/shagrat164/system-monitoring-daemon/internal/config"
	"github.com/shagrat164/system-monitoring-daemon/internal/logger"
	pb "github.com/shagrat164/system-monitoring-daemon/proto"
)

// MockFileReader - мок для FileReader в тестах.
type MockFileReader struct {
	Data []byte
	Err  error
}

func (m MockFileReader) ReadFile(filename string) ([]byte, error) { //nolint:revive
	return m.Data, m.Err
}

// TestGetLoadAvg - проверяет функцию GetLoadAvg для разных случаев.
func TestGetLoadAvg(t *testing.T) {
	tests := []struct {
		name        string
		reader      FileReader
		wantLoad1   float64
		wantLoad5   float64
		wantLoad15  float64
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid data",
			reader:     MockFileReader{Data: []byte("0.15 0.25 0.35 1/100 12345"), Err: nil},
			wantLoad1:  0.15,
			wantLoad5:  0.25,
			wantLoad15: 0.35,
			wantErr:    false,
		},
		{
			name:        "file read error",
			reader:      MockFileReader{Data: nil, Err: errors.New("file not found")},
			wantLoad1:   0,
			wantLoad5:   0,
			wantLoad15:  0,
			wantErr:     true,
			errContains: "file not found",
		},
		{
			name:        "invalid format",
			reader:      MockFileReader{Data: []byte("0.15 0.25"), Err: nil},
			wantLoad1:   0,
			wantLoad5:   0,
			wantLoad15:  0,
			wantErr:     true,
			errContains: "invalid loadavg format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			load1, load5, load15, err := GetLoadAvg(tt.reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetLoadAvg() error = nil, want error containing %q", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetLoadAvg() error = %v, want error containing %q", err, tt.errContains)
				}
			} else if err != nil {
				t.Errorf("GetLoadAvg() unexpected error = %v", err)
			}

			if load1 != tt.wantLoad1 {
				t.Errorf("GetLoadAvg() load1 = %v, want %v", load1, tt.wantLoad1)
			}
			if load5 != tt.wantLoad5 {
				t.Errorf("GetLoadAvg() load5 = %v, want %v", load5, tt.wantLoad5)
			}
			if load15 != tt.wantLoad15 {
				t.Errorf("GetLoadAvg() load15 = %v, want %v", load15, tt.wantLoad15)
			}
		})
	}
}

// TestCollectLoadAvg - проверяет логику CollectLoadAvg.
func TestCollectLoadAvg(t *testing.T) {
	tests := []struct {
		name      string
		interval  int32
		duration  int32
		reader    FileReader
		wantLoads []struct{ l1, l5, l15 float64 }
		wantCount int
	}{
		{
			name:     "basic averaging",
			interval: 1,
			duration: 3,
			reader:   MockFileReader{Data: []byte("0.1 0.2 0.3 1/100 12345"), Err: nil},
			wantLoads: []struct{ l1, l5, l15 float64 }{
				{0.1, 0.2, 0.3},
				{0.1, 0.2, 0.3},
				{0.1, 0.2, 0.3},
				{0.1, 0.2, 0.3},
			},
			wantCount: 4,
		},
	}

	const epsilon = 0.0001

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewConfig()
			cfg.Enabled.LoadAvg = true
			log, _ := logger.New(cfg.Logger)
			loadChan := make(chan *pb.StatsResponse)

			go CollectLoadAvg(cfg, log, loadChan, tt.interval, tt.duration, tt.reader)

			var gotLoads []struct{ l1, l5, l15 float64 }
			timeout := time.After(time.Duration(tt.wantCount*2) * time.Second)
			for i := 0; i < tt.wantCount; i++ {
				select {
				case stats := <-loadChan:
					gotLoads = append(gotLoads, struct{ l1, l5, l15 float64 }{
						l1:  stats.LoadAverage_1Min,
						l5:  stats.LoadAverage_5Min,
						l15: stats.LoadAverage_15Min,
					})
				case <-timeout:
					t.Fatalf("Timed out waiting for stats")
				}
			}

			if len(gotLoads) != tt.wantCount {
				t.Errorf("CollectLoadAvg sent %d stats, want %d", len(gotLoads), tt.wantCount)
			}

			for i, got := range gotLoads {
				want := tt.wantLoads[i]
				if math.Abs(got.l1-want.l1) > epsilon ||
					math.Abs(got.l5-want.l5) > epsilon ||
					math.Abs(got.l15-want.l15) > epsilon {
					t.Errorf("CollectLoadAvg got load%d = %v, want %v", i+1, got, want)
				}
			}
		})
	}
}
