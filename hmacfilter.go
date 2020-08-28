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
	"os"
	"strconv"
	"time"
)

type hmacFilter struct {
	next     http.Handler
	validate func(*http.Request, []byte) (bool, error)
}

type HmacParams struct {
	provider      string
	secretEnv     string
	hmacSource    string
	nonceSource   string
	timeSource    string
	base64Encoded bool
	includeURL    bool
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
	switch params.provider {
	case "github":
		{
			validateFn = func(r *http.Request, message []byte) (bool, error) {
				messageMAC := []byte(r.Header.Get("X-Hub-Signature"))
				if len(messageMAC) == 0 {
					err := fmt.Errorf("missing HMAC header")
					return false, err
				}
				secret, ok := os.LookupEnv(params.secretEnv)
				if !ok {
					err := fmt.Errorf("missing HMAC enironment variable")
					return false, err
				}
				mac := hmac.New(sha1.New, []byte(secret))
				mac.Write(message)
				expected := mac.Sum(nil)
				return hmac.Equal(messageMAC[5:], expected), nil
			}

		}
	default:
		{
			validateFn = func(r *http.Request, message []byte) (bool, error) {
				messageMAC := []byte(r.Header.Get(params.hmacSource))
				secretEnv, ok := os.LookupEnv(params.secretEnv)
				secret := []byte(secretEnv)
				if !ok {
					err := fmt.Errorf("missing HMAC enironment variable")
					return false, err
				}
				var err error
				if params.base64Encoded {
					secret, err = base64.StdEncoding.DecodeString(string(secret))
					if err != nil {
						return false, fmt.Errorf("invalid HMAC in environment variable %s: %w", params.secretEnv, err)
					}
					messageMAC, err = base64.StdEncoding.DecodeString(string(messageMAC))
					if err != nil {
						return false, fmt.Errorf("invalid signature in header %s: %w", params.hmacSource, err)
					}
				}

				mac := hmac.New(sha256.New, secret)
				if params.includeURL {
					mac.Write([]byte(r.URL.String()))
				}
				if params.nonceSource != "" {
					mac.Write([]byte(r.Header.Get(params.nonceSource)))
				}

				if params.timeSource != "" {
					timestamp, err := strconv.ParseInt(r.Header.Get(params.timeSource), 10, 64)
					if err != nil {
						return false, fmt.Errorf("error in timestamp from header %s: %w", params.timeSource, err)
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
	}
	fn := func(next http.Handler) http.Handler {
		return hmacFilter{next, validateFn}
	}
	return fn
}

func ValidMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}
