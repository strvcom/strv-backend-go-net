Package for parsing path and query parameters from http request into struct, similar to parsing body as json to struct.

```
type MyInputStruct struct {
	UserID   int   `param:"path=id"`
	SomeFlag *bool `param:"query=flag"`
}
```

Then a request like `http://somewhere.com/users/9?flag=true` can be parsed as follows.
In this example, using chi to access path parameters that has a `{id}` wildcard in configured chi router

```
	parsedInput := MyInputStruct{}
	param.DefaultParser().WithPathParamFunc(chi.URLParam).Parse(request, &parsedInput)
```
