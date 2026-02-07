# VPNServer

VPNServer is a custom VPN server implementation written in Go.  
It is designed as a low-level networking project to explore secure tunneling,
UDP-based communication, and protocol design.

The server handles client registration, encrypted session establishment,
and packet routing between connected clients.  
It is intended primarily for Linux environments.

This project focuses on understanding how VPN-like systems work internally,
rather than providing a production-ready VPN solution.

---

## Key Features

- Custom VPN server written in Go
- UDP-based data transport
- Encrypted handshake inspired by TLS mechanisms (without certificate authority)
- Client registration and authentication
- Dynamic IP address allocation managed by the server
- Secure packet forwarding between VPN clients
- Linux-focused networking (TUN/TAP)

---

## Architecture Overview

The VPNServer is responsible for:

- Accepting incoming client connections
- Performing encrypted handshakes
- Assigning virtual IP addresses to clients
- Managing active VPN sessions
- Routing encrypted packets between connected clients
- Interfacing with the system network stack using TUN devices

The protocol design and encryption logic are intentionally kept explicit
to support learning, experimentation, and extensibility.

---

## Usage

> ⚠️ Requires Linux and access to TUN/TAP devices.

Basic example:

```bash
go run main.go
