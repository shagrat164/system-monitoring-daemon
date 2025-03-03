# Проектная работа

Необходимо реализовать:
* [Системный мониторинг](./docs/SYSSTATS.md)

---
Используемая версия [golangci-lint]: <b>v1.64.5</b>
```
$ golangci-lint --version
golangci-lint has version v1.64.5 built with go1.24.0
```
---
Используемая версия [go]: <b>v1.24.0</b>
```
$ go version
go version go1.24.0 linux/amd64
```

# System Monitoring Daemon

Этот проект представляет собой демон системного мониторинга, который собирает и передаёт статистику о производительности системы клиентам через gRPC. Он предоставляет данные в реальном времени о загрузке системы, использовании CPU, операциях с дисками и состоянии файловых систем.

## Функциональность

- **Метрики**:
  - Средняя загрузка системы (load average).
  - Загрузка CPU (%user, %system, %idle).
  - Загрузка дисков (tps, KB/s).
  - Информация о дисках по файловым системам (объём, иноды).

- **Особенности**:
  - Настройка через файл конфигурации в формате TOML.
  - Конкурентный сбор метрик для эффективного использования ресурсов.
  - Усреднение данных за заданный период.
  - Клиентское приложение для отображения метрик в табличном формате.

## Технологии

- **Go**: Основной язык разработки демона.
- **gRPC**: Для передачи данных между демоном и клиентами.
- **TOML**: Формат файла конфигурации.

## Установка

1. Склонируйте репозиторий:
   ```bash
   git clone https://github.com/shagrat164/system-monitoring-daemon.git
   ```
2. Перейдите в директорию проекта:
   ```bash
   cd system-monitoring-daemon
   ```
3. Установите зависимости:
   ```bash
   go get
   ```

## Использование

- **Запуск сервера**:
  ```bash
  make run
  ```
  Это запустит демон мониторинга на порту, указанном в конфигурации (по умолчанию `50051`).

- **Запуск клиента**:
  ```bash
  make run-client
  ```
  Это запустит клиента для получения данных с сервера который запущен локально.
  Для запуска с определёнными параметрами воспользуйтесь командой:
  ```bash
  go run cmd/client/main.go -addr localhost:50051 -i 5 -d 15
  ```
  - `addr localhost:50051`: Адрес сервера.
  - `-i 5`: Интервал обновления данных в секундах.
  - `-d 15`: Период усреднения данных в секундах.

## Конфигурация

Настройки задаются через файл `config.toml`. Пример конфигурации:

```toml
grpc_port = "50051"

[logger]
level = "info"
path = "/var/log/daemon.log"

[metrics]
loadavg_enabled = true
cpu_enabled = true
disk_enabled = false
filesystem_enabled = true
```

- `grpc_port`: Порт, на котором работает сервер.
- `[logger]`: Настройки логгера (уровень logging и путь к лог-файлу).
- `[metrics]`: Включение/выключение сбора конкретных метрик.

## Тестирование

Для проверки кода используются юнит-тесты с обнаружением гонок данных:
```bash
make test
```
