package store

import (
	"net/url"

	"github.com/google/uuid"
	"github.com/vynovikov/study/notifications_example/internal/pkg/model"
)

type Store interface {
	Read(uuid.UUID, url.Values) ([]map[string]interface{}, error)
	Write([]model.NotificationDataStructured, uuid.UUID) error
	Count(uuid.UUID, url.Values) (int, error)
	Close() error
}

// Store implementation
