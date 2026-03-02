package models

import (
	"time"
)

// Файл для работы с моделями для базы данных, которые доступны извне.
// Обработчики создают экземляры моделей, заполняют их данными и
// передают в соответсвующую функцию в БД.

// UserModel определяет модель для таблицы пользователей
type UserModel struct {
	ID         int
	TelegramID int64
	FullName   string
	Role       string
	Group      string
	CreatedAt  time.Time
}

// InfoModel определяет модель для таблицы с информацией о квизах
type InfoModel struct {
	ID             int
	Name           string
	File           []byte
	CreatedAt      time.Time
	OwnerID int64
}

// StatisticModel определяет модель для таблицы с результатами квизов
type StatisticModel struct {
	ID        int
	QuizID    int
	StudentID  int
	Answers   []string
	Score    int
	StartedAt time.Time
	FinishedAt time.Time
}
