package common

import (
	"encoding/binary"
	"io"
	"net"
)

const packetHeaderSize = 4

type Packet struct {
	Data []byte
}

func NewPacket(data []byte) *Packet {
	return &Packet{Data: data}
}

func (p *Packet) Write(conn net.Conn) error {
	header := make([]byte, packetHeaderSize)
	binary.BigEndian.PutUint32(header, uint32(len(p.Data)))
	fullData := append(header, p.Data...)
	return writeFull(conn, fullData)
}

func ReadPacket(conn net.Conn) (*Packet, error) {
	header := make([]byte, packetHeaderSize)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}
	totalLength := binary.BigEndian.Uint32(header)
	data := make([]byte, totalLength)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}
	return &Packet{Data: data}, nil
}

func (p *Packet) Bytes() []byte {
	header := make([]byte, packetHeaderSize)
	binary.BigEndian.PutUint32(header, uint32(len(p.Data)))
	return append(header, p.Data...)
}

func writeFull(conn net.Conn, data []byte) error {
	total := 0
	for total < len(data) {
		n, err := conn.Write(data[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return nil
}