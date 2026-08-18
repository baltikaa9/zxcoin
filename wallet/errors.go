package wallet

import "fmt"

type InsufficientFundsError struct {
	Available int
	Requested int
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("недостаточно средств: доступно %d, запрошено %d", e.Available, e.Requested)
}