This package includes an extension that can be used as `"github.com/99designs/gqlgen/graphql".HandlerExtension`, 
e.g. in `"github.com/99designs/gqlgen/graphql/handler".Server.Use`.

The extension `RecursionLimitByTypeAndField` limits the number of times the same field of a type can be accessed
in a request (query/mutation).

Usage:
```go
gqlServer := handler.New()
gqlServer.Use(RecursionLimitByTypeAndField(1))
```

This allow only one of each "type.field" field access in a query. For following examples,
consider that both root `user` and `User.friends` returns a type `User` (although firends may return a list).

Allows:
```graphql
query {
  user {
    id
    friends {
      id
    }
  }
}
```

Forbids:
```graphql
query {
  user {
    friends {
      friends {
        id
      }
    }
  }
}
```

`User.friends` is accessed twice here. Once in `user.friends`, and second time on `friends.friends`.


The intention of this extension is to replace `extension.FixedComplexityLimit`, as that is very difficult to configure
properly. With `RecursionLimitByTypeAndField`, the client can query the whole graph in one query, but at least
the query does have an upper bound of its size. If needed, both extensions can be used at the same time.