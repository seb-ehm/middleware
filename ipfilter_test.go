package middleware

import (
	"net"
	"reflect"
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
		{"Invalid IP", "aabbccd", nil},
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
		remoteIP      string
		permittedNets []*net.IPNet
	}
	singleIPNoSuffix, _ := convertToIPNet([]string{"192.168.1.1"})
	localhostIPv4, _ := convertToIPNet([]string{"127.0.0.1/32"})
	localhostIPv6, _ := convertToIPNet([]string{"::1/128"})
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"Invalid IP", args{"aabbcc", localhostIPv4}, false, true},
		{"Localhost IPv4 ", args{"127.0.0.1:1234", localhostIPv4}, true, false},
		{"Localhost IPv6 ", args{"[::1]:57048", localhostIPv6}, true, false},
		{"Single IPv4 No Suffix ", args{"192.168.1.1:1234", singleIPNoSuffix}, true, false},
		{"Loopback Mix IPv6 IPv4", args{"[::1]:1234", localhostIPv4}, true, false},
		{"Loopback Mix IPv4 IPv6", args{"127.0.0.1:12345", localhostIPv6}, true, false},
		{"Loopback Mix Permitted 1", args{"127.1.0.1:12345", localhostIPv6}, true, false},
		{"Loopback Mix Permitted 2", args{"127.255.0.1:12345", localhostIPv6}, true, false},
		{"Loopback Mix IP not permitted", args{"128.0.0.1:12345", localhostIPv6}, false, false},
		{"Loopback Mix IP not permitted", args{"126.0.0.1:12345", localhostIPv6}, false, false},
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

func Test_convertToIPNet(t *testing.T) {
	type args struct {
		ipRanges []string
	}
	localhostIPv6 := net.IPNet{IP: net.IPv6loopback, Mask: net.CIDRMask(128, 128)}
	localhostIPv4 := net.IPNet{IP: net.IPv4(127, 0, 0, 0), Mask: net.CIDRMask(8, 32)}

	tests := []struct {
		name    string
		args    args
		want    []*net.IPNet
		wantErr bool
	}{
		{"invalid characters", args{[]string{"abcdefg", "hijklmnop"}}, nil, true},
		{"invalid IPNet", args{[]string{"999.999.999"}}, nil, true},
		{"ipv6 localhost no /", args{[]string{"::1"}}, []*net.IPNet{&localhostIPv4, &localhostIPv6}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToIPNet(tt.args.ipRanges)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToIPNet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToIPNet() got = %v, want %v", got, tt.want)
			}
		})
	}
}
