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

func (b *Batch) GetFinalPayloadInBytes() int {
	return len(b.Payload) + agencyIDSize
}

func (b *Batch) Serialize() []byte {
	idInt, err := strconv.Atoi(b.AgencyID)
	if err != nil {
		idInt = 0
	}
	agencyBytes := make([]byte, agencyIDSize)
	binary.BigEndian.PutUint32(agencyBytes, uint32(idInt))
	batchContent := append(agencyBytes, b.Payload...)
	batchMessage := NewBatchMessage(batchContent)
	// 1 byte for the message type, 4 bytes for the payload length, and the payload itself
	return batchMessage.Serialize()
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
		// Check if the payload is too big, if it is, then discard the leftover
		if len(subPacketBytes) > maxPayloadSize {
			log.Errorf("action: serialize_apuesta | result: skip | error: apuesta exceeds max payload size: %v", line)
		} else {
			payload = append(payload, subPacketBytes...)
			count++
		}
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
