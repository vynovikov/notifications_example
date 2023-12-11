package authorizer

import "github.com/vynovikov/study/notifications_example/internal/pkg/model"

type Authorizer interface {
	Internal(model.WrappedReq) error
	External(model.WrappedReq) error
}

// Authorizer implementation
