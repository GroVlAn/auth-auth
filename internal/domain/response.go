package domain

type Response struct {
	Error    *ErrorResponse `json:"error,omitempty"`
	Response interface{}    `json:"response,omitempty"`
	Data     interface{}    `json:"data,omitempty"`
}

type ErrorResponse struct {
	Code int    `json:"code,omitempty"`
	Text string `json:"text,omitempty"`
}
