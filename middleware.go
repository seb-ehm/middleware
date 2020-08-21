//Copyright (c) 2020 Sebastian Ehmann

//Package middleware makes it easy to chain net/http middleware in Go
package middleware

import (
	"net/http"
)

//Middleware is a type alias for the typical signature of a Go net/http middleware
type Middleware func(http.Handler) http.Handler

//New creates a new middleware that does nothing by itself, but can be used as a starting point for
//a chain of middlewares:
//  middleware := middleware.New()
//  middleware.Append(someMiddleware)
func New() Middleware {
	return func(hf http.Handler) http.Handler {
		return hf
	}
}

//End returns a handler that does nothing, which can be used to end a chain of middlewares if no further
//output is wanted:
//  mux := http.NewServeMux()
//  middlewares := middleware.New().Append(someMiddleware)
//  middlewares =  middlewares.Append(someOtherMiddleware)
//  mux.Handle("/endpoint", middlewares(middleware.End()))
func End() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

//Append adds another middleware to the end of the middleware stack.
//Given two middlewares that print greetings to stdout:
//  hello := getGreeting("Hello!")
//  namaste := getGreeting("Namaste!")
//Using them as follows:
//  mux := http.NewServeMux()
//  greetings := middleware.New().Append(hello)
//  greetings = greetings.Append(namaste)
//  mux.Handle("/greetings", greetings(middleware.End())
//Will result in the following output:
//  Hello!
//  Namaste!
func (c Middleware) Append(other Middleware) Middleware {
	oldMiddleware := c
	newMiddleware := func(hf http.Handler) http.Handler {
		return oldMiddleware(other(hf))
	}
	return newMiddleware
}

//Prepend adds another middleware to the beginning of the middleware stack.
//Given two middlewares that print greetings to stdout:
//  hello := getGreeting("Hello!")
//  namaste := getGreeting("Namaste!")
//Using them as follows:
//  mux := http.NewServeMux()
//  greetings := middleware.New().Append(hello)
//  greetings = greetings.Prepend(namaste)
//  mux.Handle("/greetings", greetings(middleware.End())
//Will result in the following output:
//  Namaste!
//  Hello!
func (c Middleware) Prepend(other Middleware) Middleware {
	return other.Append(c)
}

//Assemble can be used to chain an arbitrary number of middlewares
//  middlewares := middleware.Assemble(middlewareA, middlewareB, middlewareC)
func Assemble(middlewares ...Middleware) Middleware {
	assembly := New()
	for _, middleware := range middlewares {
		assembly = assembly.Append(middleware)
	}
	return assembly
}

//ApplyToFunc is a convenience function to apply middleware to a HandlerFunc:
//  mux := http.NewServeMux()
//	handler := func(w http.ResponseWriter, r *http.Request) {
//		fmt.Fprintf(w, "Hi!")
//	}
//  middlewares := := middleware.Assemble(middlewareA, middlewareB, middlewareC)
//  mux.Handle("/endpoint", middlewares.ApplyToFunc(handler))
func (c Middleware) ApplyToFunc(fun http.HandlerFunc) http.Handler {
	return c(fun)
}
