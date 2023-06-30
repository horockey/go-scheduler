package scheduler

import "github.com/horockey/go-scheduler/internal/model"

func EventHeaderID() string {
	return model.HeaderID
}

func EventHeaderCreatedAt() string {
	return model.HeaderCreatedAt
}
