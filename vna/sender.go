package vna

import "fmt"

func (v *VNA) RunServerSender() {
    v.wg.Add(1)
    go func() {
        defer v.wg.Done()

           for {

            select{
                case <- v.ctx.Done():
                    return
                case pkt, ok := <-v.PacketChan:
                    if !ok || pkt == nil {
                        return 
                    }
        
                    v.ClientsMu.RLock()
                    
                    for _, sess := range v.Clients {
                        _, err := v.Conn.WriteToUDP(pkt, sess.Addr)
                        if err != nil {
                            fmt.Println("UDP write error:", err)
                        }
                        fmt.Println("Sending:", pkt)

                    }
                    
                    v.ClientsMu.RUnlock()          
            }       
        }
    }()
}
