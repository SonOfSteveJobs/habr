# Architecture

## System Overview

```
                         ┌──────────┐
                         │  Client  │
                         └────┬─────┘
                              │ HTTP/JSON
                              ▼
                      ┌───────────────┐
                      │    Gateway    │
                      │   Service    │
                      └──┬────┬────┬──┘
                gRPC  /   │gRPC    \  gRPC
                     /    │         \
            ┌───────▼┐  ┌─▼──────┐  ┌▼────────────┐
            │  Auth  │  │Article │  │ Notification │
            │Service │  │Service │  │   Service    │
            └┬──┬──┬─┘  └┬────┬─-┘  └──────┬───────┘
             │  │  │     │    │            │
             │  │  ▼     │    │        ┌───▼───┐
             │  │  kafka │    │        │ Kafka │
             │  │    ▼   │    │        └──┬────┘
          ┌──▼┐ │    │  ┌▼-┐  ▼-──-┐      │
          │PG │ │    │  │PG│  │Redis      │
          └───┘ │    │  └──┘  └────┘      │
             ┌──▼──┐ │                    │
             │Redis│ └────────────────────┘
             └─────┘

        Auth ──── Kafka ────► Notification
        (producer)            (consumer)
```

## Services

### Gateway Service

**Stack:** HTTP (go-chi), gRPC client, JWT

**Role:** Единая точка входа. Принимает HTTP/JSON от клиента, маршрутизирует в микросервисы по gRPC. Не содержит бизнес-логики.

**JWT validation (локальная):**
Gateway и Auth Service используют один и тот же JWT secret. Gateway самостоятельно валидирует access token без обращения в Auth Service:

1. Разбивает токен: header, payload, signature
2. Вычисляет подпись тем же secret key
3. Сравнивает подписи
4. Проверяет expiration
5. Извлекает user_id из payload

Auth Service вызывается только для: register, login, verify-email, refresh token.

**Зачем:** Article Service доступен только авторизованным пользователям. Локальная валидация избавляет от лишних вызовов Auth Service на каждый запрос.

---

### Auth Service

**Stack:** gRPC, Postgres, Redis, Kafka, JWT

**Role:** Регистрация, логин, логаут, выдача и обновление JWT-токенов.

**Redis — хранение refresh токенов:**

- Key: `user_id`
- Value: `refresh:{SHA-256(refresh_token)}`
- TTL: время жизни refresh токена (30 дней)

Token rotation при refresh:
1. Проверить существование `refresh:{hash(old_token)}`
2. Удалить старый ключ
3. Сгенерировать новый refresh token (crypto/rand, 32 байта, base64url)
4. Записать `refresh:{hash(new_token)}` -> user_id с TTL
5. Выдать новую пару access + refresh

**Kafka — Transactional Outbox (exactly once):**

При регистрации пользователя Auth Service отправляет событие в Kafka для подтверждения email. Гарантия доставки — **exactly once**.

Паттерн Transactional Outbox:
1. В одной транзакции Postgres: `INSERT user` + `INSERT event` в таблицу outbox
2. Отдельный воркер читает outbox, отправляет в Kafka, помечает как отправленное
3. Idempotent producer: уникальный producer ID + sequence number, Kafka дедуплицирует повторные отправки

Решает проблему: user создан, а событие в Kafka не ушло.

---

### Article Service

**Stack:** gRPC, Postgres, Redis

**Role:** CRUD статей

**Курсорная пагинация:**
- Курсор: `base64(created_at:id)`
- Составной индекс: `(created_at DESC, id DESC)`
- Бесконечная лента

**Redis — кеш первой страницы:**
- Кешируется только запрос без курсора (первая страница, одинаковая для всех пользователей)
- Инвалидация: при создании/редактировании/удалении статьи — `DEL` ключа кеша, следующий GET пересоберет кеш
- TTL как страховка на случай, если инвалидация не сработала

---

### Notification Service

**Stack:** gRPC, Kafka

**Role:** Отправка email для подтверждения почты.

**Kafka (consumer):**
- Читает события регистрации из Kafka (от Auth Service)
- Отправляет email с кодом подтверждения
- Коммитит offset после отправки
- TTL на событие: если с момента регистрации прошло > N минут — событие дропается
- Если сервис лежал и события устарели — пользователь запрашивает повторное письмо сам

**gRPC (server):**
- Gateway прокидывает код подтверждения, который пользователь получил на email

---

## Communication Map

| Маршрут | Протокол | Описание |
|---------|----------|----------|
| Client → Gateway | HTTP/JSON | Единая точка входа |
| Gateway → Auth | gRPC | Регистрация, логин, verify-email, refresh |
| Gateway → Article | gRPC | CRUD статей |
| Gateway → Notification | gRPC | Подтверждение email (код от пользователя) |
| Auth → Notification | Kafka | Событие регистрации (exactly once) |

---
