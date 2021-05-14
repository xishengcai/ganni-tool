package e

import (
	"fmt"
	"testing"
)

func TestMergeError(t *testing.T) {
	err1 := fmt.Errorf("err 1")
	err2 := fmt.Errorf("err 2")
	err := MergeError([]error{err1, err2})
	t.Log(err.Error())
}
