package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

type ipFilter struct {
	next    http.Handler
	permittedNets []* net.IPNet
	ipHeader  string
}

func (ipf ipFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var ip string
	if ipf.ipHeader != "" {
		ip = r.Header.Get(ipf.ipHeader)
	} else {
		ip = r.RemoteAddr
	}
	isPermittedIP, err := isPermittedIP(ip, ipf.permittedNets)
	if err == nil && isPermittedIP {
		ipf.next.ServeHTTP(w, r)
	} else {
		log.Printf("IP %s is not permitted to access %s \n", ip, r.URL )
	}

}

func isPermittedIP(remoteIP string, permittedNets []*net.IPNet) (bool, error) {
	ip := getIpFromString(remoteIP)
	if ip == nil {
		return false, fmt.Errorf("invalid IP: %s", remoteIP)
	}
	for _, prmn := range permittedNets {
		if prmn.Contains(ip){
			return true, nil
		}
	}
	return false, nil

}

func IPFilter(ipRanges []string, header string) func(http.Handler) http.Handler {
	permittedNets, err := convertToIPNet(ipRanges)
	if err != nil{
		panic(fmt.Sprintf("Failed to convert ip ranges %s", err))
	}
	fn := func(next http.Handler) http.Handler {
		return ipFilter{next, permittedNets, header}
	}
	return fn
}

func convertToIPNet(ipRanges []string) ([]*net.IPNet, error) {
	var ipNets []*net.IPNet
	for _, ipr := range ipRanges{
		// allow special values for localhost
		if ipr == "localhost" || ipr == "loopback"{
			ipr = "::1/128"
		}
		// allow single IP addresses without CIDR suffix
		if strings.Index(ipr, "/") == -1{
			if strings.Index(ipr, ".") !=-1 { //assume IPv4
				ipr += "/32"
			} else if strings.Index(ipr, ":") != -1 { //assume IPv6
				ipr += "/128"
			} else {
				return nil, fmt.Errorf("invalid ip: %s", ipr)
			}
		}
		_, ipNet, err := net.ParseCIDR(ipr)
		if err!=nil{
			return nil, err
		}

		if  ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil { // Add IPv6 loopback if permitted range is IPv4 loopback
				loopbackIPv6 := net.IPNet{IP: net.IPv6loopback, Mask: net.CIDRMask(128, 128)}
				ipNets = append(ipNets, &loopbackIPv6)
			} else { // Add IPv4 loopback if permitted range is IPv6 loopback
				loopbackIPv6 := net.IPNet{IP: net.IPv4(127, 0, 0, 0), Mask: net.CIDRMask(8, 32)}
				ipNets = append(ipNets, &loopbackIPv6)
			}
		}
		ipNets = append(ipNets, ipNet)
	}
	fmt.Println(ipNets)
	return ipNets, nil
}

func getIpFromString(addr string) net.IP {
	for _, c := range "[]\"" {
		addr = strings.ReplaceAll(addr, string(c), "")
	}
	colonIndex := strings.LastIndex(addr, ":")
	if colonIndex == -1 {
		return nil
	}
	addr = addr[:strings.LastIndex(addr, ":")]
	ip := net.ParseIP(addr)
	return ip
}


