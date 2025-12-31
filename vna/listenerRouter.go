package vna

import (
	"VPNServer/crypted"
	"log"
	"net"
	"time"
)

func (v *VNA) routeIncomingPacket(pkt []byte, addr *net.UDPAddr) {

     if len(pkt) < 1 {
        return
    }

    ////===Get type and payload fron Incoming packet===
    ////===pkt type is based on protocol - Handshake,Data,Ip request===
    pktType := PacketType(pkt[0])
    payload := pkt[1:]

    
    switch pktType {
    
	case PacketIPRequest:
		v.processIPRequest(addr)

    case PacketHandshake:
        v.processHandshake(addr, payload)

    case PacketData:
        v.processDataPacket(addr, payload)
	
    default:
        log.Printf("Unknown packet type %d from %s", pktType, addr)
    }
}

func (v *VNA) processIPRequest(addr *net.UDPAddr) {

    ip, err := v.IPPool.reserveIP(10 * time.Second)
    if err != nil {
        log.Printf("IP allocation failed for %s: %v", addr, err)
        return
    }

    mask := net.IPv4(255, 255, 255, 255) 

    log.Printf("Reserved IP %s for %s", ip, addr)

    v.sendIPResponse(addr, ip, mask)
}

func (v *VNA) processHandshake(addr *net.UDPAddr, pkt []byte) {

    v.ClientsMu.Lock()
    
    ///=====Remove Client Session if already exist=======
    if old, ok := v.ClientByAddr[addr.String()]; ok {
        v.removeClientLocked(addr.String(), old)
    }

    ///=====Create new Client Session=======
    sess := v.createClientSession(addr)
	
    v.ClientsMu.Unlock()

    ///======Make handshake with client=======
    if err := v.Handshake(addr, pkt); err != nil {
        log.Printf("Handshake failed %s: %v", addr, err)
        return
    }
    sess.LastSeen = time.Now()

    log.Printf("Handshake OK %s", addr)
}


func (v *VNA) processDataPacket(addr *net.UDPAddr, pkt []byte) {

    ///===Get Client session based on clients ip if not exist return===
    sess := v.getClient(addr)    
	if sess == nil || sess.Aead == nil {
        return
    }

    ///===Decrypt Payload===
    plainPkt, err := crypted.DecryptFrame(sess.Aead, pkt)
    
	if err != nil {
        log.Printf("Decrypt failed from %s: %v", addr, err)
        return
    }

    ///===UpdateClientState===
    //check if clients VPN ip changed
    //if does change it
    //also change OS route to new client VPN IP
    sess.LastSeen = time.Now()
	v.updateClientState(sess, plainPkt)


    ///===Write data in to VNA===
    if _, err := v.Iface.Write(plainPkt); err != nil {
        log.Println("TUN write error:", err)
    }

}


