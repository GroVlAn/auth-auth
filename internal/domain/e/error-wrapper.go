package e

type ErrorType int

const (
	ErrorTypeNotFound ErrorType = iota
	ErrorTypeConflict
	ErrorTypeValidation
	ErrorTypeUnauthorized
	ErrorTypeInternal
)

type ValidationJSON struct {
	ErrorType string      `json:"error_type"`
	Message   string      `json:"message"`
	Fields    [][2]string `json:"fields,omitempty"`
}

type ErrWrapper struct {
	errType ErrorType
	err     error
	msg     string
}

func NewErrNotFound(err error, msg string) *ErrWrapper {
	return &ErrWrapper{
		errType: ErrorTypeNotFound,
		err:     err,
		msg:     msg,
	}
}

func NewErrConflict(err error, msg string) *ErrWrapper {
	return &ErrWrapper{
		errType: ErrorTypeConflict,
		err:     err,
		msg:     msg,
	}
}

func NewErrUnauthorized(err error, msg string) *ErrWrapper {
	return &ErrWrapper{
		errType: ErrorTypeUnauthorized,
		err:     err,
		msg:     msg,
	}
}

func NewErrInternal(err error) *ErrWrapper {
	return &ErrWrapper{
		errType: ErrorTypeInternal,
		err:     err,
		msg:     "internal server error",
	}
}

func (ew *ErrWrapper) Error() string {
	return ew.msg
}

func (ew *ErrWrapper) Unwrap() error {
	return ew.err
}

func (ew *ErrWrapper) ErrorType() ErrorType {
	return ew.errType
}
