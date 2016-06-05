package trackback

import (
        "encoding/json"
        "fmt"
        "net"
	"os"
        "os/exec"
        "regexp"

        "github.com/golang/glog"
        psnet "github.com/shirou/gopsutil/net"
)

var (
        whitespaceRE = regexp.MustCompile("\\s+")
)

// Interfaces for test injection
type connInterface func(which string) ([]psnet.ConnectionStat, error)

type pidsInterface func() ([]int32, error)

type execInterface interface {
	Exec(cmd string, args ...string) ([]byte, error)
}

type osExec struct {}

func (osExec) Exec(cmd string, args ...string) ([]byte, error) {
	ex := exec.Command(cmd, args...)
	return ex.Output()
}

type trackbackImpl struct {
	cmd string
	exec execInterface
	conn connInterface
	pids pidsInterface
}

func NewTrackback(cmd string) (Interface, error) {
	return &trackbackImpl{
		cmd: cmd,
		exec: osExec{},
		conn: psnet.Connections,
		pids: psnet.Pids,
	}, nil
}

func (t *trackbackImpl) TrackConnectionsInNamespace(pid uint32) (ProcessToSocket, error) {
	data, err := t.exec.Exec("nsenter", fmt.Sprintf("--net=/proc/%d/ns/net", pid), t.cmd)
        if err != nil {
                return nil, err
        }
        pts := ProcessToSocket{}
        if err := json.Unmarshal(data, &pts); err != nil {
                return nil, err
        }
        return pts, nil
}

func (t *trackbackImpl) TrackConnectionsInCurrentNamespace() (ProcessToSocket, error) {
        conns, err := t.conn("all")
        if err != nil {
                return nil, err
        }
        result := ProcessToSocket{}
        for ix := range conns {
                conn := &conns[ix]
                if conn.Laddr.Port == 0 && conn.Raddr.Port == 0 {
                        glog.V(2).Infof("Skipping %v because ports are zero", conn)
                        continue
                }
                localIP := net.ParseIP(conn.Laddr.IP)
                if localIP == nil {
                        glog.Warningf("Skipping local: %s", conn.Laddr.IP)
                        continue
                }
                remoteIP := net.ParseIP(conn.Raddr.IP)
                if remoteIP == nil {
                        glog.Warningf("Skipping remote: %s", conn.Raddr.IP)
                        continue
                }
                result[uint32(conn.Pid)] = append(result[uint32(conn.Pid)], SocketInfo{
                        RemoteAddress: Address{
                                IP:   remoteIP,
                                Port: conn.Raddr.Port,
                        },
                        LocalAddress: Address{
                                IP:   localIP,
                                Port: conn.Laddr.Port,
                        },
                })
        }
        return result, nil
}

func (t *trackbackImpl) TrackConnections() (ProcessToSocket, error) {
	return trackAllConnections(t, t.pids)
}

func trackAllConnections(t Interface, pidsFn pidsInterface) (ProcessToSocket, error) {
	visitedNamespaces := map[string]bool{}
	nsName, err := os.Readlink("/proc/self/ns/net")
	if err != nil {
		return nil, err
	}
	pts, err := t.TrackConnectionsInCurrentNamespace()
	if err != nil {
		return nil, err
	}
	visitedNamespaces[nsName] = true
	pids, err := pidsFn()
	if err != nil {
		return nil, err
	}
	for _, pid := range pids {
		nsName, err := os.Readlink(fmt.Sprintf("/proc/%d/ns/net", pid))
		if err != nil {
			return nil, err
		}
		if !visitedNamespaces[nsName] {
			pts2, err := t.TrackConnectionsInNamespace(uint32(pid))
			if err != nil {
				return nil, err
			}
			for key := range pts2 {
				pts[key] = pts2[key]
			}
			visitedNamespaces[nsName] = true
		}
	}
	return pts, nil
}

