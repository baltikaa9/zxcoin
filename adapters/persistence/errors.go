package persistence

import "fmt"

type InvalidHeightError struct {
	Got, Want uint64
}

func (e InvalidHeightError) Error() string {
	return fmt.Sprintf("invalid block height: got %d, want %d", e.Got, e.Want)
}
