package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	accountBalanceGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "account_balance_cents",
			Help: "Current account balance in cents",
		},
		[]string{"account_id", "currency"},
	)

	transactionAmountGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "transaction_amount_cents",
			Help: "Transaction amount in cents",
		},
		[]string{"transaction_id", "type", "status"},
	)
)

// UpdateAccountBalanceMetric updates the account balance metric
func UpdateAccountBalanceMetric(accountID, currency string, balanceCents int64) {
	accountBalanceGauge.WithLabelValues(accountID, currency).Set(float64(balanceCents))
}

// UpdateTransactionMetric updates the transaction metric
func UpdateTransactionMetric(transactionID, txType, status string, amountCents int64) {
	transactionAmountGauge.WithLabelValues(transactionID, txType, status).Set(float64(amountCents))
}

