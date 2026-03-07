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
    owner_id BIGINT NOT NULL, -- user_telegram_id преподавателя
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS runs_history (
    id SERIAL PRIMARY KEY,
    quiz_id INTEGER NOT NULL REFERENCES quizzes_info(id),
    name VARCHAR(250) NOT NULL,
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP
);

-- Создание таблицы со статистикой квизов
CREATE TABLE IF NOT EXISTS quizzes_statistic (
    id SERIAL PRIMARY KEY,
    quiz_id INTEGER NOT NULL REFERENCES quizzes_info(id), -- для связи с таблицей info
    student_id BIGINT NOT NULL REFERENCES users(telegram_id), -- для связи с таблицей users
    run_id INTEGER NOT NULL REFERENCES runs_history(id), -- для связи с таблицей runs_history
    score INTEGER,
    UNIQUE (student_id, run_id)
);

CREATE OR REPLACE FUNCTION check_statistic_run_quiz()
RETURNS TRIGGER AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM runs_history
        WHERE id = NEW.run_id AND quiz_id = NEW.quiz_id
    ) THEN
        RAISE EXCEPTION 'run_id % does not belong to quiz_id %', NEW.run_id, NEW.quiz_id;
END IF;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_run_quiz
BEFORE INSERT OR UPDATE ON quizzes_statistic
FOR EACH ROW
EXECUTE FUNCTION check_statistic_run_quiz();