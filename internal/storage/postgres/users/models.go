package users

import "time"

type UserModel struct {
	ID           int
	Username     string
	FullName     string
	Role         string
	StudentGroup string
	CreatedAt    time.Time
}
