package vna

import (
	"context"
	"crypto/cipher"
	"log"
	"net"
	"os/exec"
	"time"
)
type ClientSession struct{
	Addr 	*net.UDPAddr
	LastSeen time.Time
	VPNIP    string   

    Aead cipher.AEAD
    HandshakeDone bool
    HandshakeAt   time.Time

}

func (v *VNA) createClientSession(addr *net.UDPAddr) *ClientSession {
    
    sess := &ClientSession{
        Addr:        addr,
        HandshakeAt: time.Now(),
        LastSeen:    time.Now(),
    }

    v.ClientByAddr[addr.String()] = sess
    return sess
}

func (v *VNA) getClient(addr *net.UDPAddr) *ClientSession {
    v.ClientsMu.RLock()
    defer v.ClientsMu.RUnlock()
    return v.ClientByAddr[addr.String()]
}

func (v *VNA) updateClientState(sess *ClientSession, plainPkt []byte) {
	///===Get Client VPN IP from packet===
    srcIP, ok := extractSrcIPv4(plainPkt)
	
    if !ok {
		return
	}

    ///===packet src IP is same as clients VPN IP===
	if sess.VPNIP == srcIP {
		return
	}

    ///===save VPN old IP=== 
	oldIP := sess.VPNIP
    ///===Replace VPN old IP with new one===
	sess.VPNIP = srcIP
    ///===Create new record in ClientByVPN: make(map[string]*ClientSession) under key - VPN IP===
    v.ClientByVPN[srcIP] = sess

	v.syncClientRoute(oldIP, srcIP)
}

func extractSrcIPv4(pkt []byte) (string, bool) {
	
    if len(pkt) < 20 {
		return "", false
	}
	
    if (pkt[0]>>4) != 4 {
		return "", false
	}
	
    return net.IP(pkt[12:16]).String(), true
}

func (v *VNA) syncClientRoute(oldIP, newIP string) {
	
    ///===If VPN old IP exist in route table delete it===
    if oldIP != "" {
		_ = exec.Command("ip", "route", "del", oldIP+"/32", "dev", v.IfName).Run()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
    
    ///===add new route with new client's VPN IP 
	cmd := exec.CommandContext(ctx, "ip", "route", "replace", newIP+"/32", "dev", v.IfName)
	
    if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Route update failed %s: %v\n%s", newIP, err, out)
	}
}

func (v *VNA) clientCleanupLoop() {
    defer v.wg.Done()

    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            v.cleanupClients()
        case <-v.ctx.Done():
            return
        }
    }
}

func (v *VNA) cleanupClients() {
    now := time.Now()

    v.ClientsMu.Lock()
    defer v.ClientsMu.Unlock()

    for key, sess := range v.ClientByAddr {

        // handshake timeout
        if !sess.HandshakeDone && now.Sub(sess.HandshakeAt) > 5*time.Second {
            v.removeClientLocked(key, sess)
            continue
        }

        // idle timeout
        if now.Sub(sess.LastSeen) > 30*time.Second {
            v.removeClientLocked(key, sess)
            continue
        }
    }
}

func (v *VNA) removeClientLocked(key string, sess *ClientSession) {
    
    ///===Delete client from ClientByAddr map
    delete(v.ClientByAddr, key)
    
    
    if sess.VPNIP != "" {
        ///===Delete client from ClientByVPN map
        delete(v.ClientByVPN, sess.VPNIP)
        
        ///===Delete OS route with clients VPN IP
        _ = exec.Command("ip", "route", "del", sess.VPNIP+"/32", "dev", v.IfName).Run()
    }

    log.Printf("Client removed: %s", key)
}


