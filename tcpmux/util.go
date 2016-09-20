package tcpmux

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"
)

func ReadHeader(c io.Reader) (string, error) {
	lenBuf := make([]byte, 1)
	_, err := io.ReadFull(c, lenBuf)
	if err != nil {
		return "", err
	}

	len := byte(lenBuf[0])
	if len <= 0 {
		return "", errors.New("invalid code header")
	}

	buf := make([]byte, len)
	_, err = io.ReadFull(c, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func WriteHeader(code string, c io.Writer) error {
	len := len(code)
	if len > 127 {
		return errors.New("code too long")
	}

	buf := make([]byte, 0)
	buf = append(buf, byte(len))
	buf = append(buf, []byte(code)...)

	_, err := c.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

func Pipeline(r, w *net.TCPConn, wg *sync.WaitGroup) error {
	defer wg.Done()
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		// log.Printf("read: %d - %v\n", n, err)
		if n > 0 {
			n2, err := w.Write(buf[:n])
			// log.Printf("write: %d - %v\n", n2, err)
			if n2 != n {
				log.Printf("failed to pipeline data: %v", err)
				return err
			}
		}

		if err != nil {
			w.CloseWrite()
			return err
		}
	}
}
