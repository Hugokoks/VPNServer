package vna

import "net"



func(v *VNA) InitConnection()error{

	laddr,err := net.ResolveUDPAddr("udp",v.LocalAddr)


	if err != nil{

		return err
	}

	conn,err := net.ListenUDP("udp",laddr)

	if err != nil{

		return  err

	}
	v.Conn = conn
	return  nil

}