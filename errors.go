package circuitbreaker

import (
	"errors"
)

// OnlyTheseErrors used when you want to consider only these errors.
func OnlyTheseErrors(err error, errs ...error) error {
	if err == nil {
		return nil
	}

	for _, er := range errs {
		if errors.Is(err, er) {
			return err
		}
	}

	return nil
}
