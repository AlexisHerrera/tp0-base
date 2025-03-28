package common

import (
	"encoding/binary"
	"net"
)

const packetHeaderSize = 4

type Packet struct {
	Data []byte
}

func NewPacket(data []byte) *Packet {
	// A Packet just wraps a byte slice
	return &Packet{Data: data}
}

func (p *Packet) Write(conn net.Conn) error {
	fullData := p.Serialize()
	return writeFull(conn, fullData)
}

func ReadPacket(conn net.Conn) (*Packet, error) {
	// First read the length of the data
	// Then read the data itself
	header := make([]byte, packetHeaderSize)
	if err := readFull(conn, header); err != nil {
		return nil, err
	}
	totalLength := binary.BigEndian.Uint32(header)
	data := make([]byte, totalLength)
	// ReadFull avoids short reads
	if err := readFull(conn, data); err != nil {
		return nil, err
	}
	return &Packet{Data: data}, nil
}

func (p *Packet) Serialize() []byte {
	// When writing a packet, we first write the length of the data, then the data itself
	// So when reading a packet, we first read the length, then the data
	header := make([]byte, packetHeaderSize)
	binary.BigEndian.PutUint32(header, uint32(len(p.Data)))
	return append(header, p.Data...)
}

func writeFull(conn net.Conn, data []byte) error {
	// Write all the data to the connection, avoids short writes
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

func readFull(conn net.Conn, data []byte) error {
	// Read all the data from the connection, avoids short reads
	total := 0
	for total < len(data) {
		n, err := conn.Read(data[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return nil
}
