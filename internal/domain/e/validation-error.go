package e

type ErrValidation struct {
	ErrWrapper
	fields [][2]string
}

func NewErrValidation(msg string) *ErrValidation {
	return &ErrValidation{
		ErrWrapper: ErrWrapper{
			errType: ErrorTypeValidation,
			msg:     msg,
		},
		fields: make([][2]string, 0),
	}
}

func (ev *ErrValidation) AddField(field, reason string) {
	ev.fields = append(ev.fields, [2]string{field, reason})
}

func (ev *ErrValidation) Error() string {
	return ev.msg
}

func (ev *ErrValidation) Data() ValidationJSON {
	return ValidationJSON{
		ErrorType: "validation_error",
		Message:   ev.msg,
		Fields:    ev.fields,
	}
}

func (ev *ErrValidation) FirstError() (field, reason string, ok bool) {
	if len(ev.fields) == 0 {
		return "", "", false
	}

	return ev.fields[0][0], ev.fields[0][1], true
}

func (ev *ErrValidation) IsEmpty() bool {
	return len(ev.fields) == 0
}
