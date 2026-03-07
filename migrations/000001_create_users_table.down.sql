DROP TRIGGER IF EXISTS trg_check_run_quiz ON quizzes_statistic;
DROP FUNCTION IF EXISTS check_statistic_run_quiz();

DROP TABLE IF EXISTS quizzes_statistic;
DROP TABLE IF EXISTS runs_history;
DROP TABLE IF EXISTS quizzes_info;
DROP TABLE IF EXISTS users;