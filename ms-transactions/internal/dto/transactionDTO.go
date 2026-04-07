package dto

type TransactionOutput struct {
	ID     string  `json:"id"`
	UserID string  `json:"user_id" validate:"required"`
	Type   string  `json:"type" validate:"required,oneof=CREDIT DEBIT"`
	Amount float64 `json:"amount" validate:"required,gt=0"`
}
