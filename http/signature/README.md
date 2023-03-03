This package is intended to reduce duplication of common steps in http handlers implemented as `http.HandlerFunc`.
Those common steps are parsing an input from request, unmarshaling the result into `http.ResponseWriter`
and handling errors that can occur in any of those steps or inside the handler logic itself.

It does this by allowing handlers to be defined with a new function signature, that can include an input type
as parameter, a response type in return values, and always has `error` as last return value.

The handlers with enhanced signature can be than wrapped using function like `signature.WrapHandler` so it can be
used as a `http.HandlerFunc` type

Example:
```
main() {
	w := signature.DefaultWrapper()

	r := chi.NewRouter()
	r.Get("/endpoint1", signature.WrapHandler(w, handleEndpoint1))
}

func handleEndpoint1(w http.ResponseWriter, r *http.Request, input MyInputStruct) (MyResponseStruct, error){
	// access the input variable of type MyInputStruct, do handler logic, return MyResponseStruct or error
	return theLogic(input)
}
```

Instead of using the repetetive http.HandlerFunc:
```
main() {
	r := chi.NewRouter()
	r.Get("/endpoint1", handleEndpoint1)
}

func handleEndpoint1(w http.ResponseWriter, r *http.Request) {
	if err := parseIntoMyStruct(&MyInputStruct{}); err != nil {
		writeError(w, err)
		return
	}
    
	// in this case, the actual logic is still just one line
	result, err := theLogic(input)
	if err != nil {
		writeError(w, err)
		return
	}
	
	if err := writeResult(result); err != nil {
		writeError(w, err)
		return
	}
}
```
