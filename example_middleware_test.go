package middleware_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/seb-ehm/middleware"
)

//MiddlewareA is an example of a direct definition of a middleware.
//It uses a cast to http.HandlerFunc to enable its own ServeHTTP function
func MiddlewareA(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Calling A")
		next.ServeHTTP(w, r)
	})
}

//Greeting is an example of a middleware that uses a struct to keep state
//and an explicit definition of a ServeHTTP function to be useable as http.Handler
type Greeting struct {
	next    http.Handler
	message string
}

//ServeHTTP prints both to the response body and stdout and calls the next middleware afterwards
func (g Greeting) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, g.message)
	g.next.ServeHTTP(w, r)
}

//GetGreeting is a factory function to generate a handler that prints a certain message
func GetGreeting(message string) func(http.Handler) http.Handler {
	fn := func(next http.Handler) http.Handler {
		return Greeting{next, message}
	}
	return fn
}

func GetWebsite(url string) (string, string) {
	resp, err := http.Get(url)
	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(html), resp.Status
}

func GetWebsiteWithHeader(url string, headerName string, headerValue string) (string, string) {
	client := &http.Client{}

	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add(headerName, headerValue)
	resp, err := client.Do(request)
	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(html), resp.Status
}

func ExampleMiddleware() {

	mux := http.NewServeMux()
	//Some handler that prints to the http response
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hi!")
	}
	//Create two middlewares that output different greetings
	hello := GetGreeting("Hello!")
	namaste := GetGreeting("Namaste!")

	//Use middleware.New() as start of a new middleware chain
	greetings := middleware.New().Append(hello)
	greetings = greetings.Append(namaste)

	//Instead of using middleware.New(), any existing handler can be explicitely cast to type Middleware
	//to start a chain
	direct := middleware.Middleware(namaste).Append(hello)

	//Several middlewares can be assembled at once
	assembly := middleware.Assemble(hello, namaste, hello, namaste)

	//A middleware can be prepended to the beginning of the middleware stack
	assembly = assembly.Prepend(hello)

	//Because of the no-op ServeHTTP implementation, any middleware can be used
	//in a mux without an explicit final handler
	mux.Handle("/greetings", greetings)
	mux.Handle("/direct", direct)
	//middleware can also be applied to a final handler
	mux.Handle("/assembly", assembly.ApplyToFunc(handler))

	go func() { log.Fatal(http.ListenAndServe("localhost:9191", mux)) }()

	content, _ := GetWebsite("http://localhost:9191/greetings")
	fmt.Println(content)
	content, _ = GetWebsite("http://localhost:9191/direct")
	fmt.Println(content)
	content, _ = GetWebsite("http://localhost:9191/assembly")
	fmt.Println(content)

	// Output:
	// Hello!
	// Namaste!
	//
	// Namaste!
	// Hello!
	//
	// Hello!
	// Hello!
	// Namaste!
	// Hello!
	// Namaste!
	// Hi!
}
