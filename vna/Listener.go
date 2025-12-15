package vna

import (
	"log"
	"net"
	"time"
)

func (v *VNA) RunServerListener() {
    v.wg.Add(1)

    go func() {
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
    }()
}

func (v *VNA) processIncomingPacket(pkt []byte, clientAddr *net.UDPAddr) {
    v.RegisterClientPacket(clientAddr, nil)

    key := clientAddr.String()
    v.ClientsMu.RLock()
    sess := v.ClientByAddr[key]
    v.ClientsMu.RUnlock()

    if sess == nil {
        return
    }

    // ===== HANDSHAKE=====
    v.ClientsMu.Lock()
    if !sess.HandShakeDone {
        sess.HandShakeDone = true
        v.ClientsMu.Unlock()

        if len(pkt) != 32 {

            v.ClientsMu.Lock()
            sess.HandShakeDone = false
            v.ClientsMu.Unlock()
            return
        }

        if err := v.Handshake(clientAddr, pkt); err != nil {
            log.Printf("Handshake selhal s %s: %v", clientAddr, err)
            v.ClientsMu.Lock()
            sess.HandShakeDone = false
            v.ClientsMu.Unlock()
        }
        return
    }
    v.ClientsMu.Unlock()


    if sess.Aead == nil {
        return
    }

    plainPkt, err := decryptFrame(sess.Aead, pkt)
    if err != nil {
        log.Printf("Decrypt selhal od %s: %v", clientAddr, err)
        return
    }

    v.RegisterClientPacket(clientAddr, plainPkt)

    if _, err := v.Iface.Write(plainPkt); err != nil {
        log.Println("Chyba zápisu do TUN:", err)
    }
}