# JWT Auth Service

Простое приложение для демонстрации аутентификации на Go с JWT, PostgreSQL, Docker и документацией через Swagger.

## Запуск

```bash
docker compose -f docker-compose.yml up --build
```

## Swagger

После запуска API-документация будет доступна по адресу:

[http://localhost:8000/swagger/index.html](http://localhost:8000/swagger/index.html)

---

## Заметки

- Конфигурация базы данных и другие параметры задаются через `.env` и `/configs/config.yml`
- Миграции выполняются автоматически при старте
