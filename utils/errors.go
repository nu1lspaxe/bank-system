package utils

import "fmt"

const (
	// simulator
	ErrGenerateNoContent = iota + 1
	ErrRequest

	ErrInvalidTransactionType
	ErrInvalidAmount

	// user
	ErrEmailExists
	// account
	ErrInsufficientBalance
	ErrAccountNotFound
)

type BankSystemError struct {
	code    int
	message string
}

func (e *BankSystemError) Error() string {
	return e.message
}

func NewBankSystemError(code int, opts ...string) *BankSystemError {
	message := GetErrorMessage(code, opts...)
	return &BankSystemError{code: code, message: message}
}

func GetErrorMessage(code int, opts ...string) string {
	switch code {
	case ErrGenerateNoContent:
		return fmt.Sprintf("failed to generate content: %v", opts)
	case ErrRequest:
		return fmt.Sprintf("request failed: %v", opts)
	case ErrInvalidTransactionType:
		return fmt.Sprintf("invalid transaction type: %v", opts)
	case ErrInvalidAmount:
		return fmt.Sprintf("invalid amount: %v", opts)
	case ErrEmailExists:
		return fmt.Sprintf("email already exists: %v", opts)
	case ErrInsufficientBalance:
		return fmt.Sprintf("insufficient balance: %v", opts)
	case ErrAccountNotFound:
		return fmt.Sprintf("account not found: %v", opts)
	default:
		return "unknown error"
	}
}
