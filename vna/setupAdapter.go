package vna

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"
)

// helper: parse "255.255.255.255" -> 32
func (v *VNA) maskPrefix() (int, error) {
    ip := net.ParseIP(v.Mask)
    if ip == nil {
        return 0, fmt.Errorf("invalid netmask: %s", v.Mask)
    }
    ip = ip.To4()
    if ip == nil {
        return 0, fmt.Errorf("netmask is not IPv4: %s", v.Mask)
    }

    m := net.IPv4Mask(ip[0], ip[1], ip[2], ip[3])
    ones, bits := m.Size()
    if bits != 32 {
        return 0, fmt.Errorf("unexpected mask size: %d", bits)
    }
    return ones, nil
}

// SetupAdapter set IP and turn on interface
func (v *VNA) SetupAdapter() error {
    prefix, err := v.maskPrefix()
    if err != nil {
        return err
    }

    cidr := fmt.Sprintf("%s/%d", v.IP, prefix)

    ctx, cancel := context.WithTimeout(v.ctx, 5*time.Second)
    defer cancel()

    // ip addr add 10.0.0.2/24 dev vpn0
    cmdAddr := exec.CommandContext(ctx,
        "ip", "addr", "add", cidr, "dev", v.IfName,
    )

    if out, err := cmdAddr.CombinedOutput(); err != nil {
        return fmt.Errorf("ip addr add failed: %v (out: %s)", err, string(out))
    }

    // ip link set dev vpn0 up
    cmdUp := exec.CommandContext(ctx,
        "ip", "link", "set", "dev", v.IfName, "up",
    )

    if out, err := cmdUp.CombinedOutput(); err != nil {
        return fmt.Errorf("ip link set up failed: %v (out: %s)", err, string(out))
    }

    mtuCmd := exec.CommandContext(ctx, "ip", "link", "set", "dev", v.IfName, "mtu", "1400")
    if out, err := mtuCmd.CombinedOutput(); err != nil {
        log.Printf("Varování: Nepodařilo se nastavit MTU na %s: %v (výstup: %s)", v.IfName, err, string(out))
    } else {
        log.Printf("MTU na %s nastaveno na 1400", v.IfName)
    }

    if err := v.LoadPriveServerKey(); err != nil{


        return fmt.Errorf("načtení server priv key: %w", err)
    }

    if err := v.InitConnection(); err != nil{

        return fmt.Errorf("Failed to Create Listener %v",err)

    }

    return nil
}
