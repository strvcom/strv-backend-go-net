package extension

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type RecursionLimit struct {
	maxRecursion int
}

func RecursionLimitByTypeAndField(limit int) *RecursionLimit {
	return &RecursionLimit{
		maxRecursion: limit,
	}
}

var _ interface {
	graphql.OperationContextMutator
	graphql.HandlerExtension
} = &RecursionLimit{}

func (r *RecursionLimit) ExtensionName() string {
	return "RecursionLimit"
}

func (r *RecursionLimit) Validate(_ graphql.ExecutableSchema) error {
	return nil
}

func (r *RecursionLimit) MutateOperationContext(_ context.Context, opCtx *graphql.OperationContext) *gqlerror.Error {
	return checkRecursionLimitByTypeAndField(recursionContext{
		maxRecursion:      r.maxRecursion,
		opCtx:             opCtx,
		typeAndFieldCount: map[nestingByTypeAndField]int{},
	}, string(opCtx.Operation.Operation), opCtx.Operation.SelectionSet)
}

type nestingByTypeAndField struct {
	parentTypeName string
	childFieldName string
}

type recursionContext struct {
	maxRecursion      int
	opCtx             *graphql.OperationContext
	typeAndFieldCount map[nestingByTypeAndField]int
}

func checkRecursionLimitByTypeAndField(rCtx recursionContext, typeName string, selectionSet ast.SelectionSet) *gqlerror.Error {
	if selectionSet == nil {
		return nil
	}

	collected := graphql.CollectFields(rCtx.opCtx, selectionSet, nil)
	for _, collectedField := range collected {
		nesting := nestingByTypeAndField{
			parentTypeName: typeName,
			childFieldName: collectedField.Name,
		}
		newCount := rCtx.typeAndFieldCount[nesting] + 1
		if newCount > rCtx.maxRecursion {
			return gqlerror.Errorf("too many nesting on %s.%s", nesting.parentTypeName, nesting.childFieldName)
		}
		rCtx.typeAndFieldCount[nesting] = newCount
		err := checkRecursionLimitByTypeAndField(rCtx, collectedField.Definition.Type.Name(), collectedField.SelectionSet)
		if err != nil {
			return err
		}
		rCtx.typeAndFieldCount[nesting]--
	}

	return nil
}
