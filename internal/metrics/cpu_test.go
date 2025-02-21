package metrics

import (
	"reflect"
	"testing"
)

// MockCommandRunner мок для CommandRunner.
type MockCommandRunner struct {
	Output []byte
	Err    error
}

func (m *MockCommandRunner) CombinedOutput() ([]byte, error) {
	return m.Output, m.Err
}

func TestGetCPUStats(t *testing.T) { //nolint:dupl
	tests := []struct {
		name     string
		output   string
		expected *CPUStats
		wantErr  bool
	}{
		{
			name: "valid output",
			output: `Linux 6.11.0-17-generic (7US5DQ) 	21.02.2025 	_x86_64_	(4 CPU)

14:17:24        CPU     %user     %nice   %system   %iowait    %steal     %idle
14:17:25        all      1,52      0,00      0,25      0,00      0,00     98,22
Среднее:     all      1,52      0,00      0,25      0,00      0,00     98,22`,
			expected: &CPUStats{User: 1.52, System: 0.25, Idle: 98.22},
			wantErr:  false,
		},
		{
			name:     "invalid output",
			output:   "invalid data",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Мокаем CommandRunner
			runner := &MockCommandRunner{
				Output: []byte(tt.output),
				Err:    nil,
			}

			got, err := GetCPUStats(runner)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCPUStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GetCPUStats() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
