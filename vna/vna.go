package vna

import (
	"context"
	"crypto/cipher"
	"net"
	"sync"

	"crypto/ed25519"

	"github.com/songgao/water"
)


type VNA struct {
	Iface *water.Interface

	IfName string
	IP string
	Mask string

	
	ctx context.Context
	cancel context.CancelFunc
	closeOnce sync.Once
	
	wg sync.WaitGroup
	PacketChan chan []byte

    Conn *net.UDPConn       
	
	///Clients
	ClientsMu sync.RWMutex
	ClientByAddr map[string]*ClientSession 		  //Public client IP "89.24.88.10:53000" -> sess
	ClientByVPN map[string]*ClientSession       //VPN Netowrk IP "10.0.0.10" -> sess

	LocalAddr string 
	Aead cipher.AEAD

	ServerPriv ed25519.PrivateKey
}



func New(rootCtx context.Context,ifName string,ip string,mask string,portListener string)(*VNA,error){

	cfg := water.Config{

		DeviceType: water.TUN,

	}
	cfg.Name = ifName

	iface, err := water.New(cfg)
	if err != nil{
		return nil,err
	}
	
	ctx,cancel := context.WithCancel(rootCtx)


	v := &VNA{

		Iface:      iface,
		IfName:     ifName,
		IP:         ip,
		Mask:       mask,
		ctx:        ctx,
		cancel:     cancel,
		PacketChan: make(chan []byte, 4096),

		LocalAddr: portListener,
		ClientByAddr:    make(map[string]*ClientSession),
		ClientByVPN: make(map[string]*ClientSession),
	}
	
	return v,nil
}
