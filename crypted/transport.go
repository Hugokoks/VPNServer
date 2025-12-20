package crypted

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"net"
)

func SendEncryptedTo(aead cipher.AEAD, conn *net.UDPConn, addr *net.UDPAddr, plain []byte)  error {
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("generování nonce selhalo: %w", err)
	}

	//nonce dst → return [nonce][ciphertext+tag]
	out := aead.Seal(nonce, nonce, plain, nil)

	_, err := conn.WriteToUDP(out, addr)
	return err
}

func DecryptFrame(aead cipher.AEAD, packet []byte) ([]byte, error) {
	
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