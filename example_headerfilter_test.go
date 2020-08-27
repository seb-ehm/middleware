package middleware_test

import (
	"fmt"
	"github.com/seb-ehm/middleware"
	"log"
	"net/http"
)

func ExampleHeaderFilter() {
	mux := http.NewServeMux()
	//Some handler that prints both to stdout and the http response
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hi!")
	}

	requireHeader := middleware.FilterHeaders(map[string][]string{"mysecretkey": {"mysecretvalue"}})

	authenticated := middleware.New().Append(requireHeader)

	mux.Handle("/authenticated", authenticated.ApplyToFunc(handler))

	go func() { log.Fatal(http.ListenAndServe("localhost:9193", mux)) }()

	_, status := GetWebsiteWithHeader("http://localhost:9193/authenticated", "mysecretkey", "wrongvalue")
	content, _ := GetWebsiteWithHeader("http://localhost:9193/authenticated", "mysecretkey", "mysecretvalue")

	fmt.Println(status)
	fmt.Println(content)
	// Output:
	// 403 Forbidden
	// Hi!
}
