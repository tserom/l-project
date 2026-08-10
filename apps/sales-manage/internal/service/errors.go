package service

import "errors"

var (
	ErrValidation          = errors.New("validation error")
	ErrOrderNotFound       = errors.New("销售单不存在")
	ErrInvoiceNotFound     = errors.New("开票单不存在")
	ErrOrderHasInvoice     = errors.New("销售单已有开票明细，不能删除")
	ErrInvoiceAlreadyVoid  = errors.New("开票单已作废")
	ErrInvoiceLinesEmpty   = errors.New("至少保留一行明细")
)
