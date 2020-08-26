package middleware

import (
	"net"
	"testing"
)

func Test_getIpFromString(t *testing.T) {

	tests := []struct {
		name string
		addr string
		want net.IP
	}{
		{"localhost IPv4", "127.0.0.1:1234", net.ParseIP("127.0.0.1")},
		{"IPv4 in quotes", "\"127.0.0.1:1234\"", net.ParseIP("127.0.0.1")},
		{"localhost IPv6", "[::1]:57048", net.ParseIP("::1")},
		{"IPv6 in quotes", "\"[::1]\":57048", net.ParseIP("::1")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getIpFromString(tt.addr); got.String() != tt.want.String() {
				t.Errorf("getIpFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isPermittedIP(t *testing.T) {
	type args struct {
		remoteIP string
		permittedNets []*net.IPNet
	}
	net1, _ := convertToIPNet([]string{"127.0.0.1/32"})
	net2, _ := convertToIPNet([]string{"::1/128"})
	net3, _ := convertToIPNet([]string{"127.0.0.1/32"})
	net4, _ := convertToIPNet([]string{"::1/128"})
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"Localhost IPv4 ", args{"127.0.0.1:1234", net1}, true, false},
		{"Localhost IPv6 ", args{"[::1]:57048", net2}, true, false},
		{"Loopback Mix IPv6 IPv4", args{"[::1]:1234", net3 }, true, false},
		{"Loopback Mix IPv4 IPv6", args{"127.0.0.1:12345", net4}, true, false},
		{"Loopback Mix Permitted 1", args{"127.1.0.1:12345", net4}, true, false},
		{"Loopback Mix Permitted 2", args{"127.255.0.1:12345", net4}, true, false},
		{"Loopback Mix IP not permitted", args{"128.0.0.1:12345", net4}, false, false},
		{"Loopback Mix IP not permitted", args{"126.0.0.1:12345", net4}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isPermittedIP(tt.args.remoteIP, tt.args.permittedNets)
			if (err != nil) != tt.wantErr {
				t.Errorf("isPermittedIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isPermittedIP() got = %v, want %v", got, tt.want)
			}
		})
	}
}

