package vna

import (
	"fmt"
	"time"
)

func (v *VNA) RunServerListener() {
	v.wg.Add(1)

	go func() {

		defer v.wg.Done()

		buf := make([]byte,65535)
	
		for {

			if v.CtxStopped() {

				return
			}

            v.Conn.SetReadDeadline(time.Now().Add(1 * time.Second)) ////wait for data max 1sec 

			////Listen for packet and connection
			n, clientAddr, err :=  v.Conn.ReadFromUDP(buf)


			if err != nil {
				continue
			}

			key := clientAddr.String()

			///lock map
			v.ClientsMu.Lock()

			////check if Client exists
			sess, ok := v.Clients[key]

			if !ok {
				////create sess with client if not
				sess = &ClientSession{Addr: clientAddr}
				
				v.Clients[key] = sess
			}

			///update sess last seen time
			sess.LastSeen = time.Now()
			v.ClientsMu.Unlock()

			////save incoming address to vna object 

			pkt := make([]byte,n)
			copy(pkt,buf[:n])

			//fmt.Printf("Listened packet %d",pkt)
			_, err = v.Iface.Write(pkt)

			if err != nil {
				fmt.Println("TUN write error:", err)

			}

		}
	}()

}