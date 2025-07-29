# Subscription Aggregator Service

Небольшой REST-сервис на Go для управления онлайн-подписками пользователей.

## 🚀 Быстрый старт

1. Склонируйте репозиторий:
   ```bash
   git clone https://github.com/Novice-prog/subscription_service_EM
   cd rest-service
   
2. Создайте файл .env в корне:

    ```bash
    # .env
    DB_HOST=db
    DB_PORT=5432
    DB_USER=postgres
    DB_PASSWORD=yourpassword
    DB_NAME=subscriptions
    PORT=8080 
   ```

3. Запустите всё через Docker Compose:
    ```bash
    docker-compose down --volumes            # сбросить старые данные
    docker-compose up --build
   ```
   ---
| Метод  | Путь                  | Описание                                |
| ------ | --------------------- | --------------------------------------- |
| GET    | `/subscriptions`      | Список подписок (опция фильтрации)      |
| POST   | `/subscriptions`      | Создать новую подписку                  |
| GET    | `/subscriptions/{id}` | Получить подписку по UUID               |
| PUT    | `/subscriptions/{id}` | Обновить подписку                       |
| DELETE | `/subscriptions/{id}` | Удалить подписку                        |
| GET    | `/summary`            | Суммарная стоимость за период (MM-YYYY) |
