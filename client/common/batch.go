package common

import (
	"encoding/binary"
	"strconv"
)

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
