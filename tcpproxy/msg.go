package tcpproxy

type PortMappingInfo struct {
	LocalPort  int    `json:"localPort"`
	RemoteAddr string `json:"remoteAddr"`
}
