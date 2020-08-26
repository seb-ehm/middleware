package middleware_test

import (
	"fmt"
	"github.com/seb-ehm/middleware"
	"log"
	"net/http"
)

func ExampleIPFilter() {
	mux := http.NewServeMux()
	//Some handler that prints both to stdout and the http response
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hi!")
	}

	allowSomeIP:= middleware.IPFilter([]string{"123.240.189.1/32"}, "")
	noLocalhost := middleware.New().Append(allowSomeIP)

	allowLocalhost:= middleware.IPFilter([]string{"localhost"}, "")
	localhost := middleware.New().Append(allowLocalhost)
	mux.Handle("/nolocalhost", noLocalhost.ApplyToFunc(handler))
	mux.Handle("/localhost", localhost.ApplyToFunc(handler))

	go func() { log.Fatal(http.ListenAndServe("localhost:9192", mux)) }()

	fmt.Println(GetWebsite("http://localhost:9192/nolocalhost"))
	fmt.Println(GetWebsite("http://localhost:9192/localhost"))

	// Output:
	// Hi!
}

