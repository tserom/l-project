package repository

import "errors"

var (
	ErrInboundOrderNotFound  = errors.New("inbound order not found")
	ErrOutboundOrderNotFound = errors.New("outbound order not found")
	ErrNotDraft              = errors.New("document is not draft")
)
