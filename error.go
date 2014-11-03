package mock

import (
	"fmt"
)

type UnmockedError struct {
	Method string
	URL    string
}

func (e UnmockedError) Error() string {
	return fmt.Sprintf("mock error: called to unmocked URL: [%s] %s", e.Method,
		e.URL)
}
