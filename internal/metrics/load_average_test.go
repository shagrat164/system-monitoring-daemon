package metrics

import (
	"reflect"
	"testing"
)

func TestGetLoadAverage(t *testing.T) { //nolint:dupl
	tests := []struct {
		name     string
		output   string
		expected *LoadAverage
		wantErr  bool
	}{
		{
			name:     "valid output",
			output:   "13:32:29 up 1 day,  8:35,  1 user,  load average: 0,34, 0,29, 0,40",
			expected: &LoadAverage{OneMinute: 0.34, FiveMinutes: 0.29, FifteenMinutes: 0.40},
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
			// Мокаем команду
			runner := &MockCommandRunner{
				Output: []byte(tt.output),
				Err:    nil,
			}

			got, err := GetLoadAverage(runner)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLoadAverage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GetLoadAverage() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
