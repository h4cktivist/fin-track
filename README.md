# FinTrack

- **fin-api** - Сохраняет пользовательские транзакции, публикует все транзакции пользователя в Kafka после каждого добавления и предоставляет gRPC-метод для выборки данных fin-analytics.
- **fin-analytics** - Читает Kafka для пересчета статистики по транзакциям, использует Redis как кеш быстрых ответов и при необходимости подтягивает данные через gRPC fin-api.

## Технологии

- Go 1.25
- REST (chi), gRPC (google.golang.org/grpc)
- PostgreSQL (pgx), Redis (go-redis)
- Kafka (sarama)

## Запуск

```bash
docker compose up --build
```

Сервисы поднимутся на:

- fin-api HTTP `http://localhost:8080`, gRPC `localhost:9090`
- fin-analytics HTTP `http://localhost:8081`

## Маршруты

- `POST /v1/users/{userID}/transactions` — добавить транзакцию
- `GET /v1/users/{userID}/transactions` — все транзакции пользователя
- `PUT|PATCH /v1/users/{userID}/transactions/{transactionID}` — отредактировать транзакцию
- `DELETE /v1/users/{userID}/transactions/{transactionID}` — удалить транзакцию
- `GET /v1/users/{userID}/stats` — агрегированная статистика
- Swagger UI:
  - fin-api — `http://localhost:8080/swagger`
  - fin-analytics — `http://localhost:8081/swagger`

## Примеры запросов

```bash
# создать транзакцию
curl -X POST http://localhost:8080/v1/users/1/transactions \
  -H "Content-Type: application/json" \
  -d '{"amount":1200,"category":"salary","type":"income"}'

# обновить транзакцию
curl -X PUT http://localhost:8080/v1/users/1/transactions/1 \
  -H "Content-Type: application/json" \
  -d '{"amount":900,"category":"salary","type":"income"}'

# удалить транзакцию
curl -X DELETE http://localhost:8080/v1/users/1/transactions/1

# получить статистику
curl http://localhost:8081/v1/users/1/stats
```

## Тесты

```bash
go test -cover ./...
```

