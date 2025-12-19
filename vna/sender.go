package vna



func (v *VNA) runServerSender() {
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
}
