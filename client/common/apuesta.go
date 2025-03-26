package common

import (
	"bytes"
	"fmt"
	"strings"
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
	// Escribimos cada campo de la apuesta usando la función auxiliar
	if err := writeTLV(buf, byte(FieldNombre), apuesta.Nombre); err != nil {
		return nil, fmt.Errorf("WriteTLV Error, Type: %s, Value: %s, Error: %w", FieldNombre, apuesta.Nombre, err)
	}
	if err := writeTLV(buf, byte(FieldApellido), apuesta.Apellido); err != nil {
		return nil, fmt.Errorf("WriteTLV Error, Type: %s, Value: %s, Error: %w", FieldApellido, apuesta.Apellido, err)
	}
	if err := writeTLV(buf, byte(FieldDocumento), apuesta.Documento); err != nil {
		return nil, fmt.Errorf("WriteTLV Error, Type: %s, Value: %s, Error: %w", FieldDocumento, apuesta.Documento, err)
	}
	if err := writeTLV(buf, byte(FieldNacimiento), apuesta.Nacimiento); err != nil {
		return nil, fmt.Errorf("WriteTLV Error, Type: %s, Value: %s, Error: %w", FieldNacimiento, apuesta.Nacimiento, err)
	}
	if err := writeTLV(buf, byte(FieldNumero), apuesta.Numero); err != nil {
		return nil, fmt.Errorf("WriteTLV Error, Type: %s, Value: %s, Error: %w", FieldNumero, apuesta.Numero, err)
	}

	return buf.Bytes(), nil
}

// Reads a CSV line and returns an Apuesta struct
func ApuestaFromCSVLine(line string) (apuesta Apuesta, err error) {
	fields := strings.Split(line, ",")
	if len(fields) < 5 {
		err = fmt.Errorf("invalid line: %s", line)
		return
	}
	apuesta = Apuesta{
		Nombre:     strings.TrimSpace(fields[0]),
		Apellido:   strings.TrimSpace(fields[1]),
		Documento:  strings.TrimSpace(fields[2]),
		Nacimiento: strings.TrimSpace(fields[3]),
		Numero:     strings.TrimSpace(fields[4]),
	}
	return apuesta, nil
}

func DeserializeApuesta(data []byte) (Apuesta, error) {
	var apuesta Apuesta
	buf := bytes.NewBuffer(data)
	for buf.Len() > 0 {
		field, value, err := readTLV(buf)
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
