package tcp

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// DialFunc dials an outbound connection. Callers may pass net.Dial or
// tsnet.Server.Dial depending on whether the target is reachable via the
// host network or the Tailscale network respectively.
type DialFunc func(ctx context.Context, network, address string) (net.Conn, error)

func Forward(sourceConn net.Conn, targetAddr string, targetPort int) error {
	defer sourceConn.Close()

	targetConn, err := net.Dial("tcp", net.JoinHostPort(targetAddr, fmt.Sprintf("%d", targetPort)))
	if err != nil {
		return fmt.Errorf("failed to dial target: %w", err)
	}

	defer targetConn.Close()

	if tcpConn, ok := targetConn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	once := sync.Once{}

	closeBoth := func() {
		once.Do(func() {
			sourceConn.Close()
			targetConn.Close()
		})
	}

	var (
		wg      sync.WaitGroup
		errOnce sync.Once
		copyErr error
	)

	wg.Add(2)

	pipe := func(dst, src net.Conn) {
		defer wg.Done()
		defer closeBoth()

		if _, err := io.Copy(dst, src); err != nil && !isExpectedCopyError(err) {
			errOnce.Do(func() { copyErr = err })
		}
	}

	go pipe(targetConn, sourceConn)
	go pipe(sourceConn, targetConn)

	wg.Wait()

	if copyErr != nil {
		return fmt.Errorf("failed to copy data: %w", copyErr)
	}

	return nil
}

// ForwardEgress forwards an inbound sourceConn to a target address using the
// provided dial function. Use ts.Dial as the dial function to reach targets
// over the Tailscale network (e.g. MagicDNS hostnames).
func ForwardEgress(sourceConn net.Conn, dial DialFunc, targetAddr string, targetPort int) error {
	defer sourceConn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	targetConn, err := dial(ctx, "tcp", net.JoinHostPort(targetAddr, fmt.Sprintf("%d", targetPort)))
	if err != nil {
		return fmt.Errorf("failed to dial egress target: %w", err)
	}

	defer targetConn.Close()

	if tcpConn, ok := targetConn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	once := sync.Once{}

	closeBoth := func() {
		once.Do(func() {
			sourceConn.Close()
			targetConn.Close()
		})
	}

	var (
		wg      sync.WaitGroup
		errOnce sync.Once
		copyErr error
	)

	wg.Add(2)

	pipe := func(dst, src net.Conn) {
		defer wg.Done()
		defer closeBoth()

		if _, err := io.Copy(dst, src); err != nil && !isExpectedCopyError(err) {
			errOnce.Do(func() { copyErr = err })
		}
	}

	go pipe(targetConn, sourceConn)
	go pipe(sourceConn, targetConn)

	wg.Wait()

	if copyErr != nil {
		return fmt.Errorf("failed to copy egress data: %w", copyErr)
	}

	return nil
}
