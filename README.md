```markdown
# Online Song Library API

## Описание

Это REST API для онлайн библиотеки песен. Он позволяет пользователям добавлять, удалять, изменять и просматривать информацию о песнях, а также получать тексты песен с пагинацией по куплетам. API использует PostgreSQL для хранения данных и может интегрироваться с внешним музыкальным API для получения дополнительной информации о песнях.

## Основные возможности

*   **CRUD операции для песен:**
    *   Добавление новой песни.
    *   Получение списка песен с фильтрацией по различным полям и пагинацией.
    *   Получение текста песни с пагинацией по куплетам.
    *   Изменение данных существующей песни.
    *   Удаление песни.
*   **Интеграция с внешним API:**  При добавлении песни происходит запрос к внешнему API для получения дополнительной информации (например, даты релиза, текста песни, ссылки на YouTube).
*   **Хранение данных в PostgreSQL:** Используется надежная и масштабируемая база данных PostgreSQL.
*   **Миграции БД:** Структура базы данных создается и обновляется с помощью миграций.
*   **Логирование:**  В коде реализовано подробное логирование для отслеживания работы API и отладки.
*   **Конфигурация через .env:** Все параметры конфигурации (например, параметры подключения к БД, URL внешнего API) вынесены в `.env` файл.
*   **Автоматическая генерация документации Swagger:** Документация API генерируется автоматически на основе комментариев в коде.
*   **Graceful Shutdown:** API поддерживает корректное завершение работы.

## Технологии

*   **Go:**  Язык программирования.
*   **PostgreSQL:** База данных.
*   **github.com/gorilla/mux:**  Роутер HTTP запросов.
*   **github.com/jackc/pgx/v5:**  Драйвер PostgreSQL.
*   **github.com/joho/godotenv:**  Загрузка конфигурации из .env файлов.
*   **github.com/golang-migrate/migrate/v4:**  Миграции базы данных.
*   **go.uber.org/zap:**  Логирование.
*   **github.com/golang/mock/gomock:** Создание mock объектов для unit тестов.
*   **github.com/stretchr/testify:**  Библиотека для написания тестов.
*   **Swaggo:**  Генерация документации Swagger.

## Требования

*   Go 1.20 или выше
*   Docker (для запуска PostgreSQL)
*   Docker Compose (опционально, для упрощения запуска)

## Запуск

**1. Клонируйте репозиторий:**

```bash
git clone <your_repository_url>
cd <your_repository_directory>
```

**2. Настройте .env файл:**

Создайте файл `.env` в корневом каталоге проекта и заполните его необходимыми параметрами (см. `.env.example`):

```
API_URL=http://localhost:8080/info
SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name
```

**3. Запустите PostgreSQL (опционально, с помощью Docker Compose):**

```bash
docker-compose up -d db
```

**4. Запустите API:**

```bash
go run ./cmd/songlibrary/main.go
```

## Использование API

API будет доступен по адресу `http://localhost:8080`.

**Документация Swagger:**

Документация API доступна по адресу `http://localhost:8080/swagger/index.html`.

**Примеры запросов:**

*   **Получение списка песен:**

    ```
    GET /songs?page=1&pageSize=10&group=Muse
    ```

*   **Добавление новой песни:**

    ```
    POST /songs
    Content-Type: application/json

    {
      "group": "Muse",
      "song": "Uprising"
    }
    ```

*   **Получение текста песни (с пагинацией):**

    ```
    GET /songs/1/text?page=1&pageSize=1
    ```

*   **Изменение данных песни:**

    ```
    PUT /songs/1
    Content-Type: application/json

    {
      "id": 1,
      "group": "New Muse",
      "song": "Uprising",
      "releaseDate": "2009-09-14",
      "text": "New lyrics",
      "link": "https://youtube.com/..."
    }
    ```

*   **Удаление песни:**

    ```
    DELETE /songs/1
    ```

## Запуск тестов

```bash
go test ./...
```

## Зависимости

Зависимости проекта управляются с помощью Go Modules. Вы можете обновить или добавить зависимости, используя команды `go get` и `go mod tidy`.

## Развертывание

Пример systemd unit файла, который можно использовать для развертывания сервиса

```
[Unit]
Description=Song Library API
After=network.target postgresql.service  # Adjust if your DB has a different service name
Requires=postgresql.service

[Service]
Type=simple
User=youruser  # Replace with the user you want to run the service as
WorkingDirectory=/path/to/your/songlibrary  # Replace with the actual path
ExecStart=/path/to/your/songlibrary/songlibrary  # Replace with the actual path
Restart=on-failure
EnvironmentFile=/path/to/your/songlibrary/.env # Replace with the actual path

[Install]
WantedBy=multi-user.target
```


