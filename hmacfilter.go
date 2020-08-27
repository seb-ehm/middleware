package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"io/ioutil"
	"log"
	"net/http"
)

type hmacFilter struct{
	next http.Handler
	secret string
	hmacHeader string
}

func (hm hmacFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err!=nil{
		log.Printf("Could not read request from IP %s to %s", r.RemoteAddr, r.URL)
		w.WriteHeader(403)
		return
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	receivedMAC := r.Header.Get(hm.hmacHeader)
	if ValidMAC(body, []byte(receivedMAC), []byte(hm.secret)){
		hm.next.ServeHTTP(w,r)
	} else {
		w.WriteHeader(403)
		log.Printf("IP %s is not permitted to access %s : invalid HMAC: \n", r.RemoteAddr, r.URL)
	}
}

func HmacFilter(secret string, hmacHeader string) func(http.Handler) http.Handler {
	_ = secret
	_ = hmacHeader
	return func(next http.Handler) http.Handler{
		return hmacFilter{next, secret, hmacHeader}
	}
}

func ValidMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}