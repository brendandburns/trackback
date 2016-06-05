package trackback

import (
	"net"
)

func FindProcesses(pts ProcessToSocket, address net.IP, port uint32) []uint32 {
	result := []uint32{}
	for proc, sockets := range pts {
		for ix := range sockets {
			socket := &sockets[ix]
			if socket.LocalAddress.IP.Equal(address) && socket.LocalAddress.Port == port {
				result = append(result, proc)
				continue
			}
			if socket.RemoteAddress.IP.Equal(address) && socket.RemoteAddress.Port == port {
				result = append(result, proc)
			}
		}
	}
	return result
}
