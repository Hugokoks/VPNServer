package vna

import (
	"VPNServer/crypted"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/curve25519"
)

type HandshakeInit struct {
	ClientIdentityPub ed25519.PublicKey
	ClientEphPub      [32]byte
}
func (v *VNA) Handshake(addr *net.UDPAddr, hi *HandshakeInit) error {

	// =========================================================
	// 1. Derive ClientID (FIXED SIZE, MAP SAFE)
	// =========================================================
	clientID := sha256.Sum256(hi.ClientIdentityPub) // [32]byte

	// =========================================================
	// 2. Generate server ephemeral key
	// =========================================================
	var serverEphPriv [32]byte
	if _, err := io.ReadFull(rand.Reader, serverEphPriv[:]); err != nil {
		return err
	}

	serverEphPriv[0] &= 248
	serverEphPriv[31] &= 127
	serverEphPriv[31] |= 64

	var serverEphPub [32]byte
	curve25519.ScalarBaseMult(&serverEphPub, &serverEphPriv)

	// =========================================================
	// 3. ECDH + AEAD
	// =========================================================
	var shared [32]byte
	curve25519.ScalarMult(&shared, &serverEphPriv, &hi.ClientEphPub)

	key := sha256.Sum256(shared[:])
	aead, err := crypted.NewAEAD(key[:])
	if err != nil {
		return err
	}

	// =========================================================
	// 4. SESSION CREATE / REPLACE (BY [32]byte ID)
	// =========================================================
	v.ClientsMu.Lock()

	if old, ok := v.ClientByID[clientID]; ok {
		v.removeClientLocked(old)
	}

	sess := &ClientSession{
		ID:            clientID,
		Addr:          addr,
		Aead:          aead,
		HandshakeDone: true,
		HandshakeAt:   time.Now(),
		LastSeen:      time.Now(),
	}

	v.ClientByID[clientID] = sess
	v.ClientsMu.Unlock()

	// =========================================================
	// 5. Sign ( clientID || serverEphPub )
	// =========================================================
	signedData := make([]byte, 0, 64)
	signedData = append(signedData, clientID[:]...)
	signedData = append(signedData, serverEphPub[:]...)

	sig := ed25519.Sign(v.ServerPriv, signedData)

	// =========================================================
	// 6. HandshakeRes
	// [ clientID | serverEphPub | signature ]
	// =========================================================
	payload := make([]byte, 0, 32+32+64)
	payload = append(payload, clientID[:]...)
	payload = append(payload, serverEphPub[:]...)
	payload = append(payload, sig...)

	resp := buildPacket(PacketHandshakeRes, payload)
	_, err = v.Conn.WriteToUDP(resp, addr)
	return err
}

func parseHandshakeInit(pkt []byte) (*HandshakeInit, error) {

	if len(pkt) != 128 {
		return nil, fmt.Errorf("invalid handshake size")
	}

	clientPub := pkt[0:32]
	clientEph := pkt[32:64]
	signature := pkt[64:128]

	if !ed25519.Verify(
		ed25519.PublicKey(clientPub),
		clientEph,
		signature,
	) {
		return nil, fmt.Errorf("invalid client signature")
	}

	var eph [32]byte
	copy(eph[:], clientEph)

	return &HandshakeInit{
		ClientIdentityPub: ed25519.PublicKey(clientPub),
		ClientEphPub:      eph,
	}, nil
}

