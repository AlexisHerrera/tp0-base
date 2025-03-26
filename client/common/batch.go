package common

import (
	"bufio"
	"encoding/binary"
	"net"
	"strconv"
)

// 8kb - 4 bytes of header - 4 bytes of agency ID
const agencyIDSize = 4
const maxPayloadSize = 8*1024 - agencyIDSize - packetHeaderSize

type Batch struct {
	AgencyID string
	Count    int
	Payload  []byte
}

func (b *Batch) GetCount() int {
	return b.Count
}

func (b *Batch) Serialize() []byte {
	idInt, err := strconv.Atoi(b.AgencyID)
	if err != nil {
		idInt = 0
	}
	agencyBytes := make([]byte, agencyIDSize)
	binary.BigEndian.PutUint32(agencyBytes, uint32(idInt))
	finalPayload := append(agencyBytes, b.Payload...)
	// Wraps it in a packet just to know how long the batch will be
	packet := NewPacket(finalPayload)
	return packet.Serialize()
}

func NewBatchPacket(agency string, payload []byte) *Packet {
	idInt, err := strconv.Atoi(agency)
	if err != nil {
		idInt = 0
	}
	// Just appends the agency ID to the payload
	agencyBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(agencyBytes, uint32(idInt))
	finalPayload := append(agencyBytes, payload...)
	return NewPacket(finalPayload)
}

func (b *Batch) Write(conn net.Conn) error {
	data := b.Serialize()
	return writeFull(conn, data)
}

func ReadBatch(scanner *bufio.Scanner, config ClientConfig, leftover *string) (*Batch, error) {
	var payload []byte
	count := 0
	// If there is an Apuesta leftover, then process it first
	if *leftover != "" {
		line := *leftover
		*leftover = ""
		apuesta, err := ApuestaFromCSVLine(line)
		if err != nil {
			return nil, err
		}
		serialized, err := SerializeApuesta(apuesta)
		if err != nil {
			return nil, err
		}
		subpacket := NewPacket(serialized)
		subPacketBytes := subpacket.Serialize()
		// There is no need to check if the payload is too big, as it was already checked before
		payload = append(payload, subPacketBytes...)
		count++
	}
	// Reads the rest of the lines
	for count < config.BatchMaxAmount && scanner.Scan() {
		line := scanner.Text()
		apuesta, err := ApuestaFromCSVLine(line)
		if err != nil {
			return nil, err
		}
		serialized, err := SerializeApuesta(apuesta)
		if err != nil {
			return nil, err
		}
		if len(serialized) > maxPayloadSize {
			log.Errorf("action: serialize_apuesta | result: skip | error: apuesta exceeds max payload size: %v", line)
			continue
		}
		subpacket := NewPacket(serialized)
		subPacketBytes := subpacket.Serialize()
		// Checks if the payload is too big
		if len(payload)+len(subPacketBytes) <= maxPayloadSize {
			payload = append(payload, subPacketBytes...)
			count++
		} else {
			*leftover = line
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &Batch{
		AgencyID: config.ID,
		Count:    count,
		Payload:  payload,
	}, nil
}
