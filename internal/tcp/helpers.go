package tcp

import (
	"errors"
	"net"
	"strings"
)

func isExpectedCopyError(err error) bool {
	if err == nil {
		return true
	}

	for _, target := range expectedCopyErrs {
		if errors.Is(err, target) {
			return true
		}
	}

	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}

	msg := err.Error()
	for _, s := range netstackDisconnectErrs {
		if strings.Contains(msg, s) {
			return true
		}
	}

	return false
}
