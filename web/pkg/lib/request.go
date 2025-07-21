package lib

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

const (
	HeaderXForwardedFor = "X-Forwarded-For"
	HederXRealIP        = "X-Real-IP"
)

/*
GetRequestIP returns the IP address of the client that made the request.

It first checks the `X-Forwarded-For` header, then the `X-Real-IP` header, and finally the remote address of the request.
*/
func GetRequestIP(request *http.Request) (string, error) {
	var (
		clientAddr string
		err        error
	)

	if clientAddr = request.Header.Get(HeaderXForwardedFor); clientAddr == "" {
		clientAddr = request.Header.Get(HederXRealIP)
	}

	if clientAddr == "" {
		clientAddr, _, err = net.SplitHostPort(request.RemoteAddr)
		if err != nil {
			log.Error().Err(err).Msg("failed to split remote address")
			return "", errors.New("request does not contain remote address")
		}
	}

	if strings.Index(clientAddr, ",") > 0 {
		splitAddresses := strings.Split(clientAddr, ",")
		clientAddr = strings.TrimSpace(splitAddresses[len(splitAddresses)-1])
	}

	ip := net.ParseIP(clientAddr)

	if ip == nil {
		return "", errors.New("failed to parse IP address")
	}

	if ip.String() == "::1" {
		ip = net.ParseIP("127.0.0.1")
	}

	return ip.String(), nil
}
