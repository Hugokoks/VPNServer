package vna

import (
	"VPNServer/crypted"
	"log"
	"net"
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
        v.processIncomingPacket(pkt,clientAddr)
    
    }
}

func (v *VNA) processIncomingPacket(pkt []byte, addr *net.UDPAddr) {

    if isHandshakePacket(pkt) {
        v.processHandshake(addr, pkt)
        return
    }

    v.processDataPacket(addr, pkt)
}



func (v *VNA) processHandshake(addr *net.UDPAddr, pkt []byte) {

    v.ClientsMu.Lock()

    if old, ok := v.ClientByAddr[addr.String()]; ok {
        v.removeClientLocked(addr.String(), old)
    }

    sess := v.createClientSession(addr)
    
	v.ClientsMu.Unlock()

    if err := v.Handshake(addr, pkt); err != nil {
        log.Printf("Handshake failed %s: %v", addr, err)
        return
    }

    sess.HandshakeDone = true
    sess.LastSeen = time.Now()

    log.Printf("Handshake OK %s", addr)
}


func (v *VNA) processDataPacket(addr *net.UDPAddr, pkt []byte) {

    sess := v.getClient(addr)
    
	if sess == nil || sess.Aead == nil {
        return
    }

    plainPkt, err := crypted.DecryptFrame(sess.Aead, pkt)
    
	if err != nil {
        log.Printf("Decrypt failed from %s: %v", addr, err)
        return
    }

    sess.LastSeen = time.Now()
	v.updateClientState(sess, plainPkt)


    if _, err := v.Iface.Write(plainPkt); err != nil {
        log.Println("TUN write error:", err)
    }


}