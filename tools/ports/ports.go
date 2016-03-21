package ports

import (
	"net"
	"strconv"
	"strings"
)

func GetFreePort() (int, error) {
	p, err := GetFreePorts(1)
	if err != nil {
		return 0, err
	}
	return p[0], nil
}

func GetFreePorts(howMany int) ([]int, error) {
	ports := make([]int, howMany)
	for i := 0; i < howMany; i++ {
		ln, err := net.Listen("tcp", ":0")
		defer ln.Close()
		portStr := strings.Trim(ln.Addr().String(), "[]:")
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}
		ports[i] = p
	}
	return ports, nil
}
