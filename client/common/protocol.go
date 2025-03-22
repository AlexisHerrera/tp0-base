package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type FieldType byte

const (
	FieldNombre FieldType = iota + 1
	FieldApellido
	FieldDocumento
	FieldNacimiento
	FieldNumero
)

type Apuesta struct {
	Nombre     string
	Apellido   string
	Documento  string
	Nacimiento string
	Numero     string
}

var fieldTypeName = map[FieldType]string{
	FieldNombre:     "Nombre",
	FieldApellido:   "Apellido",
	FieldDocumento:  "Documento",
	FieldNacimiento: "Nacimiento",
	FieldNumero:     "Numero",
}

// Implementación de FieldType a String
func (ft FieldType) String() string {
	if name, ok := fieldTypeName[ft]; ok {
		return name
	}
	return fmt.Sprintf("FieldType(%d)", ft)
}

func SerializeApuesta(apuesta Apuesta) ([]byte, error) {
	buf := new(bytes.Buffer)

	writeTLV := func(field FieldType, value string) error {
		// Primero escribo el tipo en el buffer como byte
		if err := buf.WriteByte(byte(field)); err != nil {
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

	// Escribimos cada campo de la apuesta usando la función auxiliar
	if err := writeTLV(FieldNombre, apuesta.Nombre); err != nil {
		return nil, fmt.Errorf("WriteTLV Error, Type: %s, Value: %s, Error: %w", FieldNombre, apuesta.Nombre, err)
	}
	if err := writeTLV(FieldApellido, apuesta.Apellido); err != nil {
		return nil, fmt.Errorf("WriteTLV Error, Type: %s, Value: %s, Error: %w", FieldApellido, apuesta.Apellido, err)
	}
	if err := writeTLV(FieldDocumento, apuesta.Documento); err != nil {
		return nil, fmt.Errorf("WriteTLV Error, Type: %s, Value: %s, Error: %w", FieldDocumento, apuesta.Documento, err)
	}
	if err := writeTLV(FieldNacimiento, apuesta.Nacimiento); err != nil {
		return nil, fmt.Errorf("WriteTLV Error, Type: %s, Value: %s, Error: %w", FieldNacimiento, apuesta.Nacimiento, err)
	}
	if err := writeTLV(FieldNumero, apuesta.Numero); err != nil {
		return nil, fmt.Errorf("WriteTLV Error, Type: %s, Value: %s, Error: %w", FieldNumero, apuesta.Numero, err)
	}

	return buf.Bytes(), nil
}

func DeserializeApuesta(data []byte) (Apuesta, error) {
	var apuesta Apuesta
	buf := bytes.NewBuffer(data)

	readTLV := func() (FieldType, string, error) {
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

	for buf.Len() > 0 {
		field, value, err := readTLV()
		if err != nil {
			return apuesta, err
		}
		// Se va cargando el struct a medida que se lea cada (Type, value)
		switch field {
		case FieldNombre:
			apuesta.Nombre = value
		case FieldApellido:
			apuesta.Apellido = value
		case FieldDocumento:
			apuesta.Documento = value
		case FieldNacimiento:
			apuesta.Nacimiento = value
		case FieldNumero:
			apuesta.Numero = value
		default:
			return apuesta, fmt.Errorf("unknown: %d", field)
		}
	}

	return apuesta, nil
}
