// sender.go
package vna



func (v *VNA) RunServerSender() {
    v.wg.Add(1)
    go func() {
        defer v.wg.Done()

        for {
            select {
            case <-v.ctx.Done():
                return

            case pkt, ok := <-v.PacketChan:
                if !ok || pkt == nil {
                    return
                }

                // z tunelu čekáme IP pakety → pošli je do routeru
                v.routeToTun(pkt)
               
            }
        }
    }()
}
