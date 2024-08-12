package filters

import (
	"encoding/json"
)

type CallbackFn func(data any) any

type ErrorResponse struct {
	ErrorMsg    string `json:"error_msg"`
	ErrorAction string `json:"error_action"` // warning message, restrict, restrict + warning message
}

func (e *ErrorResponse) Error() string {
	bt, _ := json.Marshal(e)
	return string(bt)
}

func (r *Rule) SetCallback(handler CallbackFn) {
	r.callback = handler
}

func (r *Rule) SetErrorResponse(msg, action string) {
	r.errorResponse = ErrorResponse{ErrorMsg: msg, ErrorAction: action}
}

func (r *Rule) Apply(data any, callback ...CallbackFn) (any, error) {
	var defaultCallbackFn CallbackFn
	if r.callback != nil {
		defaultCallbackFn = r.callback
	}
	if len(callback) > 0 {
		defaultCallbackFn = callback[0]
	}
	if defaultCallbackFn == nil {
		defaultCallbackFn = func(data any) any {
			return data
		}
	}
	matched := r.Match(data)
	var err error
	if !matched {
		err = &ErrorResponse{ErrorMsg: r.errorResponse.ErrorMsg, ErrorAction: r.errorResponse.ErrorAction}
	}
	return defaultCallbackFn(data), err
}
