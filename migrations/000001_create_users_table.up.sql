-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL UNIQUE,
    full_name VARCHAR(150),
    role VARCHAR(20),
    user_group VARCHAR,
    created_at TIMESTAMP NOT NULL
);

-- Создание таблицы с инфой о квизах
CREATE TABLE IF NOT EXISTS quizzes_info (
    id SERIAL PRIMARY KEY,
    name VARCHAR(250) NOT NULL UNIQUE, -- название квиза должно быть уникальным
    file JSONB NOT NULL, -- можно получать и изменять json, не выгружая его полностью
    owner_id BIGINT NOT NULL, -- username преподавателя
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS runs_history (
    id SERIAL PRIMARY KEY,
    name VARCHAR(250) NOT NULL UNIQUE,
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP
)

-- Создание таблицы со статистикой квизов
CREATE TABLE IF NOT EXISTS quizzes_statistic (
    id SERIAL PRIMARY KEY,
    quiz_id INTEGER NOT NULL REFERENCES quizzes_info(id), -- для связи с таблицей info
    student_id INTEGER NOT NULL REFERENCES users(id), -- для связи с таблицей users
    run_id INTEGER NOT NULL REFERENCES runs_history(id), -- для связи с таблицей runs_history
    score INTEGER
);