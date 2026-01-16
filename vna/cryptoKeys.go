package vna

import (
	"encoding/base64"
	"fmt"
	"os"
)

func (v *VNA) LoadPriveServerKey()error{


	serverPrivB64 := os.Getenv("SERVER_PRIVATE_KEY")

	if serverPrivB64 == ""{
		return fmt.Errorf("SERVER_PRIVATE_KEY is not set")
	}

	privBytes,err := base64.StdEncoding.DecodeString(serverPrivB64)

	if err != nil{

		return fmt.Errorf("wrong SERVER_PRIVATE_KEY: %w", err)
	}
	

	v.ServerPriv = privBytes

	return  nil

}