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
    creator VARCHAR(32) NOT NULL, -- username преподавателя
    created_at TIMESTAMP NOT NULL
);

-- Создание таблицы со статистикой квизов
CREATE TABLE IF NOT EXISTS quizzes_statistic (
    id SERIAL PRIMARY KEY,
    quizID INTEGER NOT NULL REFERENCES quizzes_info(id), -- для связи с таблицей info
    answers VARCHAR(1)[] NOT NULL,
    score INTEGER NOT NULL,
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP NOT NULL
);