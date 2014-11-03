package mock

import (
	"fmt"
)

// tTesting contains several of the testing package's actual methods
type tTesting struct {
	ErrorMsg string
	FatalMsg string
}

func (t *tTesting) Error(s ...interface{}) {
	t.ErrorMsg = fmt.Sprint(s...)
}

func (t *tTesting) Errorf(pat string, s ...interface{}) {
	t.ErrorMsg = fmt.Sprintf(pat, s...)
}

func (t *tTesting) Fatal(s ...interface{}) {
	t.FatalMsg = fmt.Sprint(s...)
}

func (t *tTesting) Fatalf(pat string, s ...interface{}) {
	t.FatalMsg = fmt.Sprintf(pat, s...)
}
