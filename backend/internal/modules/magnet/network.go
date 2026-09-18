package magnet

import (
	"context"
	"errors"
	"net"
	"net/url"
	"time"

	"github.com/anacrolix/torrent/iplist"
)

func publicIP(ip net.IP) bool {
	if ip == nil || !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	if v4 := ip.To4(); v4 != nil {
		if v4[0] == 0 || v4[0] >= 224 || v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127 {
			return false
		}
	}
	return true
}

type publicPeersOnly struct{}

func (publicPeersOnly) NumRanges() int { return 1 }
func (publicPeersOnly) Lookup(ip net.IP) (iplist.Range, bool) {
	return iplist.Range{First: ip, Last: ip, Description: "non-public address"}, !publicIP(ip)
}
func lookupPublic(ctx context.Context, host string) ([]net.IP, error) {
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, 0, len(addresses))
	for _, address := range addresses {
		if !publicIP(address.IP) {
			return nil, errors.New("tracker resolves to a non-public address")
		}
		ips = append(ips, address.IP)
	}
	if len(ips) == 0 {
		return nil, errors.New("tracker has no public addresses")
	}
	return ips, nil
}
func publicTrackerIPs(u *url.URL) ([]net.IP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return lookupPublic(ctx, u.Hostname())
}
func publicDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := lookupPublic(ctx, host)
	if err != nil {
		return nil, err
	}
	var dialer net.Dialer
	dialer.Timeout = 5 * time.Second
	for _, ip := range ips {
		conn, e := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if e == nil {
			return conn, nil
		}
		err = e
	}
	return nil, err
}
