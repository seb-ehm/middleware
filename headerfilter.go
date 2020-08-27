package middleware

import (
	"log"
	"net/http"
	"net/textproto"
)

type headerFilter struct {
	next    http.Handler
	headers http.Header
}

func (he headerFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	allHeadersPresent := AllHeadersPresent(he.headers, r.Header)

	if allHeadersPresent {
		he.next.ServeHTTP(w, r)
	} else {
		w.WriteHeader(403)
		log.Printf("IP %s is not permitted to access %s : header verification failed. Supplied headers: %v \n", r.RemoteAddr, r.URL, r.Header)
	}
}

func AllHeadersPresent(requiredHeaders http.Header, requestHeaders http.Header) bool {
	allHeadersPresent := true
	for requiredKey, requiredValues := range requiredHeaders {
		requiredKey = textproto.CanonicalMIMEHeaderKey(requiredKey)
		requestValues := requestHeaders.Values(requiredKey)
		if requestValues == nil {
			allHeadersPresent = false
			break
		}
		if len(requiredValues) > len(requestValues) {
			allHeadersPresent = false
			break
		}

		requestValueMap := make(map[string]bool)
		for _, v := range requestValues {
			requestValueMap[v] = true
		}
		for _, requiredValue := range requiredValues {
			if !requestValueMap[requiredValue] {
				allHeadersPresent = false
				break
			}
		}
	}
	return allHeadersPresent
}

func FilterHeaders(headers http.Header) func(http.Handler) http.Handler {


	fn := func(next http.Handler) http.Handler {
		return headerFilter{next, headers}
	}
	return fn
}

