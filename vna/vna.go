package vna

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/songgao/water"
)

type ClientSession struct{
	Addr 	*net.UDPAddr
	LastSeen time.Time

}


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
    Clients map[string]*ClientSession
	ClientsMu sync.RWMutex
	
	LocalAddr string ////where server listening
}



func New(rootCtx context.Context,ifName string,ip string,mask string)(*VNA,error){

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

		LocalAddr: ":5000",
		Clients:    make(map[string]*ClientSession),


	}

	return v,nil
}

func (v * VNA)Start(){

	v.RunReader()
	v.RunServerListener()
	v.RunServerSender()         

}

func (v * VNA)Stop(){

	v.Close()

}

func (v * VNA)Close(){
   v.closeOnce.Do(func() {
        v.cancel()

        if v.Conn != nil {
            _ = v.Conn.Close()  
        }

        if v.Iface != nil {
            _ = v.Iface.Close() 
        }

        v.wg.Wait()
    })
}