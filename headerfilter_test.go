package middleware

import (
	"net/http"
	"testing"
)

func TestAllHeadersPresent(t *testing.T) {
	type args struct {
		requiredHeaders http.Header
		requestHeaders  http.Header
	}

	singleHeader := map[string][]string{"Secretkey": {"secretvalue"}}
	wrongValue := map[string][]string{"Secretkey": {"wrongvalue"}}
	wrongHeader := map[string][]string{"Wrongkey": {"wrongvalue"}}
	multipleHeaders:= map[string][]string{"Secretkey": {"secretvalue"}, "Otherkey": {"othervalue"}}
	multipleValues := map[string][]string{"Secretkey": {"secretvalue", "secretvalue2"}}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{ "Single Header Single Value present", args{singleHeader,singleHeader}, true },
		{ "Single Header Single Value missing", args{singleHeader,nil}, false },
		{ "Single Header wrong value", args{singleHeader,wrongValue}, false },
		{ "Single Header wrong header", args{singleHeader,wrongHeader}, false },
		{ "Multiple Headers present", args{multipleHeaders,multipleHeaders}, true },
		{ "Multiple Headers, one missing", args{multipleHeaders,singleHeader}, false },
		{ "Multiple Values present", args{multipleValues,multipleValues}, true },
		{ "Multiple Values, one missing", args{multipleValues,singleHeader}, false },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AllHeadersPresent(tt.args.requiredHeaders, tt.args.requestHeaders); got != tt.want {
				t.Errorf("AllHeadersPresent() = %v, want %v", got, tt.want)
			}
		})
	}
}
