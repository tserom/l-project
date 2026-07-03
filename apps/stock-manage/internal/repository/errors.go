package repository

import "errors"

var (
	ErrInboundOrderNotFound   = errors.New("inbound order not found")
	ErrOutboundOrderNotFound  = errors.New("outbound order not found")
	ErrSalesOrderNotFound     = errors.New("sales order not found")
	ErrSalesShipmentNotFound    = errors.New("sales shipment not found")
	ErrProcessingOrderNotFound  = errors.New("processing order not found")
	ErrNotDraft                 = errors.New("document is not draft")
)
