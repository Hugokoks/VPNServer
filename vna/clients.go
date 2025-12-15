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
    HandShakeDone bool
}

func (v *VNA) RegisterClientPacket(clientAddr *net.UDPAddr, plainPkt []byte) {
    key := clientAddr.String()

    v.ClientsMu.Lock()
    defer v.ClientsMu.Unlock()

    // find or create client session
    sess, ok := v.ClientByAddr[key]
    if !ok {
        sess = &ClientSession{}
        v.ClientByAddr[key] = sess
    }

    // update lastSeen and client addr
    sess.Addr = clientAddr
    sess.LastSeen = time.Now()

    // register VPN IP 
    if len(plainPkt) >= 20 && (plainPkt[0]>>4) == 4 {
        srcIP := net.IP(plainPkt[12:16]).String()

        // === NOVÁ ČÁST: přidej host route jen pokud se IP změnilo nebo je nový klient ===
        if sess.VPNIP != srcIP {
            oldIP := sess.VPNIP
            sess.VPNIP = srcIP
            v.ClientByVPN[srcIP] = sess

            // Smaž starou route (pokud existovala)
            if oldIP != "" {
                cmd := exec.Command("ip", "route", "del", oldIP+"/32", "dev", v.IfName)
                _ = cmd.Run() // klidně ignoruj chybu – route už možná neexistuje
            }

            // Přidej novou route
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            cmd := exec.CommandContext(ctx, "ip", "route", "replace", srcIP+"/32", "dev", v.IfName)
            if out, err := cmd.CombinedOutput(); err != nil {
                log.Printf("CHYBA: Nepodařilo se přidat route pro klienta %s: %v\nVýstup: %s", srcIP, err, string(out))
            } else {
                log.Printf("Route přidána: %s/32 dev %s", srcIP, v.IfName)
            }
        }
        // === KONEC NOVÉ ČÁSTI ===
    }
}