# Проектная работа

Необходимо реализовать:
* [Системный мониторинг](./05-system-stats-daemon.md)

### Обязательные требования
* Наличие юнит-тестов на ключевые алгоритмы (core-логику) сервиса.
* Наличие валидных Dockerfile и Makefile/Taskfile для сервиса.
* Ветка master успешно проходит пайплайн в CI-CD системе 
(на ваш вкус, GitHub Actions, Circle CI, Travis CI, Jenkins, GitLab CI и пр.).
**Пайплайн должен в себе содержать**:
    - запуск последней версии `golangci-lint` на весь проект с
    [конфигом, представленным в данном репозитории](./.golangci.yml);
    - запуск юнит тестов командой вида `go test -race -count 100`;
    - сборку бинаря сервиса для версии Go не ниже 1.22. 

При невыполнении хотя бы одного из требований выше - максимальная оценка за проект **4 балла**
(незачёт), несмотря на, например, полностью написанный код сервиса.

Более подробная разбалловка представлена в описании.

### Использование сторонних библиотек для core-логики
Не допускается

---

Для упрощения проверки вашего репозитория, рекомендуем использовать значки GitHub
([GitHub badges](https://github.com/dwyl/repo-badges)), а также 
[Go Report Card](https://goreportcard.com/).

---
Авторы ТЗ:
- [Дмитрий Смаль](https://github.com/mialinx)
- [Антон Телышев](https://github.com/Antonboom)
