package vna

import (
	"time"
)

func (v *VNA) runServerListener() {
    defer v.wg.Done()

    buf := make([]byte, 65535)

    for {
        if v.CtxStopped() {
            return
        }
        _ = v.Conn.SetReadDeadline(time.Now().Add(1 * time.Second))
        n, clientAddr, err := v.Conn.ReadFromUDP(buf)
        
        if err != nil {    
            continue
        }

        pkt := make([]byte, n)
        copy(pkt, buf[:n])
        v.routeIncomingPacket(pkt,clientAddr)
    
    }
}

