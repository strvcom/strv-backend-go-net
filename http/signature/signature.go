package signature

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	httpx "go.strv.io/net/http"
)

var (
	// ErrInputGet is passed to ErrorHandlerFunc when WrapHandler (or derived) fails in the first step (parsing input)
	ErrInputGet = errors.New("parsing input")
	// ErrInnerHandler is passed to ErrorHandlerFunc when WrapHandler (or derived) fails in the second step (inner handler)
	ErrInnerHandler = errors.New("inner handler")
	// ErrResponseMarshal is passed to ErrorHandlerFunc when WrapHandler (or derived) fails in the third step (marshaling response object)
	ErrResponseMarshal = errors.New("marshaling response")
)

// InputGetterFunc is a function that is used in WrapHandler and WrapHandlerInput to parse request into declared input type.
// Before calling the inner handler, the InputGetterFunc is called to fill the struct that is then passed to the inner handler.
// If inner handler does not declare an input type (i.e. WrapHandlerResponse and WrapHandlerError), this function is not called at all.
type InputGetterFunc func(r *http.Request, dest any) error

// ResponseMarshalerFunc is a function that is used in WrapHandler and related functions to marshal declared response type.
// After the inner handler succeeds, the ResponseMarshalerFunc receives http.ResponseWriter and http.Request of handled request,
// and a type that an inner handler function declared as its first (non-error) return value.
// If the inner handler does not declare such return value (i.e. for WrapHandlerInput and WrapHandlerError),
// the ResponseMarshalerFunc receives http.NoBody as the src parameter.
type ResponseMarshalerFunc func(w http.ResponseWriter, r *http.Request, src any) error

// ErrorHandlerFunc is a function that is used in WrapHandler and related functions if any of the steps fail.
// The passed err is wrapped in one of ErrInputGet, ErrInnerHandler or ErrResponseMarshal to distinguish the
// step that failed.
//
// Note that if the error occurs on unmarshaling response with still valid http.ResponseWriter,
// and that step already wrote into the writer, the unmarshaled response (including e.g. http headers)
// may be inconsistent if error handler also writes.
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)

// Wrapper needs to be passed to WrapHandler and related functions. It contains the common handling of parsing http.Request
// to needed type, marshaling the response of needed type, and handling the errors that occur in any of those steps or
// in the inner handler (with modified signature)
type Wrapper struct {
	inputGetter       InputGetterFunc
	responseMarshaler ResponseMarshalerFunc
	errorHandler      ErrorHandlerFunc
}

// DefaultWrapper Creates a Wrapper with default functions for each needed step.
//
// Input is parsed only from http.Request body, using JSON unmarshal.
// A custom InputGetterFunc is needed to parse also the query and path parameters, but param package can be used to do most.
//
// Response is marshaled using a WriteResponse wrapper in parent package, which uses JSON marshal.
//
// Error handler also uses a WriteErrorResponse of parent package.
// It is recommended to replace this to implement any custom error handling (matching any application errors).
// Default handler only returns http code 400 on unmarshal error and 500 otherwise.
func DefaultWrapper() Wrapper {
	return Wrapper{
		inputGetter:       UnmarshalRequestBody,
		responseMarshaler: DefaultResponseMarshal,
		errorHandler:      InputGetErrorHandle,
	}
}

// WithInputGetter returns a copy of Wrapper with new InputGetterFunc
func (w Wrapper) WithInputGetter(f InputGetterFunc) Wrapper {
	w.inputGetter = f
	return w
}

// WithResponseMarshaler returns a copy of Wrapper with new ResponseMarshalerFunc
func (w Wrapper) WithResponseMarshaler(f ResponseMarshalerFunc) Wrapper {
	w.responseMarshaler = f
	return w
}

// WithErrorHandler returns a copy of Wrapper with new ErrorHandlerFunc
func (w Wrapper) WithErrorHandler(f ErrorHandlerFunc) Wrapper {
	w.errorHandler = f
	return w
}

func inputErrorWithType(target any, innerError error) error {
	return fmt.Errorf("%w into type %T: %w", ErrInputGet, target, innerError)
}

func responseErrorWithType(src any, innerError error) error {
	if src == nil {
		return fmt.Errorf("%w without response object: %w", ErrResponseMarshal, innerError)
	}
	return fmt.Errorf("%w from type %T: %w", ErrResponseMarshal, src, innerError)
}

func wrapInnerHandlerError(innerError error) error {
	return fmt.Errorf("%w: %w", ErrInnerHandler, innerError)
}

// WrapHandler enables a handler with signature of second parameter to be used as a http.HandlerFunc.
// 1. Before calling such inner handler, the http.request is used
// to get the input parameter of type TInput for the handler, using InputGetterFunc in Wrapper.
// 2. Then the inner handler is called with such created TInput.
// 3. If the handler succeeds (returns nil error), The first return value
// (of type TResponse) is passed to ResponseMarshalerFunc of Wrapper.
// If any of the above steps returns error, the ErrorHandlerFunc is called with that error.
func WrapHandler[TInput any, TResponse any](wrapper Wrapper, handler func(http.ResponseWriter, *http.Request, TInput) (TResponse, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input TInput
		err := wrapper.inputGetter(r, &input)
		if err != nil {
			wrapper.errorHandler(w, r, inputErrorWithType(input, err))
			return
		}
		response, err := handler(w, r, input)
		if err != nil {
			wrapper.errorHandler(w, r, wrapInnerHandlerError(err))
			return
		}
		err = wrapper.responseMarshaler(w, r, response)
		if err != nil {
			wrapper.errorHandler(w, r, responseErrorWithType(response, err))
			return
		}
	}
}

// WrapHandlerResponse enables a handler with signature of second parameter to be used as a http.HandlerFunc.
// See WrapHandler for general idea.
// Compared to WrapHandler, the first step is skipped (no parsed input for inner handler is provided)
func WrapHandlerResponse[TResponse any](wrapper Wrapper, handler func(http.ResponseWriter, *http.Request) (TResponse, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := handler(w, r)
		if err != nil {
			wrapper.errorHandler(w, r, wrapInnerHandlerError(err))
			return
		}
		err = wrapper.responseMarshaler(w, r, response)
		if err != nil {
			wrapper.errorHandler(w, r, responseErrorWithType(response, err))
			return
		}
	}
}

// WrapHandlerInput enables a handler with signature of second parameter to be used as a http.HandlerFunc.
// See WrapHandler for general idea.
// Compared to WrapHandler, in the last step, the ResponseMarshalerFunc receives http.NoBody as a response object
// (and as such, the ResponseMarshalerFunc should handle the http.NoBody value gracefully)
func WrapHandlerInput[TInput any](wrapper Wrapper, handler func(http.ResponseWriter, *http.Request, TInput) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input TInput
		err := wrapper.inputGetter(r, &input)
		if err != nil {
			wrapper.errorHandler(w, r, inputErrorWithType(input, err))
			return
		}
		err = handler(w, r, input)
		if err != nil {
			wrapper.errorHandler(w, r, wrapInnerHandlerError(err))
			return
		}
		err = wrapper.responseMarshaler(w, r, http.NoBody)
		if err != nil {
			wrapper.errorHandler(w, r, responseErrorWithType(nil, err))
			return
		}
	}
}

// WrapHandlerError enables a handler with signature of second parameter to be used as a http.HandlerFunc.
// See WrapHandler for general idea.
// Compared to WrapHandler, the first step is skipped (no parsed input for inner handler is provided),
// and in the last step, the ResponseMarshalerFunc receives http.NoBody as a response object
// (and as such, the ResponseMarshalerFunc should handle the http.NoBody value gracefully)
func WrapHandlerError(wrapper Wrapper, handler func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			wrapper.errorHandler(w, r, wrapInnerHandlerError(err))
			return
		}
		err = wrapper.responseMarshaler(w, r, http.NoBody)
		if err != nil {
			wrapper.errorHandler(w, r, responseErrorWithType(nil, err))
			return
		}
	}
}

// UnmarshalRequestBody decodes a body into a struct.
// This function expects the request body to be a JSON object and target to be a pointer to expected struct.
// If the request body is invalid, it returns an error.
func UnmarshalRequestBody(r *http.Request, target any) error {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return err
	}
	return nil
}

// FixedResponseCodeMarshal returns a ResponseMarshalerFunc that always writes provided http status code on success.
func FixedResponseCodeMarshal(statusCode int) ResponseMarshalerFunc {
	return func(w http.ResponseWriter, _ *http.Request, obj any) error {
		return httpx.WriteResponse(w, obj, statusCode)
	}
}

// DefaultResponseMarshal is a ResponseMarshalerFunc that writes 200 OK http status code with JSON marshaled object.
// 204 No Content http status code is returned if no response object is provided (i.e. when using WrapHandlerInput or WrapHandlerError)
func DefaultResponseMarshal(w http.ResponseWriter, _ *http.Request, src any) error {
	if src == http.NoBody {
		return httpx.WriteResponse(w, src, http.StatusNoContent)
	}
	return httpx.WriteResponse(w, src, http.StatusOK)
}

// AlwaysInternalErrorHandle is a function usable as ErrorHandlerFunc.
// It writes 500 http status code on error.
// Error message not returned in response and is lost.
func AlwaysInternalErrorHandle(w http.ResponseWriter, _ *http.Request, _ error) {
	_ = httpx.WriteErrorResponse(w, http.StatusInternalServerError)
}

// InputGetErrorHandle is a function usable as ErrorHandlerFunc.
// It writes a 400 Bad Request http status code to http.ResponseWriter if the error is from parsing input.
// Otherwise, writes 500 Internal Server Error http status code on error.
// In either case, error message is not returned in response and is lost
func InputGetErrorHandle(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, ErrInputGet) {
		_ = httpx.WriteErrorResponse(w, http.StatusBadRequest)
		return
	}
	AlwaysInternalErrorHandle(w, r, err)
}
