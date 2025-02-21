package metrics

import (
	"fmt"
	"strconv"
	"strings"
)

// LoadAverage представляет собой среднюю загрузку системы.
type LoadAverage struct {
	OneMinute      float64
	FiveMinutes    float64
	FifteenMinutes float64
}

// GetLoadAverage возвращает среднюю загрузку системы.
func GetLoadAverage(runner CommandRunner) (*LoadAverage, error) {
	// Выполняем команду uptime
	output, err := runner.CombinedOutput()
	if err != nil {
		return nil, err
	}

	// Парсим вывод uptime
	line := string(output)
	loadIndex := strings.Index(line, "load average:")
	if loadIndex == -1 {
		return nil, fmt.Errorf("invalid uptime output format")
	}

	loadFields := strings.Fields(line[loadIndex:])
	if len(loadFields) < 5 {
		return nil, fmt.Errorf("invalid load average format")
	}

	oneMinute, _ := strconv.ParseFloat(strings.Replace(strings.TrimRight(loadFields[2], ","), ",", ".", 1), 64)
	fiveMinutes, _ := strconv.ParseFloat(strings.Replace(strings.TrimRight(loadFields[3], ","), ",", ".", 1), 64)
	fifteenMinutes, _ := strconv.ParseFloat(strings.Replace(loadFields[4], ",", ".", 1), 64)

	return &LoadAverage{
		OneMinute:      oneMinute,
		FiveMinutes:    fiveMinutes,
		FifteenMinutes: fifteenMinutes,
	}, nil
}
