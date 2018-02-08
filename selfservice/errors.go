package selfservice

import (
	"fmt"
	"time"
)

type SamadError struct {
	When time.Time
	What string
}

func (e SamadError) Error() string {
	return fmt.Sprintf("%v: %v", e.When, e.What)
}
