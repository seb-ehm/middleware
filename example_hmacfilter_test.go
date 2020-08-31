package middleware_test

import (
	"bytes"
	"fmt"
	"github.com/seb-ehm/middleware"
	"log"
	"net/http"
)

func ExampleHmacFilter() {
	mux := http.NewServeMux()
	//Some handler that prints both to stdout and the http response
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hi!")
	}

	githubSignature := middleware.HmacFilter(middleware.HmacParams{Provider: "github", Secret: "ThisIsMySecret"})
	authenticated := middleware.New().Append(githubSignature)

	mux.Handle("/authenticated", authenticated.ApplyToFunc(handler))

	go func() { log.Fatal(http.ListenAndServe("localhost:9194", mux)) }()
	client := &http.Client{}
	request, _ := http.NewRequest("POST", "http://localhost:9194/authenticated", bytes.NewBuffer([]byte("ThisIsARequest")))
	request.Header.Add("X-Hub-Signature", "sha1=8c08e9b7e2bdb4d87982f40d6bf6d36c0d0caab4")
	resp, _ := client.Do(request)

	fmt.Println(resp.Status)
	// Output:
	// 200 OK
}
