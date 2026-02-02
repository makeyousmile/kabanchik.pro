# Kabanchik.pro

Онлайн сервис заказа услуг. Стек: HTML + Pure.css (frontend), Go (backend), MongoDB.

## Запуск через Docker Compose

```
docker compose up --build
```

Сервис будет доступен на `http://localhost:8080`.

Остановить и удалить данные:

```
docker compose down -v
```

## Переменные окружения

- `PORT` — порт сервера (по умолчанию `8080`).
- `MONGO_URI` — строка подключения к MongoDB.
- `DB_NAME` — имя базы данных.
- `JWT_SECRET` — секрет для подписи токенов.
- `JWT_TTL` — срок жизни токена, например `24h`.

## Быстрые ссылки

- `http://localhost:8080/` — главная страница.
- `http://localhost:8080/docs.html` — документация.
- `http://localhost:8080/openapi.yaml` — OpenAPI.
