package transaction

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type (
	TransactionType string

	Transaction struct {
		UserID string          `json:"user_id" validate:"required"`
		Type   TransactionType `json:"type" validate:"required"`
		Amount float64         `json:"amount" validate:"required,gt=0"`
	}
)

var validate = validator.New()

const (
	Credit TransactionType = "credit"
	Debit  TransactionType = "debit"
)

func (t TransactionType) IsValid() bool {
	return t == Credit || t == Debit
}

func New(userID string, transactionType TransactionType, amount float64) (*Transaction, error) {
	t := &Transaction{
		UserID: userID,
		Type:   transactionType,
		Amount: amount,
	}
	if !transactionType.IsValid() {
		return nil, fmt.Errorf("invalid transaction type: %s", transactionType)
	}
	if err := validate.Struct(t); err != nil {
		return nil, err
	}
	return t, nil
}
