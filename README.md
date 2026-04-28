```markdown
# service-subs

![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)
[![Swagger](https://img.shields.io/badge/Swagger-Documentation-brightgreen)](http://localhost:1323/swagger/index.html)

REST-сервис для агрегации данных об онлайн-подписках пользователей.

## 🚀 Функционал

### CRUDL операции с подписками

| Метод | Эндпоинт | Описание |
|-------|----------|----------|
| POST | `/sub` | Создание подписки (JSON body) → `{"id": 1}` |
| GET | `/sub?id=1` | Получение подписки по ID |
| PUT | `/sub?id=1` | Обновление подписки (JSON body) |
| DELETE | `/sub?id=1` | Удаление подписки |
| GET | `/subs` | Получение списка всех подписок |

### Расчет стоимости

| Метод | Эндпоинт | Описание |
|-------|----------|----------|
| GET | `/sum` | Сумма подписок за период (JSON body с фильтрами) → `{"sum": 150}` |

**Параметры фильтрации для `/sum`:**
- `user_id` (обязательный)
- `service_name` (опциональный)
- `start_date` (опциональный, формат MM-YYYY)
- `end_date` (опциональный, формат MM-YYYY)

## 🛠 Технологии

- **Go** 1.21+
- **PostgreSQL** 15+
- **Docker** / Docker Compose
- **Swagger** для документации API

## 📦 Установка и запуск

### Docker (рекомендуется)

```bash
# Запуск сервиса
sh docker_start.sh

# Остановка
sh docker_stop.sh
```
### Локальный запуск
```bash
# 1. Настройте конфигурационный файл
# Отредактируйте .env/config.yaml

# 2. Создайте базу данных PostgreSQL

# 3. Установите зависимости
go mod download

# 4. Сгенерируйте Swagger документацию
swag init

# 5. Запустите сервис
go run main.go
```

## 📚 Документация API
Swagger UI доступен по адресу: http://localhost:1323/swagger/index.html

## ⚙️ Конфигурация
Конфигурационные файлы:
  .env/config.yaml - настройки сервера и логирования

Пример config.yaml:
```yaml
app:
  name: service-subs
  version: "0.1.0"

http:
  host: localhost
  port: 8080

postgresql:
  name: webdb
  user: user
  password: password
  host: postgres
  port: 5432
    
logs:
  loglevel: -4
  # slog.LevelError = 8
  # slog.LevelWarn = 4
  # slog.LevelInfo = 0
  # slog.LevelDebug = -4
```

## 📊 Логирование
Код покрыт структурированными логами (slog)
Уровень логирования настраивается в конфигурационном файле

## 🛑 Graceful Shutdown
Сервер поддерживает плавное завершение работы:
Не принимает новые запросы после получения сигнала остановки
Дожидается завершения текущих запросов (таймаут 30 секунд)
Корректно закрывает соединения с базой данных