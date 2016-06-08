package trackback

import (
        "encoding/json"
        "fmt"
        "net"
        "strconv"
)

// Address combines an IP and a port.  (is there no go builtin for this?)
type Address struct {
        IP   net.IP `json:"ip"`
        Port uint32 `json:"port"`
}

func (a *Address) Equals(b *Address) bool {
	return a.IP.Equal(b.IP) && a.Port == b.Port
}

// SocketInfo holds information about a particular open socket.
type SocketInfo struct {
        LocalAddress  Address `json:"localAddress"`
        RemoteAddress Address `json:"remoteAddress"`
        Inode         uint32  `json:"inode"`
        Type          string  `json:"type"`
}

// ProcessToSocket holds a map from PID to sockets open in that process' namespace
type ProcessToSocket map[uint32][]SocketInfo

func (p ProcessToSocket) MarshalJSON() ([]byte, error) {
        info := map[string][]SocketInfo{}
        for proc := range p {
                info[fmt.Sprintf("%d", proc)] = p[proc]
        }
        return json.Marshal(info)
}

func (p ProcessToSocket) UnmarshalJSON(data []byte) error {
        info := map[string][]SocketInfo{}
        if err := json.Unmarshal(data, &info); err != nil {
                return err
        }
        for id := range info {
                pid, err := strconv.Atoi(id)
                if err != nil {
                        return err
                }
                p[uint32(pid)] = info[id]
        }
        return nil
}

// Tracker represents the interface for introspecting information from the system
type Tracker interface {
	// TrackConnections returns info for all connections on the machine, both in the host
	// namespace as well as any other namespaces
	TrackConnections() (ProcessToSocket, error)

	// TrackConnectionsInNamespace returns info for all connections in the network namespace that
	// the specified PID lives in.  Uses /proc/${pid}/ns/net to determine the namespace.
	TrackConnectionsInNamespace(pid uint32) (ProcessToSocket, error)

	// TrackConnectionsInCurrentNamespace returns connection info for the current namespace
	TrackConnectionsInCurrentNamespace() (ProcessToSocket, error)
}

// LookupParameters represents the criteria for what matches for a lookup
type LookupParameters struct {
	// LocalAddress is the local address to match, nil matches all addresses.
	LocalAddress  *Address
	// RemoteAddress is the remote address to match, nil matches all addresses.
	RemoteAddress *Address
}

// Lookup represents the interface for doing lookups
type Lookup interface {
	// FindProcesses gets the processes that holds a particular network connection.  If no matching
	// process is found, nil is returned.
	FindProcesses(params LookupParameters) ([]uint32, error)
}


