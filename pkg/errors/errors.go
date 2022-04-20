package errors

// Used to indicate handled failure
type SimulationError struct {
	Err string
}

func (e *SimulationError) Error() string {
	return e.Err
}
