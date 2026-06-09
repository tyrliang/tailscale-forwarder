package main

import (
	"errors"
	"fmt"
	"io"
	"main/internal/util"
	"net"
	"sync"
	"time"
)

// copyBufSize is the per-direction buffer used when relaying data. Buffers are
// pooled to avoid per-connection allocation under concurrent load.
const copyBufSize = 64 * 1024

var bufPool = sync.Pool{
	New: func() any {
		b := make([]byte, copyBufSize)
		return &b
	},
}

func fwdTCP(sourceConn net.Conn, targetAddr string, targetPort int) error {
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

	// When either direction finishes — a graceful FIN, an RST, or any copy
	// error — close both connections so the other goroutine's blocked Read
	// returns immediately. This guarantees that when a client disconnects the
	// matching upstream connection is torn down too, rather than leaking until
	// the process restarts.
	var once sync.Once
	closeBoth := func() {
		once.Do(func() {
			sourceConn.Close()
			targetConn.Close()
		})
	}

	var (
		wg   sync.WaitGroup
		errs [2]error
	)

	wg.Add(2)

	pipe := func(i int, dst, src net.Conn) {
		defer wg.Done()
		defer closeBoth()

		buf := bufPool.Get().(*[]byte)
		defer bufPool.Put(buf)

		if _, err := io.CopyBuffer(dst, src, *buf); err != nil && !util.IsExpectedCopyError(err) {
			errs[i] = err
		}
	}

	go pipe(0, targetConn, sourceConn)
	go pipe(1, sourceConn, targetConn)

	wg.Wait()

	if err := errors.Join(errs[0], errs[1]); err != nil {
		return fmt.Errorf("connection error: %w", err)
	}

	return nil
}
