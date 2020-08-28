package middleware

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type hmacFilter struct {
	next        http.Handler
	secret      string
	parameterFn func(r *http.Request) (string, string)
	verifyHmac  func([]byte, []byte, []byte, byte[]) bool
}

func (hm hmacFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Could not read request from IP %s to %s", r.RemoteAddr, r.URL)
		w.WriteHeader(403)
		return
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	receivedMAC := r.Header.Get(hm.hmacHeader)
	if ValidMAC(body, []byte(receivedMAC), []byte(hm.secret)) {
		hm.next.ServeHTTP(w, r)
	} else {
		w.WriteHeader(403)
		log.Printf("IP %s is not permitted to access %s : invalid HMAC: \n", r.RemoteAddr, r.URL)
	}
}

func HmacFilter(provider string, secretEnv string) func(http.Handler) http.Handler {
	var parameterFn func(*http.Request, string) (string, string, string)
	var verifyFn func([]byte, []byte, []byte, []byte) bool
	switch provider {
	case "github":
		{
			parameterFn = func(r *http.Request, secretEnv string) (string, string, string) {
				messageMAC := r.Header.Get("X-Hub-Signature")
				secret := os.Getenv(secretEnv)
				messageNonce := ""
				return messageMAC, messageNonce, secret
			}
			verifyFn = func(message, messageMAC, messageNonce, key []byte) bool {
				mac := hmac.New(sha1.New, key)
				mac.Write(message)
				expected := mac.Sum(nil)
				return hmac.Equal(messageMAC[5:], expected)
			}
		}
	default:
		{
			parameterFn = func(r *http.Request, secretEnv string) (string, string) {
				messageMAC := r.Header.Get()
				return "", ""
			}
			verifyFn = func(message, messageMAC, messageNonce, key []byte) bool {
				mac := hmac.New(sha1.New, key)
				mac.Write(message)
				expected := mac.Sum(nil)
				return hmac.Equal(messageMAC[5:], expected)
			}
		}

	}

	return func(next http.Handler) http.Handler {
		return hmacFilter{next, os.Getenv(secretEnv), parameterFn, verifyFn}
	}
}

func ValidMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}
