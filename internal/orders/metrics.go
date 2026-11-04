package orders

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

var OrdersCreatedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "orders_created_total",
	Help: "Total number of order creation attempts by result and failure reason",
}, []string{"result", "reason"})

const (
	resultSuccess = "success"
	resultFailure = "failure"

	reasonInvalidBody       = "invalid_body"
	reasonValidation        = "validation"
	reasonNotFound          = "not_found"
	reasonInsufficientStock = "insufficient_stock"
	reasonInternal          = "internal"
)

func recordOrderCreated(err error) {
	if err == nil {
		OrdersCreatedTotal.WithLabelValues(resultSuccess, "").Inc()
		return
	}
	OrdersCreatedTotal.WithLabelValues(resultFailure, failureReason(err)).Inc()
}

func failureReason(err error) string {
	switch {
	case errors.Is(err, ErrCustomerNameRequired),
		errors.Is(err, ErrCustomerPhoneRequired),
		errors.Is(err, ErrItemsRequired),
		errors.Is(err, ErrInvalidQuantity),
		errors.Is(err, ErrDuplicateVariant):
		return reasonValidation
	case errors.Is(err, ErrVariantNotFound):
		return reasonNotFound
	case errors.Is(err, ErrInsufficientStock):
		return reasonInsufficientStock
	default:
		return reasonInternal
	}
}
