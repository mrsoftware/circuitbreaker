package circuitbreaker_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mrsoftware/circuitbreaker"
	"github.com/stretchr/testify/assert"
)

func TestOnlyTheseErrors(t *testing.T) {
	err1 := errors.New("error1")
	err2 := errors.New("error2")
	err3 := errors.New("error3")
	err4 := errors.New("error4")
	err5 := errors.New("error5")

	testCases := []struct {
		Error    error
		List     []error
		Expected error
	}{
		{Error: err1, List: []error{err1, err2, err3}, Expected: err1},
		{Error: err1, List: []error{err4, err2, err5}, Expected: nil},
		{Error: nil, List: []error{err4, err2, err3, err4}, Expected: nil},
		{Error: err2, List: []error{err4, err2, err3, err4}, Expected: err2},
	}

	for index, item := range testCases {
		item := item
		t.Run(fmt.Sprintf("running test %d", index), func(t *testing.T) {
			assert.Equal(t, item.Expected, circuitbreaker.OnlyTheseErrors(item.Error, item.List...))
		})
	}
}
