package metrics

import (
	"reflect"
	"testing"
)

func TestGetDiskStats(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected map[string]DiskStats
		wantErr  bool
	}{
		{
			name: "valid output",
			output: `Linux 6.11.0-17-generic (7US5DQ) 	21.02.2025 	_x86_64_	(4 CPU)

Device             tps    kB_read/s    kB_wrtn/s    kB_dscd/s    kB_read    kB_wrtn    kB_dscd
sdb               0,01         0,32         0,00         0,00      39202          0          0
sda               6,20        83,50       106,04         0,00   10112281   12843072          0`,
			expected: map[string]DiskStats{
				"sda": {TPS: 6.20, KBRead: 83.5, KBWrite: 106.04},
				"sdb": {TPS: 0.01, KBRead: 0.32, KBWrite: 0},
			},
			wantErr: false,
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

			got, err := GetDiskStats(runner)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDiskStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GetDiskStats() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestGetFilesystemStats(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected map[string]FilesystemStats
		wantErr  bool
	}{
		{
			name: "valid output",
			output: `Файл.система   Использовано Использовано% IИспользовано IИспользовано% Cмонтировано в
/dev/sda2            15475M           18%        344528             6% /
tmpfs                   67M            4%           208             1% /dev/shm`,
			expected: map[string]FilesystemStats{
				"/":        {UsedMB: 15475, UsedPercent: 18, UsedInodes: 344528, InodesPercent: 6},
				"/dev/shm": {UsedMB: 67, UsedPercent: 4, UsedInodes: 208, InodesPercent: 1},
			},
			wantErr: false,
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

			got, err := GetFilesystemStats(runner)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFilesystemStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GetFilesystemStats() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
