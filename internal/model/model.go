package model

// LoadAvgRecord - структура для хранения одного замера load average.
type LoadAvgRecord struct {
	Load1min  float64 // Нагрузка за 1 минуту
	Load5min  float64 // Нагрузка за 5 минут
	Load15min float64 // Нагрузка за 15 минут
}

// CPUStats - структура для хранения замеров CPU.
type CPUStats struct {
	User   float64 // Процент времени в user mode
	System float64 // Процент времени в system mode
	Idle   float64 // Процент времени в idle mode
}

// DiskStats - структура для хранения статистики дисков.
type DiskStats struct {
	Device string  // Имя устройства (например, "sda")
	Tps    float64 // Транзакции в секунду
	KBs    float64 // Килобайты в секунду (чтение + запись)
}

// FilesystemStats - структура для хранения статистики файловых систем.
type FilesystemStats struct {
	Filesystem    string  // Имя файловой системы
	MountPoint    string  // Точка монтирования фейловой системы
	UsedMB        float64 // Использовано мегабайт
	UsedPercent   float64 // Процент использованного объёма
	InodesUsed    float64 // Использовано инодов
	InodesPercent float64 // Процент использованных инодов
}
