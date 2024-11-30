
package node

import (
	"net"
	"strings"
)

// GetIPWithoutDots returns the IP address of the computer in dotted notation without dots
func GetIPWithoutDots() (string) {
	// Get the list of network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Get the addresses for this interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			// Check if it's an IP network
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			// Check if it's an IPv4 address
			ip := ipNet.IP.To4()
			if ip == nil {
				continue
			}

			// Convert IP to dotted notation
			ipString := ip.String()

			// Remove dots from the IP string
			ipWithoutDots := strings.ReplaceAll(ipString, ".", "")

			return ipWithoutDots
		}
	}
	return ""
}