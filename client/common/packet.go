package common

import (
	"encoding/binary"
	"net"
)

type Packet struct {
	Data []byte
}

func NewPacket(data []byte) *Packet {
	return &Packet{Data: data}
}

func (p *Packet) Write(conn net.Conn) error {
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(len(p.Data)))
	fullData := append(header, p.Data...)
	return writeFull(conn, fullData)
}

func ReadPacket(conn net.Conn) (*Packet, error) {
	header := make([]byte, 4)
	// Usamos readFull para leer exactamente 4 bytes del header.
	if err := readFull(conn, header); err != nil {
		return nil, err
	}
	totalLength := binary.BigEndian.Uint32(header)
	data := make([]byte, totalLength)
	// Leemos exactamente totalLength bytes.
	if err := readFull(conn, data); err != nil {
		return nil, err
	}
	return &Packet{Data: data}, nil
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
