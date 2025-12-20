package vna

import (
	"fmt"
	"strings"
)


func (v *VNA) runReader() {
		
    ////check if this work group is done
    defer v.wg.Done()
    
    buf := make([]byte,65535)
    for {
        if v.CtxStopped() {
			return
		}
    
		//read packet's from interface
        n, err := v.Iface.Read(buf)
        
        if err != nil {
            
            if v.CtxStopped() || strings.Contains(err.Error(), "file already closed") {
                return
            }
            fmt.Println("read error:", err)
            continue
        }

		pkt := make([]byte,n)
		copy(pkt,buf[:n])
        
		//fmt.Printf("\nrecieved packet %d",pkt)
        v.PacketChan <- pkt
    }
}
