package vna

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/chacha20poly1305"
)



func newAEAD(key []byte) (cipher.AEAD, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("špatná délka klíče: očekáváno %d, má %d", chacha20poly1305.KeySize, len(key))
	}
	return chacha20poly1305.New(key)
}


// Moderní verze – stejná jako klient
func sendEncryptedTo(aead cipher.AEAD, conn *net.UDPConn, addr *net.UDPAddr, plain []byte) error {
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("generování nonce selhalo: %w", err)
	}

	// Magie – nonce jako dst → vrátí [nonce][ciphertext+tag]
	out := aead.Seal(nonce, nonce, plain, nil)

	_, err := conn.WriteToUDP(out, addr)
	return err
}

// decryptFrame 
func decryptFrame(aead cipher.AEAD, packet []byte) ([]byte, error) {
	
    nonceSize := aead.NonceSize()
	if len(packet) < nonceSize {
		return nil, fmt.Errorf("packet too short")
	}

	nonce := packet[:nonceSize]
	ciphertext := packet[nonceSize:]

	plain, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plain, nil
}