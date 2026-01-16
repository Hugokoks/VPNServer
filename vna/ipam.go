package vna

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)



type IPPool struct {
	network net.IPNet
	used    map[string]string // in future use clients key hash as values
	reserved map[string]time.Time
	mu      sync.Mutex
	gateway  net.IP


}

func NewIPPool(cidr,gateway string) (*IPPool, error) {
	_, netw, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	return &IPPool{
		network: *netw,
		used:    make(map[string]string),
        reserved: make(map[string]time.Time),
		gateway: net.ParseIP(gateway),
	}, nil
}

func (p *IPPool) reserveIP(ttl time.Duration) (net.IP, error) {
    p.mu.Lock()
    defer p.mu.Unlock()

    now := time.Now()

    ////set expire time for reserved ip 
    expires := now.Add(ttl)

    /////Do cycle if pool contains current ip
    for ip := p.firstIP(); p.network.Contains(ip); incIP(ip) {
        
        ////if current ip is gateway
        if ip.Equal(p.gateway) {
            continue
        }

        ipStr := ip.String()

        ////if current ip already exist in used ip's
        if _, ok := p.used[ipStr]; ok {
            continue
        }

        ///if current ip is already in reserved list 
        if _, ok := p.reserved[ipStr]; ok {
            continue
        }

        p.reserved[ipStr] = expires
        return ip, nil
    }

    return nil, fmt.Errorf("IP pool exhausted")
}

/////Commit is trigger after successfull handshake with client
func (p *IPPool) Commit(ip string, sessionID string) error {
    p.mu.Lock()
    defer p.mu.Unlock()

    ////if clients try to connect to server and reserved ip doesn't exist
    if _, ok := p.reserved[ip]; !ok {
        return fmt.Errorf("IP not reserved")
    }

    ////free reserved ip
    delete(p.reserved, ip)
    ////save ip to used IP Pool 
    p.used[ip] = sessionID
    return nil
}



func (p *IPPool) Release(ip string) {
    p.mu.Lock()
    defer p.mu.Unlock()
   	delete(p.used, ip)
    delete(p.reserved, ip)
}


/////take first ip from IP range 10.0.0.1
func (p *IPPool) firstIP() net.IP {
    ip := make(net.IP, len(p.network.IP))
    copy(ip, p.network.IP)
    incIP(ip) 
    return ip
}

////increace ip by one host 10.0.0.1 -> 10.0.0.2 ...
func incIP(ip net.IP) {
    for i := len(ip) - 1; i >= 0; i-- {
        ip[i]++
        if ip[i] != 0 {
            break
        }
    }
}

////sending Reserved ip to client then he will use it to connect
func (v *VNA) sendIPResponse(addr *net.UDPAddr, ip net.IP, mask net.IP) {

    /*
    [0] = PacketIPResponse
    [1] = IP oktet 1
    [2] = IP oktet 2
    [3] = IP oktet 3
    [4] = IP oktet 4
    [5] = MASK oktet 1
    [6] = MASK oktet 2
    [7] = MASK oktet 3
    [8] = MASK oktet 4
    */
 

    payload := make([]byte, 8)
    
    copy(payload[0:4], ip.To4())
    copy(payload[4:8], mask.To4())

    pkt := buildPacket(PacketIPResponse,payload)

    if _, err := v.Conn.WriteToUDP(pkt, addr); err != nil {
        log.Printf("Failed to send IP response to %s: %v", addr, err)
    }
}

////cleanup reserved ip's in ip pool
func (v *VNA) ipPoolCleanupLoop() {
    defer v.wg.Done()

    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            v.IPPool.CleanupExpiredReserved()
        case <-v.ctx.Done():
            return
        }
    }
}

func (p *IPPool) CleanupExpiredReserved() {
    now := time.Now()

    p.mu.Lock()
    defer p.mu.Unlock()

    for ip, expires := range p.reserved {
        if now.After(expires) {
            delete(p.reserved, ip)
        }
    }
}