package main

import (
	"fmt"
	"time"
)

func (s *Server) writable(stream *Stream) func([]byte, bool) (int, error) {
	writable := func(data []byte, lossy bool) (int, error) {
		// Randomly drop packets
		if lossy && s.rand.Intn(100) < s.lossRate {
			s.logger.Info(fmt.Sprintf("Dropped packet from %v", stream.addr))
			time.Sleep(s.timeout * time.Second)
			return 0, nil
		}
		...
		return stream.Write(data)
	}
}
