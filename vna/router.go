// router.go
package vna

import (
	"fmt"
	"net"
)

///sending to tunel
func (v *VNA) routeToTun(pkt []byte) {
    
	if len(pkt) < 20 || (pkt[0]>>4) != 4 {
        fmt.Println("drop non-IPv4 packet")
        return
    }

	////get destination IP from pkt
    dstIP := net.IP(pkt[16:20]).String()
	
	///broadcast
    if dstIP == "255.255.255.255" {
        v.forwardBroadcast(pkt)
        return
    }
	///unicast
    v.forwardUnicast(dstIP, pkt)
}

// forwardBroadcast
func (v *VNA) forwardBroadcast(pkt []byte) {
    
	v.ClientsMu.RLock()
    defer v.ClientsMu.RUnlock()

    fmt.Println("Broadcast → forwarding to all clients")

    for _, sess := range v.ClientByVPN {
        if sess.Addr == nil {
            continue
        }
        if err := sendEncryptedTo(sess.Aead, v.Conn, sess.Addr, pkt); err != nil {
            fmt.Println("broadcast send error to", sess.VPNIP, ":", err)
        }
    }
}


// forwardUnicast
func (v *VNA) forwardUnicast(dstIP string, pkt []byte) {
    v.ClientsMu.RLock()
    sess, ok := v.ClientByVPN[dstIP]
    
	if !ok || sess.Addr == nil {
		v.ClientsMu.RUnlock()
        fmt.Println("Unicast drop: no client for", dstIP)
        return
    }
	
	addrCopy := *sess.Addr
	v.ClientsMu.RUnlock()

    if err := sendEncryptedTo(sess.Aead, v.Conn, &addrCopy, pkt); err != nil {
        fmt.Println("Unicast send error:", err)
        return
    }

    fmt.Println("Sent to", dstIP, "→", sess.Addr)
}