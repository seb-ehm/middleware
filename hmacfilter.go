package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type hmacFilter struct {
	next     http.Handler
	validate func(*http.Request, []byte) (bool, error)
}

type HmacParams struct {
	Provider      string
	Secret        string
	HmacSource    string
	NonceSource   string
	TimeSource    string
	Base64Encoded bool
	IncludeURL    bool
}

func (hm hmacFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Could not read request from IP %s to %s", r.RemoteAddr, r.URL)
		w.WriteHeader(403)
		return
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	valid, err := hm.validate(r, body)
	if err != nil {
		log.Printf("error validating HMAC for request from IP %s to %s: %v", r.RemoteAddr, r.URL, err)
	}
	if valid {
		hm.next.ServeHTTP(w, r)
	} else {
		w.WriteHeader(403)
		log.Printf("IP %s is not permitted to access %s : invalid HMAC: \n", r.RemoteAddr, r.URL)
	}
}

func HmacFilter(params HmacParams) func(http.Handler) http.Handler {
	var validateFn func(*http.Request, []byte) (bool, error)
	switch params.Provider {
	case "github":
		{
			validateFn = GithubValidation(params)

		}
	default:
		{
			validateFn = DefaultValidation(params)
		}
	}
	fn := func(next http.Handler) http.Handler {
		return hmacFilter{next, validateFn}
	}
	return fn
}

func DefaultValidation(params HmacParams) func(r *http.Request, message []byte) (bool, error) {
	return func(r *http.Request, message []byte) (bool, error) {
		messageMAC := []byte(r.Header.Get(params.HmacSource))
		if len(params.Secret) == 0 {
			err := fmt.Errorf("empty HMAC secret")
			return false, err
		}
		secret := []byte(params.Secret)

		var err error
		if params.Base64Encoded {
			secret, err = base64.StdEncoding.DecodeString(string(secret))
			if err != nil {
				return false, fmt.Errorf("invalid secret: %w", err)
			}
			messageMAC, err = base64.StdEncoding.DecodeString(string(messageMAC))
			if err != nil {
				return false, fmt.Errorf("invalid signature in header %s: %w", params.HmacSource, err)
			}
		}

		mac := hmac.New(sha256.New, secret)
		if params.IncludeURL {
			mac.Write([]byte(r.URL.String()))
		}
		if params.NonceSource != "" {
			mac.Write([]byte(r.Header.Get(params.NonceSource)))
		}

		if params.TimeSource != "" {
			timestamp, err := strconv.ParseInt(r.Header.Get(params.TimeSource), 10, 64)
			if err != nil {
				return false, fmt.Errorf("error in timestamp from header %s: %w", params.TimeSource, err)
			}
			sendTime := time.Unix(timestamp, 0)
			receiveTime := time.Now()
			if receiveTime.Sub(sendTime) > time.Second*2 {
				return false, nil
			}
		}

		expected := mac.Sum(nil)
		return hmac.Equal(messageMAC, expected), nil

	}
}

func GithubValidation(params HmacParams) func(r *http.Request, message []byte) (bool, error) {
	return func(r *http.Request, message []byte) (bool, error) {
		messageMAC := []byte(r.Header.Get("X-Hub-Signature"))
		if len(messageMAC) == 0 {
			err := fmt.Errorf("missing HMAC header")
			return false, err
		}

		if len(params.Secret) == 0 {
			err := fmt.Errorf("empty HMAC secret")
			return false, err
		}
		mac := hmac.New(sha1.New, []byte(params.Secret))
		mac.Write(message)
		expected := mac.Sum(nil)
		test := messageMAC[5:]
		_ = test
		test2,_ :=  base64.StdEncoding.DecodeString(string(test))
		_ = test2
		return hmac.Equal(messageMAC[5:], expected), nil
	}
}

func ValidMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}
