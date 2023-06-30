package scheduler

import "github.com/horockey/go-scheduler/internal/model"

// Get standart header for event ID.
func EventHeaderID() string {
	return model.HeaderID
}

// Get standart header for event creation time.
func EventHeaderCreatedAt() string {
	return model.HeaderCreatedAt
}
