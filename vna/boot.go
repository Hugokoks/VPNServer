package vna

import "fmt"



func (v *VNA)Boot() error{

	
    if err := v.InitConnection(); err != nil{

        return fmt.Errorf("failed to Create UDP Listener %v",err)

    }

	if err := v.LoadPriveServerKey(); err != nil{


        return fmt.Errorf("load priv key: %w", err)
    }

	if err := v.SetupAdapter();err != nil{

		return  fmt.Errorf("failed to setup vpn_adapter %w",err)

	}

	return nil
}