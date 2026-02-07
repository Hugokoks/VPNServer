# VPNServer

VPNServer is a custom VPN server written in Go, designed to run on Linux systems.
It manages client authentication, encrypted communication, virtual IP allocation,
and packet forwarding for connected VPN clients.

This project focuses on low-level networking, UDP-based tunneling, and
secure communication without relying on third-party VPN frameworks.

---

## Key Features

- Custom VPN server written in Go
- Linux-based deployment
- UDP-based encrypted tunnel
- Client registration and secure handshake
- Dynamic virtual IP address allocation
- Encrypted packet routing between clients and the network
- Server-side traffic forwarding using NAT

---

## Architecture Overview

The VPNServer is responsible for:

- Accepting incoming VPN client connections
- Performing encrypted handshakes (TLS-inspired, no CA)
- Registering clients and assigning virtual IP addresses
- Receiving encrypted packets from clients
- Decrypting and forwarding packets to the target network
- Encrypting and routing response traffic back to clients

---

## System Requirements

- Linux operating system
- Root privileges (required for networking and routing)
- Go installed (Go 1.20+ recommended)
- Enabled IP forwarding
- NAT (MASQUERADE) configured

---

## Network Configuration (Required)

Before running the VPN server, the Linux system must be properly configured
to forward traffic from VPN clients to the external network.

### 1. Enable IP Forwarding

Temporarily enable IP forwarding:

```bash
sudo sysctl -w net.ipv4.ip_forward=1
```
Make the change persistent by editing /etc/sysctl.conf

Apply the configuration:
```bash
sudo sysctl -p
```

```bash
net.ipv4.ip_forward=1
```

### 2. Configure NAT (MASQUERADE)

```bash
sudo iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
sudo iptables -A FORWARD -i eth0 -j ACCEPT
sudo iptables -A FORWARD -o eth0 -j ACCEPT
```


