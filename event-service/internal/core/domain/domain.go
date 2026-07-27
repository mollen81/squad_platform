package domain

import "time"

type Event struct {
	ID int64
	Name string
	UserCreateID int64
	UserCount int64
	TimeStart time.Time
	CreateTime time.Time
}
