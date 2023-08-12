package circuitbreaker

// Circuit is a Circuit manager.
type Circuit struct {
	ops Options
}

// NewCircuit breaker.
func NewCircuit(options ...Option) Circuit {
	circuit := Circuit{ops: Options{}}

	for _, op := range options {
		op(&circuit.ops)
	}

	return circuit
}
