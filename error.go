package gomqtt

import (
	"fmt"
)

// ErrorDecodeMore
type ErrorDecodeMore struct{}

func (e ErrorDecodeMore) Error() string {
	return fmt.Sprintf("ErrorDecodeMore{}")
}
