package e

import (
	"fmt"
	"testing"

	"gotest.tools/assert"
)

func TestMergeError(t *testing.T) {
	err1 := fmt.Errorf("err 1")
	err2 := fmt.Errorf("err 2")
	err3 := fmt.Errorf("err 3")
	err := MergeError([]error{err1, err2, err3})
	assert.ErrorContains(t, err, "err 1")
	assert.ErrorContains(t, err, "err 3")
	t.Log(err)

}
