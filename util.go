package trackback

func FindProcesses(pts ProcessToSocket, local *Address, remote *Address) []uint32 {
	result := []uint32{}
	for proc, sockets := range pts {
		for ix := range sockets {
			socket := &sockets[ix]
			if local != nil && !socket.LocalAddress.Equals(local) {
				continue
			}
			if remote == nil || socket.RemoteAddress.Equals(remote) {
				result = append(result, proc)
			}
		}
	}
	return result
}
