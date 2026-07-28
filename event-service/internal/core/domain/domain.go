package domain

import "time"

type Event struct {
	ID string
	Name string
	UserCreateID string
	UserCount int64
	TimeStart time.Time
	CreateTime time.Time
}
