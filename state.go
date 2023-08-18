package circuitbreaker

// State is circuit state.
type state int64

const (
	// StateClose mean circuit is close and can accept request.
	StateClose state = iota

	// StateOpen mean circuit is open and can not accept request.
	StateOpen

	// StateHalfOpen mean circuit is half open and can accept request.
	StateHalfOpen
)

const (
	stateCloseText    = "Close"
	stateOpenText     = "Open"
	stateHalfOpenText = "HalfOpen"
	stateNotValidText = "NotValid"
)

// GetStateText of circuit breaker.
func GetStateText(state state) string {
	switch state {
	case StateClose:
		return stateCloseText
	case StateOpen:
		return stateOpenText
	case StateHalfOpen:
		return stateHalfOpenText
	}

	return stateNotValidText
}
