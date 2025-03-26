package common

import (
	"encoding/binary"
	"net"
)

const (
	MsgTypeBatchBet        byte = 1
	MsgTypeNotificacionFin byte = 2
	MsgTypeConsulta        byte = 3
	MsgTypeRespuesta       byte = 4
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

func NewConsultaMessage() Message {
	return &BaseMessage{MsgType: MsgTypeConsulta, Payload: []byte{}}
}

func NewNotificacionFinMessage() Message {
	return &BaseMessage{MsgType: MsgTypeNotificacionFin, Payload: []byte{}}
}

func NewBatchMessage(payload []byte) Message {
	return &BaseMessage{MsgType: MsgTypeBatchBet, Payload: payload}
}
