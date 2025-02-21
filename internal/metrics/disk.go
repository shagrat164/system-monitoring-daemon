package metrics

import (
	"fmt"
	"strconv"
	"strings"
)

// DiskStats представляет собой статистику использования дисков.
type DiskStats struct {
	TPS     float64 // Количество операций ввода-вывода в секунду
	KBRead  float64 // Количество прочитанных килобайт в секунду
	KBWrite float64 // Количество записанных килобайт в секунду
}

// GetDiskStats возвращает статистику использования дисков.
func GetDiskStats(runner CommandRunner) (map[string]DiskStats, error) {
	// Выполняем команду iostat
	output, err := runner.CombinedOutput()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]DiskStats)
	lines := strings.Split(string(output), "\n")

	// Парсим вывод iostat
	/* sample
	Linux 6.11.0-17-generic (7US5DQ) 	21.02.2025 	_x86_64_	(4 CPU)

	Device             tps    kB_read/s    kB_wrtn/s    kB_dscd/s    kB_read    kB_wrtn    kB_dscd
	loop0             0,19         8,59         0,00         0,00    1043282          0          0
	loop1             0,00         0,00         0,00         0,00        369          0          0
	loop10            0,00         0,00         0,00         0,00        380          0          0
	loop11            0,00         0,01         0,00         0,00       1142          0          0
	loop12            0,04         0,23         0,00         0,00      27710          0          0
	loop13            0,00         0,00         0,00         0,00        348          0          0
	loop14            0,10         2,37         0,00         0,00     287688          0          0
	loop15            0,04         1,86         0,00         0,00     225820          0          0
	loop16            0,00         0,00         0,00         0,00         54          0          0
	loop17            0,00         0,01         0,00         0,00        801          0          0
	loop18            0,00         0,00         0,00         0,00        116          0          0
	loop2             0,01         0,44         0,00         0,00      53040          0          0
	loop3             0,09         4,91         0,00         0,00     596192          0          0
	loop4             0,00         0,00         0,00         0,00         17          0          0
	loop5             0,00         0,01         0,00         0,00       1068          0          0
	loop6             0,00         0,00         0,00         0,00        347          0          0
	loop7             0,00         0,10         0,00         0,00      12753          0          0
	loop8             0,00         0,01         0,00         0,00       1074          0          0
	loop9             0,01         0,32         0,00         0,00      39202          0          0
	sda               6,20        83,26       105,85         0,00   10112613   12856332          0
	*/
	for i, line := range lines {
		if i < 3 {
			continue // Пропускаем заголовок
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		device := fields[0]
		tps, _ := strconv.ParseFloat(strings.ReplaceAll(fields[1], ",", "."), 64)
		kbRead, _ := strconv.ParseFloat(strings.ReplaceAll(fields[2], ",", "."), 64)
		kbWrite, _ := strconv.ParseFloat(strings.ReplaceAll(fields[3], ",", "."), 64)

		stats[device] = DiskStats{
			TPS:     tps,
			KBRead:  kbRead,
			KBWrite: kbWrite,
		}
	}

	if len(stats) == 0 {
		return nil, fmt.Errorf("invalid iostat format")
	}

	return stats, nil
}

// FilesystemStats представляет собой статистику использования файловой системы.
type FilesystemStats struct {
	UsedMB        float64 // Использовано мегабайт
	UsedPercent   float64 // Процент использования
	UsedInodes    float64 // Использовано inodes
	InodesPercent float64 // Процент использования inodes
}

// GetFilesystemStats возвращает статистику использования файловых систем.
func GetFilesystemStats(runner CommandRunner) (map[string]FilesystemStats, error) {
	// Выполняем команду df
	output, err := runner.CombinedOutput()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]FilesystemStats)
	lines := strings.Split(string(output), "\n")

	// Парсим вывод df
	for i, line := range lines {
		if i == 0 {
			continue // Пропускаем заголовок
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		mountPoint := fields[5]
		usedMB, _ := strconv.ParseFloat(strings.TrimSuffix(fields[1], "M"), 64)
		usedPercent, _ := strconv.ParseFloat(strings.TrimSuffix(fields[2], "%"), 64)
		usedInodes, _ := strconv.ParseFloat(fields[3], 64)
		inodesPercent, _ := strconv.ParseFloat(strings.TrimSuffix(fields[4], "%"), 64)

		stats[mountPoint] = FilesystemStats{
			UsedMB:        usedMB,
			UsedPercent:   usedPercent,
			UsedInodes:    usedInodes,
			InodesPercent: inodesPercent,
		}
	}

	if len(stats) == 0 {
		return nil, fmt.Errorf("invalid df format")
	}

	return stats, nil
}
