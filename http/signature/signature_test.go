package signature_test

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-chi/chi/v5"

	httpparam "go.strv.io/net/http/param"
	"go.strv.io/net/http/signature"
)

type User struct {
	UserName string `json:"user_name"`
	Group    int    `json:"group"`
}

type ListUsersInput struct {
	Group   int `pathparam:"group"`
	Page    int `queryparam:"page"`
	PerPage int `queryparam:"per_page"`
}

type CreateUserInput struct {
	UserName string `json:"user_name"`
	Group    int    `json:"group"`
}

func hasStructJsonTag(obj any) bool {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
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
	if hasStructJsonTag(dest) {
		if err := signature.UnmarshalRequestBody(r, dest); err != nil {
			return err
		}
	}

	// After UnmarshalRequestBody, as it possibly fills all fields, even those tagged as query or path parameter.
	// This way, such filled fields will be reassigned in httpparam.ParamParser.
	return httpparam.DefaultParamParser().WithPathParamFunc(chi.URLParam).Parse(r, dest)
}

// TODO change from main to some tests
func main() {
	w := signature.DefaultWrapper().
		WithInputGetter(parseInputFunc).
		WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
			fmt.Println(err)
			signature.InputGetErrorHandle(w, r, err)
		})

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

	_ = http.ListenAndServe(":3000", r)
}

func listUsersHandler(_ http.ResponseWriter, _ *http.Request, input ListUsersInput) ([]User, error) {
	fmt.Println(input)
	return []User{{
		UserName: "Testowic",
		Group:    input.Group,
	}}, nil
}

func createUserHandler(_ http.ResponseWriter, _ *http.Request, input CreateUserInput) error {
	fmt.Println(input)
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
