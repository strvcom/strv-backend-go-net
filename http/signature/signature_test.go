package signature_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httpparam "go.strv.io/net/http/param"
	"go.strv.io/net/http/signature"
)

type User struct {
	UserName string `json:"user_name"`
	Group    int    `json:"group"`
}

type ListUsersInput struct {
	Group   int `param:"path=group"`
	Page    int `param:"query=page"`
	PerPage int `param:"query=per_page"`
}

type CreateUserInput struct {
	UserName string `json:"user_name"`
	Group    int    `json:"group"`
}

func hasStructJSONTag(obj any) bool {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	t := v.Type()
	for i := range t.NumField() {
		_, exists := t.Field(i).Tag.Lookup("json")
		if exists {
			return true
		}
	}
	return false
}

func parseInputFunc(r *http.Request, dest any) error {
	// Don't call json unmarshal if dest has no json tag, which means request body may be empty,
	// as only expected input are path and query parameters.
	// causes error if json.Unmarshal is called on empty body (EOF)
	if hasStructJSONTag(dest) {
		if err := signature.UnmarshalRequestBody(r, dest); err != nil {
			return err
		}
	}

	// After UnmarshalRequestBody, as it possibly fills all fields, even those tagged as query or path parameter.
	// This way, such filled fields will be reassigned in httpparam.Parser.
	return httpparam.DefaultParser().WithPathParamFunc(chi.URLParam).Parse(r, dest)
}

func TestWrapper(t *testing.T) {
	testCases := []struct {
		method         string
		url            string
		inputBody      string
		expectedBody   string
		expectedStatus int
	}{
		{
			method:         http.MethodGet,
			url:            "https://test.com/healthcheck",
			inputBody:      "",
			expectedBody:   "",
			expectedStatus: http.StatusNoContent,
		},
		{
			method:         http.MethodGet,
			url:            "https://test.com/dependency-check",
			inputBody:      "",
			expectedBody:   `{"payment-provider":"ready","company-registry":"unreachable"}`,
			expectedStatus: http.StatusOK,
		},
		{
			method:         http.MethodGet,
			url:            "https://test.com/group/55/users",
			inputBody:      "",
			expectedBody:   `[{"user_name":"Testowic","group":55}]`,
			expectedStatus: http.StatusOK,
		},
		{
			method:         http.MethodGet,
			url:            "https://test.com/users",
			inputBody:      "",
			expectedBody:   `[{"user_name":"Testowic","group":0}]`,
			expectedStatus: http.StatusOK,
		},
		{
			method:         http.MethodPost,
			url:            "https://test.com/users",
			inputBody:      `{"user_name":"NewUser","group":5}`,
			expectedBody:   "",
			expectedStatus: http.StatusCreated,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.url, func(t *testing.T) {
			w := signature.DefaultWrapper().
				WithInputGetter(parseInputFunc)

			r := chi.NewRouter()

			r.Get("/healthcheck", signature.WrapHandlerError(w, healthcheckHandler))
			r.Get("/dependency-check", signature.WrapHandlerResponse(w, dependencyCheckHandler))
			r.Get("/group/{group}/users", signature.WrapHandler(w, listUsersHandler))
			r.Route("/users", func(r chi.Router) {
				r.Get("/", signature.WrapHandler(w, listUsersHandler))
				r.Post("/", signature.WrapHandlerInput(
					w.WithResponseMarshaler(signature.FixedResponseCodeMarshal(http.StatusCreated)),
					createUserHandler,
				))
			})

			var body io.Reader
			if tc.inputBody != "" {
				body = strings.NewReader(tc.inputBody)
			}
			req := httptest.NewRequest(tc.method, tc.url, body)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rec.Body.String())
			} else {
				assert.Nil(t, rec.Body.Bytes())
			}
		})
	}
}

func listUsersHandler(_ http.ResponseWriter, _ *http.Request, input ListUsersInput) ([]User, error) {
	return []User{{
		UserName: "Testowic",
		Group:    input.Group,
	}}, nil
}

func createUserHandler(_ http.ResponseWriter, _ *http.Request, _ CreateUserInput) error {
	return nil
}

func healthcheckHandler(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

func dependencyCheckHandler(_ http.ResponseWriter, _ *http.Request) (map[string]DependencyStatus, error) {
	return map[string]DependencyStatus{
		"payment-provider": DependencyStatusReady,
		"company-registry": DependencyStatusUnreachable,
	}, nil
}

type DependencyStatus int

const (
	DependencyStatusReady DependencyStatus = iota + 1
	DependencyStatusUnreachable
)

func (d DependencyStatus) MarshalText() ([]byte, error) {
	var name string
	switch d {
	case DependencyStatusReady:
		name = "ready"
	case DependencyStatusUnreachable:
		name = "unreachable"
	default:
		return nil, fmt.Errorf("invalid DependencyStatus value (%d) when marshaling", d)
	}
	return []byte(name), nil
}

func TestWrapper_Error(t *testing.T) {
	w := signature.DefaultWrapper()

	var interceptedError error
	w = w.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
		interceptedError = err
		signature.InputGetErrorHandle(w, r, err)
	})

	testCases := []struct {
		name           string
		inputBody      string
		expectedBody   string
		expectedStatus int
		handler        http.Handler
		targetErr      error
		isABug         bool
	}{
		{
			name:           "parsing body returns 400",
			inputBody:      `{"incomplete_json":`,
			expectedStatus: http.StatusBadRequest,
			handler:        signature.WrapHandler(w, buggyHandler),
			targetErr:      signature.ErrInputGet,
			isABug:         false,
		},
		{
			name:           "internal handler error returns 500",
			inputBody:      `{"bug":true}`,
			expectedStatus: http.StatusInternalServerError,
			handler:        signature.WrapHandler(w, buggyHandler),
			targetErr:      signature.ErrInnerHandler,
			isABug:         true,
		},
		{
			name:      "marshaling error returns 500? Well actually 200",
			inputBody: `{"bug":false}`,
			// Header was already written at the time of marshal error, there is no way to change it.
			// The only way would be to unmarshal the object into buffer to see if it returns error.
			//
			// This is behaviour of httptest.ResponseRecorder.WriteHeader(), but the http.response.WriteHeader()
			// behaves the same way. I guess it's not good to have TextMarshalers that can error on valid ResponseWriter
			expectedStatus: http.StatusOK,
			handler:        signature.WrapHandler(w, buggyHandler),
			targetErr:      signature.ErrResponseMarshal,
			isABug:         true,
		},
		{
			name:           "parsing body returns 400 (only input)",
			inputBody:      `{"incomplete_json":`,
			expectedStatus: http.StatusBadRequest,
			handler:        signature.WrapHandlerInput(w, buggyHandlerInput),
			targetErr:      signature.ErrInputGet,
			isABug:         false,
		},
		{
			name:           "internal handler error returns 500 (only input)",
			inputBody:      `{"bug":true}`,
			expectedStatus: http.StatusInternalServerError,
			handler:        signature.WrapHandlerInput(w, buggyHandlerInput),
			targetErr:      signature.ErrInnerHandler,
			isABug:         true,
		},
		{
			name:           "internal handler error returns 500 (only response)",
			inputBody:      "",
			expectedStatus: http.StatusInternalServerError,
			handler:        signature.WrapHandlerResponse(w, buggyHandlerResponse),
			targetErr:      signature.ErrInnerHandler,
			isABug:         true,
		},
		{
			name:           "marshaling error returns 200 (only response)",
			inputBody:      "",
			expectedStatus: http.StatusOK, // same as above problem
			handler:        signature.WrapHandlerResponse(w, buggyHandlerBuggyResponse),
			targetErr:      signature.ErrResponseMarshal,
			isABug:         true,
		},
		{
			name:           "internal handler error returns 500 (only error)",
			inputBody:      "",
			expectedStatus: http.StatusInternalServerError,
			handler:        signature.WrapHandlerError(w, buggyHandlerError),
			targetErr:      signature.ErrInnerHandler,
			isABug:         true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.inputBody != "" {
				body = strings.NewReader(tc.inputBody)
			}
			req := httptest.NewRequest(http.MethodGet, "https://test.com/error", body)
			rec := httptest.NewRecorder()

			tc.handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			require.ErrorIs(t, interceptedError, tc.targetErr)
			if tc.isABug {
				require.ErrorIs(t, interceptedError, errBug)
			}
			assert.JSONEq(t, `{"errorCode":"ERR_UNKNOWN"}`, rec.Body.String())
		})
	}
}

var errBug = errors.New("that's a bug, and should propagate properly")

type willBug bool

func (w willBug) MarshalText() (text []byte, err error) {
	return nil, errBug
}

type buggyInputOrOutput struct {
	WillBug *willBug `json:"bug"`
}

func buggyHandler(_ http.ResponseWriter, _ *http.Request, input buggyInputOrOutput) (*buggyInputOrOutput, error) {
	if input.WillBug != nil && *input.WillBug {
		return nil, errBug
	}
	x := willBug(true)
	return &buggyInputOrOutput{WillBug: &x}, nil
}

func buggyHandlerResponse(_ http.ResponseWriter, _ *http.Request) (*buggyInputOrOutput, error) {
	return nil, errBug
}

func buggyHandlerBuggyResponse(_ http.ResponseWriter, _ *http.Request) (*buggyInputOrOutput, error) {
	x := willBug(true)
	return &buggyInputOrOutput{WillBug: &x}, nil
}

func buggyHandlerInput(_ http.ResponseWriter, _ *http.Request, input buggyInputOrOutput) error {
	if input.WillBug != nil && *input.WillBug {
		return errBug
	}
	return nil
}

func buggyHandlerError(_ http.ResponseWriter, _ *http.Request) error {
	return errBug
}
