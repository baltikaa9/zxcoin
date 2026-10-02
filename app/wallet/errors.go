package wallet

import "fmt"

type InsufficientFundsError struct {
	Available uint64
	Requested uint64
}

func (e InsufficientFundsError) Error() string {
	return fmt.Sprintf("недостаточно средств: доступно %d, запрошено %d", e.Available, e.Requested)
}
