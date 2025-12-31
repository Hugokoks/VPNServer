package vna



func (v * VNA)Start(){
	
	v.wg.Add(1)
	go v.runReader()
	
	v.wg.Add(1)
	go v.runServerListener()
	
	v.wg.Add(1)
	go v.runServerSender()         

    v.wg.Add(1)
    go v.clientCleanupLoop()
    
    v.wg.Add(1)
    go v.ipPoolCleanupLoop()

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