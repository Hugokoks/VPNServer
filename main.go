package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"VPNServer/vna"

	"github.com/joho/godotenv"
)


func main(){


	rootCtx, rootCancel := context.WithCancel(context.Background())

	_ = godotenv.Load() 

	sigs := make(chan os.Signal, 1)
	
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)


	/////Goruntine waiting for Ctrl+C signal then cancel context (end program)
	go func(){
		<-sigs
		log.Println("signal recieved, canceling root context")
		rootCancel()


	}()
		
	////creating VNA 
	vna,err := vna.New(rootCtx,"vpn0","10.0.0.1","255.255.255.255",":5000")

	if err != nil {
		log.Printf("failed to create VNA:%v",err)
		rootCancel()
		return
	}

	/////Stop vna when main ends 
	defer vna.Stop()

	log.Println("virtual network adapter created")
	
	////add ip mask etc... to vna
	if err := vna.SetupAdapter(); err != nil {
    
		log.Printf("failed to setup adapter: %v", err)
    	rootCancel()
    	return
	
	}
	
	log.Println("virtual network setup successfully")
	
	vna.Start()

	log.Println("Ctrl+C for stopping")
	<-rootCtx.Done() 
	log.Println("main exising...")


}