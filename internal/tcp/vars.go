package tcp

import (
	"context"
	"io"
	"net"
	"syscall"

	"gvisor.dev/gvisor/pkg/tcpip"
)

// expectedCopyErrs are the typed connection-ended errors matched with
// errors.Is.
var expectedCopyErrs = []error{
	io.EOF,
	net.ErrClosed,
	syscall.ECONNRESET,
	syscall.EPIPE,
	context.Canceled,
}

// netstackDisconnectErrs are the canonical messages for connection-ended errors
// from tsnet's userspace netstack (gVisor). gonet stringifies these via
// errors.New(err.String()) at its boundary, discarding the type, so they can't
// be matched with errors.Is. We compare against the strings produced by
// gVisor's own exported error types rather than hardcoding literals, so the set
// tracks the library.
var netstackDisconnectErrs = []string{
	(&tcpip.ErrConnectionReset{}).String(),
	(&tcpip.ErrConnectionAborted{}).String(),
	(&tcpip.ErrClosedForSend{}).String(),
	(&tcpip.ErrClosedForReceive{}).String(),
	(&tcpip.ErrNotConnected{}).String(),
}
