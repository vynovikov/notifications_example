package saver

import "github.com/vynovikov/study/notifications_example/internal/pkg/model"

type Saver interface {
	Save(model.WrappedLog) error
}

// Saver implementation
