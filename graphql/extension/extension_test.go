package extension

import (
	"context"
	_ "embed"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/executor"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

var (
	//go:embed test/schema.graphqls
	schema string
	//go:embed test/queries.graphql
	queries string
)

func TestRecursionLimitByTypeAndField(t *testing.T) {
	tests := []struct {
		operationName string
		expectedErr   gqlerror.List
	}{
		{
			operationName: "Allowed",
			expectedErr:   nil,
		},
		{
			operationName: "RecursionExceeded",
			expectedErr: gqlerror.List{{
				Message: "too many nesting on User.friends",
			}},
		},
		{
			operationName: "InterleavedTypesAllowed",
			expectedErr:   nil,
		},
		{
			operationName: "InterleavedTypesRecursionExceeded",
			expectedErr: gqlerror.List{{
				Message: "too many nesting on User.items",
			}},
		},

		{
			operationName: "DifferentSubtreeAllowed",
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.operationName, func(t *testing.T) {
			exec := executor.New(executableSchema{})
			exec.Use(RecursionLimitByTypeAndField(1))
			ctx := context.Background()
			ctx = graphql.StartOperationTrace(ctx)
			_, err := exec.CreateOperationContext(ctx, &graphql.RawParams{
				Query:         queries,
				OperationName: tt.operationName,
			})
			require.Equal(t, err, tt.expectedErr)
		})
	}
}

var sources = []*ast.Source{
	{Name: "schema.graphqls", Input: schema, BuiltIn: false},
}
var parsedSchema = gqlparser.MustLoadSchema(sources...)

var _ graphql.ExecutableSchema = &executableSchema{}

type executableSchema struct{}

func (e executableSchema) Schema() *ast.Schema {
	return parsedSchema
}

func (e executableSchema) Complexity(_, _ string, _ int, _ map[string]interface{}) (int, bool) {
	return 0, false
}

func (e executableSchema) Exec(_ context.Context) graphql.ResponseHandler {
	return func(ctx context.Context) *graphql.Response {
		return &graphql.Response{}
	}
}
