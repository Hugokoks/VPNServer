package vna

import (
	"VPNServer/crypted"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/curve25519"
)


func (v *VNA) Handshake(clientAddr *net.UDPAddr, clientEphPub []byte) error {
    // 1. Generuj efemerní privátní klíč serveru
    var serverEphPriv [32]byte
    if _, err := io.ReadFull(rand.Reader, serverEphPriv[:]); err != nil {
        return fmt.Errorf("generování efemerního klíče: %w", err)
    }

    // Clamping – nutný pro X25519
    serverEphPriv[0] &= 248
    serverEphPriv[31] &= 127
    serverEphPriv[31] |= 64

    // 2. Vypočítej efemerní public klíč
    var serverEphPub [32]byte
    curve25519.ScalarBaseMult(&serverEphPub, &serverEphPriv)

    // 3. Podepiš svůj efemerní pub
    signature := ed25519.Sign(v.ServerPriv, serverEphPub[:])

    // 4. Pošli odpověď
    response := append(serverEphPub[:], signature...)
    if _, err := v.Conn.WriteToUDP(response, clientAddr); err != nil {
        return fmt.Errorf("odeslání handshake odpovědi: %w", err)
    }

    // 5. ECDH – spočítej shared secret
    var clientPub [32]byte
    copy(clientPub[:], clientEphPub)
    var sharedSecret [32]byte
    curve25519.ScalarMult(&sharedSecret, &serverEphPriv, &clientPub)

    // 6. Odvoď klíč
    h := sha256.Sum256(sharedSecret[:])
    sharedKey := h[:]
    fmt.Printf("Shared key: %x", sharedKey)

    // 7. Vytvoř AEAD
    
    aead, err := crypted.NewAEAD(sharedKey)
    if err != nil {
        return fmt.Errorf("vytvoření AEAD: %w", err)
    }

    // 8. Ulož do session
    key := clientAddr.String()
    v.ClientsMu.Lock()
    if sess, ok := v.ClientByAddr[key]; ok {
        sess.Aead = aead
        sess.HandshakeDone = true
    }
    v.ClientsMu.Unlock()

    log.Printf("Handshake úspěšný s %s", clientAddr)
    return nil
}



func isHandshakePacket(pkt []byte) bool {
    return len(pkt) == 32
}