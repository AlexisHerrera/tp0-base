package common

import (
	"testing"
)

func TestSerializeDeserializeApuesta(t *testing.T) {
	original := Apuesta{
		Nombre:     "Santiago Lionel",
		Apellido:   "Lorca",
		Documento:  "30904465",
		Nacimiento: "1999-03-17",
		Numero:     "7574",
	}

	serialized, err := SerializeApuesta(original)
	if err != nil {
		t.Fatalf("Serialization Error: %v", err)
	}

	deserialized, err := DeserializeApuesta(serialized)
	if err != nil {
		t.Fatalf("Deserialization Error: %v", err)
	}

	if deserialized.Nombre != original.Nombre {
		t.Errorf("Nombre mismatch: expected %q, got %q", original.Nombre, deserialized.Nombre)
	}

	if deserialized.Apellido != original.Apellido {
		t.Errorf("Apellido mismatch: expected %q, got %q", original.Apellido, deserialized.Apellido)
	}

	if deserialized.Documento != original.Documento {
		t.Errorf("Documento mismatch: expected %q, got %q", original.Documento, deserialized.Documento)
	}

	if deserialized.Nacimiento != original.Nacimiento {
		t.Errorf("Nacimiento mismatch: expected %q, got %q", original.Nacimiento, deserialized.Nacimiento)
	}

	if deserialized.Numero != original.Numero {
		t.Errorf("Numero mismatch: expected %q, got %q", original.Numero, deserialized.Numero)
	}
}
