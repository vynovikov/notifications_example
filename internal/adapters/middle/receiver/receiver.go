package receiver

import (
	"net/http"

	"github.com/vynovikov/study/notifications_example/internal/pkg/model"
)

type Receiver interface {
	HandlePut() http.HandlerFunc
	HandleGet() http.HandlerFunc
	HandleCount() http.HandlerFunc
	Log(model.UUIDWrapper, string)
	Start()
	Stop()
}

// Receiver implementation
