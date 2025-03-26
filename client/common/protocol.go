package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func writeTLV(buf *bytes.Buffer, field byte, value string) error {
	// Primero escribo el tipo en el buffer como byte
	if err := buf.WriteByte(field); err != nil {
		return err
	}
	// Primero convertimos el string en un array de bytes y luego calculamos el largo y escribimos eso en el buffer
	data := []byte(value)
	if err := binary.Write(buf, binary.BigEndian, uint16(len(data))); err != nil {
		return err
	}
	// Finalmente escribimos el valor en el buffer
	if _, err := buf.Write(data); err != nil {
		return err
	}
	return nil
}

func readTLV(buf *bytes.Buffer) (FieldType, string, error) {
	// No hay nada en el buffer, retorno error.
	if buf.Len() == 0 {
		return 0, "", fmt.Errorf("buffer empty, can't deserialize")
	}
	// Obtengo el tipo
	fieldByte, err := buf.ReadByte()
	if err != nil {
		return 0, "", err
	}
	field := FieldType(fieldByte)
	// Obtengo la longitud
	var length uint16
	if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
		return field, "", err
	}
	// Creo un array de la length capturada
	valueBytes := make([]byte, length)
	bytesRead, err := buf.Read(valueBytes)
	if err != nil {
		return field, "", err
	}
	if bytesRead != int(length) {
		return field, "", fmt.Errorf("length Mismatch. Expected Length: %d, Bytes Read: %d", length, bytesRead)
	}
	return field, string(valueBytes), nil
}
