package common

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
)

const (
	MsgTypeBatchBet        byte = 1
	MsgTypeConsulta        byte = 2
	MsgTypeRespuestaWinner byte = 3
	MsgTypeRespuestaWait   byte = 4
)

type Message interface {
	Serialize() []byte
	Write(conn net.Conn) error
}

// Al serializarse es el header + payload (header = 1 byte + 4 bytes de longitud)
type BaseMessage struct {
	MsgType byte
	Payload []byte
}

func (m *BaseMessage) Serialize() []byte {
	payloadLen := uint32(len(m.Payload))
	header := make([]byte, 5)
	header[0] = m.MsgType
	binary.BigEndian.PutUint32(header[1:], payloadLen)
	return append(header, m.Payload...)
}

func (m *BaseMessage) Write(conn net.Conn) error {
	data := m.Serialize()
	return writeFull(conn, data)
}

func ReadMessage(conn net.Conn) (Message, error) {
	header := make([]byte, 5)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}
	msgType := header[0]
	payloadLen := binary.BigEndian.Uint32(header[1:])
	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return nil, err
	}
	return &BaseMessage{MsgType: msgType, Payload: payload}, nil
}

func NewConsultaMessage(agencyId string) Message {
	idInt, err := strconv.Atoi(agencyId)
	if err != nil {
		idInt = 0
	}
	agencyBytes := make([]byte, agencyIDSize)
	binary.BigEndian.PutUint32(agencyBytes, uint32(idInt))
	return &BaseMessage{MsgType: MsgTypeConsulta, Payload: agencyBytes}
}

// El payload de la respuesta es una lista de enteros de 4 bytes
func ParseRespuestaPayload(payload []byte) ([]int, error) {
	if len(payload) == 0 {
		return []int{}, nil
	}
	if len(payload)%4 != 0 {
		return nil, fmt.Errorf("invalid payload length for respuesta message")
	}
	count := len(payload) / 4
	winners := make([]int, count)
	for i := 0; i < count; i++ {
		offset := i * 4
		winners[i] = int(binary.BigEndian.Uint32(payload[offset : offset+4]))
	}
	return winners, nil
}

func NewBatchMessage(payload []byte) Message {
	return &BaseMessage{MsgType: MsgTypeBatchBet, Payload: payload}
}
