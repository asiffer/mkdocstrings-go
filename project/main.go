package project

// CONSTANT_VALUE is a constant string used for demonstration purposes.
const CONSTANT_VALUE = "constant"

// ErrorCode is a custom type representing error codes in the application.
type ErrorCode int

const (
	// ErrNotFound indicates that a requested resource was not found.
	ErrNotFound ErrorCode = -iota - 1
	ErrInvalid
	// You do not have permission to access this resource.
	ErrPermissionDenied
	// ???
	ErrUnknown
)

var (
	// BasicMap is a simple map of strings to strings.
	BasicMap = map[string]string{}
	// BasicSlice is a simple slice of strings.
	BasicSlice = []string{}
)

// SmallStruct is a simple struct with two fields.
type SmallStruct struct {
	// Field1 is the first field of SmallStruct
	Field1 string `json:"field1" validate:"required"`
	Field2 int
}

// Method is a method of SmallStruct that returns a string.
func (s *SmallStruct) Method() string {
	return "This is a method of SmallStruct."
}

func unexportedFunction() string {
	return "This is an unexported function."
}

// ExportedFunction is an example of an exported
// function that can be accessed from other packages.
func ExportedFunction() string {
	return "This is an exported function."
}
