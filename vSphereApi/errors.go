package vSphereApi

type Errors int

func NewError(e int) error {
	if e == 0 {
		return nil
	}
	return Errors(e)
}

// Error returns the error string for the Errors type.
func (e Errors) Error() string {
	return Flags[e]
}

func (e Errors) ErrorCode() int {
	return int(e)
}

const (
	FAILURE              = -1
	SUCCESS              = 1
	ERROR_OBJECT_IS_NULL = 2
	ERROR_NOT_FOUND      = 3
)

var Flags = map[Errors]string{
	FAILURE:              "failure.",
	SUCCESS:              "successful.",
	ERROR_OBJECT_IS_NULL: "Object is nil.",
	ERROR_NOT_FOUND:      "Object not found.",
}
