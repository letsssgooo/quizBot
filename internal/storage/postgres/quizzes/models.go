package quizzes

import "time"

type Info struct {
	ID             int
	Name           string
	File           []byte
	CreatedAt      time.Time
	AuthorUsername string
}

type Statistic struct {
	ID        int
	QuizID    int
	Username  string
	Questions []string
	Options   []string
	Answers   []string
	Points    int
	MaxPoints int
}
