# LectureQuiz Bot

Telegram бот для проведения квизов на лекциях.

## Задание

Реализуйте Telegram бота, который позволяет преподавателю проводить квизы для студентов в реальном времени.

### Основной flow

**Преподаватель:**
1. Отправляет боту JSON файл с квизом
2. Бот создаёт лобби и показывает ссылку для студентов
3. Видит счётчик подключившихся (обновляется каждые 3 сек через `editMessageText`)
4. Нажимает "Начать квиз"
5. После завершения скачивает CSV с результатами

**Студент:**
1. Переходит по ссылке, нажимает "Присоединиться"
2. Ждёт в лобби начала квиза
3. Получает вопросы синхронно со всеми (общий таймер)
4. Отвечает, отправив букву в чат (A, B, C, D, E или F)
5. В конце видит свой результат и топ-10

> **Важно:** Ответы принимаются в виде текстовых сообщений с буквами A-F (регистр не важен).

---

## Функциональные требования

### Базовая часть (5 баллов)

| Критерий | Баллы | Описание |
|----------|-------|----------|
| Загрузка квиза из JSON | 1 | Парсинг JSON, валидация обязательных полей |
| Лобби с обновлением счётчика | 0.5 | `editMessageText` каждые 3 сек |
| Синхронный квиз с таймером | 1.5 | Горутины для таймеров, рассылка вопросов всем |
| Результаты + leaderboard | 1 | Подсчёт баллов, сортировка по баллам и времени, показ топ-10 |
| CSV экспорт | 0.5 | Выгрузка результатов для преподавателя |
| CLI конфигурация + тесты | 0.5 | `--token` флаг, unit-тесты |

### Усложнённая часть (5 баллов)

| Критерий | Баллы | Описание |
|----------|-------|----------|
| Настраиваемые поля регистрации | 0.5 | В JSON: `"registration": ["name", "group"]` |
| Storage interface + in-memory | 1 | Интерфейс Storage, реализация в памяти |
| Сохранение квизов | 0.5 | Квиз сохраняется после загрузки |
| Мои квизы + скачать JSON | 0.5 | Список квизов, экспорт в JSON |
| Повторное проведение | 0.5 | Провести тот же квиз ещё раз |
| История запусков | 0.5 | Все результаты всех запусков |
| Override параметров на вопрос | 0.5 | points, time, shuffle per question |
| Лимит участников | 0.5 | `max_participants` в настройках |
| Тесты с моками | 0.5 | gomock/testify для Storage |

### Бонус (1 балл)

| Критерий | Баллы | Описание |
|----------|-------|----------|
| SQLite Storage | 0.5 (если не pg) | Реализация Storage через `modernc.org/sqlite` + `sqlx` |
| PostgreSQL Storage | 1 (если не sqlite) | Реализация Storage через `pgx` + `sqlx` |

---

## JSON формат квиза

### Базовый формат

```json
{
  "title": "Лекция 5: Конкурентность",
  "settings": {
    "time_per_question": 20,
    "shuffle_questions": true,
    "shuffle_answers": true
  },
  "questions": [
    {
      "text": "Что делает sync.Mutex?",
      "options": [
        "Блокирует доступ к ресурсу",
        "Создаёт горутину",
        "Закрывает канал",
        "Отправляет в канал"
      ],
      "correct": 0,
      "explanation": "Mutex обеспечивает взаимное исключение"
    }
  ]
}
```

### Расширенный формат (усложнённая часть)

```json
{
  "title": "Лекция 5: Конкурентность",
  "settings": {
    "time_per_question": 20,
    "shuffle_questions": true,
    "shuffle_answers": true,
    "max_participants": 100,
    "registration": ["name", "group"]
  },
  "questions": [
    {
      "text": "Что делает sync.Mutex?",
      "options": [
        "Блокирует доступ к ресурсу",
        "Создаёт горутину",
        "Закрывает канал",
        "Отправляет в канал"
      ],
      "correct": 0,
      "explanation": "Mutex обеспечивает взаимное исключение",
      "points": 10,
      "time": 30,
      "shuffle": false
    }
  ]
}
```

### Описание полей

**settings:**

| Поле | Тип | Обязательное | По умолчанию | Описание |
|------|-----|--------------|--------------|----------|
| `time_per_question` | int | да | - | Время на вопрос в секундах |
| `shuffle_questions` | bool | нет | false | Перемешивать порядок вопросов |
| `shuffle_answers` | bool | нет | false | Перемешивать варианты ответов |
| `max_participants` | int | нет | 0 (без лимита) | Максимум участников |
| `registration` | []string | нет | [] | Поля для регистрации |

**question:**

| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| `text` | string | да | Текст вопроса |
| `options` | []string | да | Варианты ответов (2-6 штук) |
| `correct` | int | да | Индекс правильного ответа (0-based) |
| `explanation` | string | нет | Пояснение к ответу |
| `points` | int | нет | Баллы за вопрос (по умолчанию 1) |
| `time` | int | нет | Переопределение времени |
| `shuffle` | bool | нет | Переопределение shuffle_answers |

---

## Интерфейсы

Вам необходимо реализовать следующие интерфейсы (определены в файлах с тегом `!change`):

### QuizEngine

```
type QuizEngine interface {
    // LoadQuiz парсит JSON и создаёт квиз
    LoadQuiz(data []byte) (*Quiz, error)
    
    // StartRun создаёт новый запуск квиза
    StartRun(ctx context.Context, quiz *Quiz) (*QuizRun, error)
    
    // JoinRun добавляет участника в запуск
    JoinRun(ctx context.Context, runID string, participant *Participant) error
    
    // GetParticipantCount возвращает количество участников
    GetParticipantCount(runID string) int
    
    // StartQuiz запускает квиз (переход из лобби)
    StartQuiz(ctx context.Context, runID string) (<-chan QuizEvent, error)
    
    // SubmitAnswer регистрирует ответ участника по индексу
    SubmitAnswer(ctx context.Context, runID string, participantID int64, questionIdx int, answerIdx int) error
    
    // SubmitAnswerByLetter регистрирует ответ по букве (A, B, C, D, E, F)
    SubmitAnswerByLetter(ctx context.Context, runID string, participantID int64, letter string) error
    
    // GetCurrentQuestion возвращает текущий номер вопроса (-1 если не запущен)
    GetCurrentQuestion(runID string) int
    
    // GetResults возвращает результаты квиза
    GetResults(runID string) (*QuizResults, error)
    
    // ExportCSV экспортирует результаты в CSV
    ExportCSV(runID string) ([]byte, error)
}
```

> **Механизм ответов:**  
> Участники отвечают, отправляя букву (A-F) в чат. Бот вызывает `SubmitAnswerByLetter`.
> Буква преобразуется в индекс (A=0, B=1, ...) и передаётся в `SubmitAnswer`.
> Регистр букв игнорируется.

### Storage (усложнённая часть)

```
type Storage interface {
    // SaveQuiz сохраняет квиз
    SaveQuiz(ctx context.Context, quiz *Quiz) error
    
    // GetQuiz возвращает квиз по ID
    GetQuiz(ctx context.Context, id string) (*Quiz, error)
    
    // ListQuizzes возвращает список квизов пользователя
    ListQuizzes(ctx context.Context, ownerID int64) ([]*Quiz, error)
    
    // SaveRun сохраняет запуск квиза
    SaveRun(ctx context.Context, run *QuizRun) error
    
    // GetRun возвращает запуск по ID
    GetRun(ctx context.Context, id string) (*QuizRun, error)
    
    // ListRuns возвращает список запусков квиза
    ListRuns(ctx context.Context, quizID string) ([]*QuizRun, error)
}
```

### TelegramClient

```
type TelegramClient interface {
    // SendMessage отправляет сообщение
    SendMessage(chatID int64, text string, opts *SendOptions) (*Message, error)
    
    // EditMessage редактирует сообщение (для обновления счётчика в лобби)
    EditMessage(chatID int64, messageID int, text string, opts *SendOptions) error
    
    // GetUpdates получает обновления (long polling)
    GetUpdates(offset int, timeout int) ([]Update, error)
}
```

---

## Тестирование

### Требования к покрытию

Вы должны написать собственные тесты для достижения покрытия **70%** кода в пакете `internal/quiz`.

```go
//go:build !change

package quiz

// min coverage: . 70%
```

### Запуск тестов

```bash
# Запуск тестов
go test ./...

# С race detector
go test -race ./...

# Проверка покрытия
go test -cover ./internal/events/...
```

### Что нужно протестировать

Помимо публичных тестов, вам нужно написать свои тесты для:

- Конкурентные операции (одновременные ответы, присоединение)
- Edge cases (пустой квиз, один вариант ответа, и т.д.)
- Graceful shutdown (отмена context)
- Shuffle (вопросов и ответов)
- Повторные ответы на один вопрос
- Таймауты (участник не ответил вовремя)
- "Отвал" участников (часть участников перестаёт отвечать)

---

## Сборка и запуск

### Сборка

```bash
cd quizbot/cmd/quizbot
go build .
```

### Запуск

```bash
./quizbot --token=YOUR_TELEGRAM_BOT_TOKEN --bot-username=your_bot_username
```

### CLI параметры

| Параметр | Обязательный | Описание |
|----------|--------------|----------|
| `--token` | да | Токен Telegram бота |
| `--bot-username` | да | Username бота (без @) для формирования ссылок |

### Получение токена и username

1. Напишите [@BotFather](https://t.me/BotFather) в Telegram
2. Отправьте команду `/newbot`
3. Следуйте инструкциям
4. Скопируйте полученный токен
5. Username бота — это то, что вы указали при создании (например, `my_quiz_bot`)

Ссылка для участников формируется как: `https://t.me/<bot-username>?start=join_<runID>`

### Запуск через Makefile

Проект содержит `Makefile` с удобными целями для разработки и деплоя. Make автоматически включает переменные из файла `.env` (Makefile делает `include .env`), поэтому рекомендуется создать `.env` рядом с `Makefile` или передавать переменные при вызове `make`.

Пример `.env` (опционально):

```
# Путь к входной точке приложения для go run / go build
MAIN=./cmd/quizbot
# Имя бинарника при сборке
APP_NAME=quizbot
# Строка подключения для миграций (используется целями migrate-up/migrate-down)
DB_CONN_URL=postgres://user:pass@localhost:5432/quizbot?sslmode=disable
```

Доступные цели и примеры использования:

- `make run`
  - Запускает приложение через `go run` (использует переменную `MAIN`).
  - Пример: `make run MAIN=./cmd/quizbot`

- `make build`
  - Собирает бинарник в `bin/${APP_NAME}` (использует `MAIN` и `APP_NAME`).
  - Пример: `make build APP_NAME=quizbot MAIN=./cmd/quizbot`

- `make test`
  - Запускает `go test ./...` по всему проекту.

- `make clean`
  - Удаляет папку `bin`.

- `make migrate-up`
  - Применяет миграции из каталога `migrations` к БД, используя `DB_CONN_URL`.
  - Пример: `make migrate-up DB_CONN_URL="postgres://user:pass@localhost:5432/quizbot?sslmode=disable"`

- `make migrate-down`
  - Откатывает миграции (использует `DB_CONN_URL`).

- `make deploy`
  - Поднимает контейнеры через `docker compose up -d`.

- `make undeploy`
  - Останавливает и удаляет приложение из Docker: `docker compose down app`.

Примеры полного рабочего сценария:

1) Быстрая локальная сборка и запуск бинарника:

```bash
make build APP_NAME=quizbot MAIN=./cmd/quizbot
./bin/quizbot --token="$TELEGRAM_TOKEN" --bot-username="your_bot_username"
```

2) Запуск без сборки (go run):

```bash
make run MAIN=./cmd/quizbot --silent
```

3) Применение миграций:

```bash
make migrate-up DB_CONN_URL="${DB_CONN_URL}"
```

Примечания:
- Если вы используете `.env`, Make автоматически загрузит переменные оттуда; в противном случае передавайте `MAIN`, `APP_NAME` и `DB_CONN_URL` при вызове `make`.
- Убедитесь, что у вас установлены `migrate` и `docker compose` если вы будете использовать соответствующие цели.

---

## Структура проекта

```
quizbot/
├── cmd/quizbot/
│   └── main.go              # Точка входа (ваш код)
├── internal/
│   ├── quiz/
│   │   ├── types.go         # Структуры данных (не менять)
│   │   ├── engine.go        # QuizEngine (ваш код)
│   │   ├── quiz_test.go     # Публичные тесты (не менять)
│   │   └── coverage_test.go # Требование покрытия (не менять)
│   ├── telegram/
│   │   ├── types.go         # Интерфейс TelegramClient (не менять)
│   │   ├── client.go        # Реализация клиента (ваш код)
│   │   └── bot.go           # Бот (ваш код)
│   └── storage/
│       ├── storage.go       # Интерфейс Storage (не менять)
│       └── memory.go        # In-memory реализация (ваш код)
├── testdata/
│   └── quizzes/             # Тестовые JSON файлы
└── README.md
```

### Файлы с тегом `!change`

Файлы с тегом `//go:build !change` **нельзя изменять** — они будут перезаписаны при проверке:
- `internal/quiz/types.go` — структуры данных
- `internal/quiz/quiz_test.go` — публичные тесты
- `internal/quiz/coverage_test.go` — требование покрытия
- `internal/storage/storage.go` — интерфейс Storage
- `internal/telegram/types.go` — интерфейс TelegramClient

---

## Пример взаимодействия

### Вопрос в чате

```
Вопрос 1

Что делает sync.Mutex?

A. Блокирует доступ к ресурсу
B. Создаёт горутину
C. Закрывает канал
D. Отправляет в канал

⏱ 20 секунд

Отправьте букву ответа (A, B, C, ...)
```

### Ответ студента

```
A
```

### Подтверждение

```
✅ Ответ принят!
```

### Результаты

```
🏆 Результаты квиза "Лекция 5: Конкурентность"

Ваш результат: 8 баллов (место 3)

Топ-10:
🥇 1. @alice — 10 баллов
🥈 2. @bob — 9 баллов
🥉 3. @charlie — 8 баллов
4. @david — 7 баллов
...
```

---

## Подсказки

### Конкурентность

- Используйте `sync.Mutex` или `sync.RWMutex` для защиты общих данных
- Таймер для каждого вопроса — отдельная горутина
- Канал `QuizEvent` для уведомления о событиях квиза
- `context.Context` для graceful shutdown

### Telegram API

- Long polling через `getUpdates`
- `editMessageText` для обновления счётчика в лобби
- Deep links: `t.me/<bot-username>?start=join_<runID>` для присоединения
- Username бота передаётся через CLI и используется для генерации ссылок
- Во время вопроса бот обновляет сообщение (таймер) у студентов и статус у преподавателя
- CSV результаты отправляются преподавателю как файл

### Тестирование

- `go.uber.org/goleak` для проверки утечек горутин
- `github.com/stretchr/testify` для assertions
- Тестируйте с `-race` флагом

---

## Логика подсчета результатов

### Подсчет времени ответа

- **Если участник ответил**: считается фактическое время ответа с момента показа вопроса
- **Если участник НЕ ответил** (таймаут): засчитывается полное время вопроса (например, 20 секунд)
- **Если участник присоединился во время квиза**: для пропущенных вопросов засчитывается полное время каждого вопроса
- **Общее время**: сумма времени по всем вопросам

### Сортировка в leaderboard

Участники сортируются по следующим критериям (в порядке приоритета):
1. **Количество баллов** (больше баллов — выше место)
2. **Общее время ответов** (при равных баллах: меньше время — выше место)

**Пример:**
- Алиса: 10 баллов, 45 секунд общего времени → 1 место
- Боб: 10 баллов, 52 секунды общего времени → 2 место  
- Чарли: 9 баллов, 30 секунд общего времени → 3 место

> **Важно:** 
> - Неответ на вопрос штрафует не только отсутствием балла, но и максимальным временем для этого вопроса
> - Позднее присоединение не даёт преимущества по времени — пропущенные вопросы считаются как неотвеченные с полным временем

---

## Архитектура

```
┌─────────────────────────────────────────────────────────────────┐
│                         Telegram Users                          │
└─────────────┬──────────────────────────┬───────────────────────┘
              │                          │
              ▼                          ▼
        ┌──────────┐              ┌──────────┐
        │ Teacher  │              │ Students │
        └────┬─────┘              └────┬─────┘
             │                         │
             └───────────┬─────────────┘
                         │
                         ▼
                ┌────────────────┐
                │  Telegram Bot  │
                │   (main.go)    │
                └────────┬───────┘
                         │
        ┌────────────────┼────────────────┐
        ▼                ▼                ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│   Telegram   │ │  QuizEngine  │ │   Storage    │
│    Client    │ │              │ │  Interface   │
│ (HTTP API)   │ │  (Управление │ │              │
└──────────────┘ │   квизами)   │ └───────┬──────┘
                 └───────┬──────┘         │
                         │                ▼
                         │         ┌──────────────┐
                         │         │   In-Memory  │
                         │         │   Storage    │
                         │         │  (map+mutex) │
                         │         └──────────────┘
                         ▼
                ┌──────────────┐
                │  QuizEvent   │
                │   Channel    │
                │ (горутины и  │
                │   таймеры)   │
                └──────────────┘
```

---

## CSV формат результатов

```csv
Rank,TelegramID,Username,FirstName,LastName,Score,CorrectCount,TotalTime
1,123456789,alice,Alice,Smith,10,5,45s
2,987654321,bob,Bob,Johnson,9,4,52s
...
```

---

Удачи! 🚀
