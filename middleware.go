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
func (m Middleware) Append(other Middleware) Middleware {
	newMiddleware := func(hf http.Handler) http.Handler {
		return m(other(hf))
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
func (m Middleware) Prepend(other Middleware) Middleware {
	return other.Append(m)
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
//  handler := func(w http.ResponseWriter, r *http.Request) {
//  	fmt.Fprintf(w, "Hi!")
//	}
//  middlewares := := middleware.Assemble(middlewareA, middlewareB, middlewareC)
//  mux.Handle("/endpoint", middlewares.ApplyToFunc(handler))
func (m Middleware) ApplyToFunc(fun http.HandlerFunc) http.Handler {
	return m(fun)
}

//Serve can be used to directly use a chain of middlewares without a final handler.
//  mux := http.NewServeMux()
//  middlewares := middleware.New().Append(someMiddleware)
//  middlewares =  middlewares.Append(someOtherMiddleware)
//  mux.Handle("/endpoint", middlewares.Serve())
func (m Middleware) Serve() http.Handler {
	return m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
}
