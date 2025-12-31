package vna

type PacketType byte

const (
	PacketIPRequest PacketType = 1
	PacketIPResponse PacketType = 2
	PacketHandshake PacketType = 3
	PacketData      PacketType = 4

)


