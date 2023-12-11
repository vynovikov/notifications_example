package application

import "github.com/vynovikov/study/notifications_example/internal/pkg/model"

type Application interface {
	Save(model.WrappedReq) error
	Extract(model.WrappedReq) ([][]byte, error)
	Count(model.WrappedReq) (int, error)
	AuthInternal(model.WrappedReq) error
	AuthExternal(model.WrappedReq) error
	Start()
	Stop()
	Log(model.UUIDWrapper, string)
}

// Receiver implementation
