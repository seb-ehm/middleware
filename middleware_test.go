package middleware_test

import (
	"net/http"
	"testing"

	"github.com/seb-ehm/middleware"
)

type MockResponseWriter struct {
	output string
}

func (m *MockResponseWriter) Header() http.Header {
	return nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	_ = statusCode
}

func (m *MockResponseWriter) Write(bytes []byte) (int, error) {
	m.output += string(bytes)
	return 0, nil
}

func Test_MiddlewareAssemble(t *testing.T) {
	hello := GetGreeting("Hello!")
	namaste := GetGreeting("Namaste!")

	middlewares := middleware.Assemble(hello, namaste, hello)
	mock := new(MockResponseWriter)
	middlewares.ServeHTTP(mock, nil)
	got := mock.output
	want := "Hello!\nNamaste!\nHello!\n"

	t.Run("Middleware Assemble", func(t *testing.T) {
		if got != want {
			t.Errorf("Append(): got %v, want %v", got, want)
		}
	})

}

func Test_MiddlewarePrepend(t *testing.T) {
	hello := GetGreeting("Hello!")
	namaste := GetGreeting("Namaste!")

	middlewares := middleware.Middleware(hello).Prepend(namaste)
	mock := new(MockResponseWriter)
	middlewares.ServeHTTP(mock, nil)
	got := mock.output
	want := "Namaste!\nHello!\n"

	t.Run("Middleware Append", func(t *testing.T) {
		if got != want {
			t.Errorf("Append(): got %v, want %v", got, want)
		}
	})

}

func Test_MiddlewareAppend(t *testing.T) {
	hello := GetGreeting("Hello!")
	namaste := GetGreeting("Namaste!")

	middlewares := middleware.Middleware(hello).Append(namaste)
	mock := new(MockResponseWriter)
	middlewares.ServeHTTP(mock, nil)
	got := mock.output
	want := "Hello!\nNamaste!\n"

	t.Run("Middleware Append", func(t *testing.T) {
		if got != want {
			t.Errorf("Append(): got %v, want %v", got, want)
		}
	})

}
