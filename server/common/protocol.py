class Apuesta:
    FIELD_NOMBRE = 1
    FIELD_APELLIDO = 2
    FIELD_DOCUMENTO = 3
    FIELD_NACIMIENTO = 4
    FIELD_NUMERO = 5

    def __init__(self, nombre="", apellido="", documento="", nacimiento="", numero=""):
        self.nombre = nombre
        self.apellido = apellido
        self.documento = documento
        self.nacimiento = nacimiento
        self.numero = numero


def serialize_apuesta(apuesta: Apuesta) -> bytes:
    result = bytearray()

    def write_field(field: int, value: str):
        # Escribo el tipo
        result.append(field)
        value_bytes = value.encode("utf-8")
        length = len(value_bytes)
        # Escribo el largo
        result.extend(length.to_bytes(2, byteorder="big"))
        # Escribo el valor
        result.extend(value_bytes)

    write_field(Apuesta.FIELD_NOMBRE, apuesta.nombre)
    write_field(Apuesta.FIELD_APELLIDO, apuesta.apellido)
    write_field(Apuesta.FIELD_DOCUMENTO, apuesta.documento)
    write_field(Apuesta.FIELD_NACIMIENTO, apuesta.nacimiento)
    write_field(Apuesta.FIELD_NUMERO, apuesta.numero)

    return bytes(result)


def deserialize_apuesta(data: bytes) -> Apuesta:
    apuesta = Apuesta()
    offset = 0

    while offset < len(data):
        # Verifico que se pueda leer el tipo y el largo (tipo 1 byte y largo 2 bytes)
        if offset + 3 > len(data):
            raise ValueError("Not enough bytes to read from stream")

        # Obtengo el tipo
        field_type = data[offset]
        offset += 1

        # Obtengo la longitud, como fue escrito en big endian, lo interpreto de esa forma
        length_bytes = data[offset:offset + 2]
        offset += 2
        length = int.from_bytes(length_bytes, byteorder="big")

        # Verifico que haya suficientes datos para leer según lo esperado
        if offset + length > len(data):
            raise ValueError(f"Not enough bytes to read ${length} bytes")

        # Obtengo el valor
        value = data[offset:offset + length].decode("utf-8")
        offset += length

        # Según el field type, cargo los valores al objeto Apuesta
        if field_type == Apuesta.FIELD_NOMBRE:
            apuesta.nombre = value
        elif field_type == Apuesta.FIELD_APELLIDO:
            apuesta.apellido = value
        elif field_type == Apuesta.FIELD_DOCUMENTO:
            apuesta.documento = value
        elif field_type == Apuesta.FIELD_NACIMIENTO:
            apuesta.nacimiento = value
        elif field_type == Apuesta.FIELD_NUMERO:
            apuesta.numero = value
        else:
            raise ValueError(f"Unknown field: {field_type}")

    return apuesta
