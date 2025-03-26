from server.common.apuesta import *
import unittest

class TestProtocol(unittest.TestCase):
    def test_serialize_deserialize(self):
        original = Apuesta(
            nombre="Santiago Lionel",
            apellido="Lorca",
            documento="30904465",
            nacimiento="1999-03-17",
            numero="7574"
        )
        serialized = serialize_apuesta(original)
        deserialized = deserialize_apuesta(serialized)

        self.assertEqual(original.nombre, deserialized.nombre, "Nombre mismatch")
        self.assertEqual(original.apellido, deserialized.apellido, "Apellido mismatch")
        self.assertEqual(original.documento, deserialized.documento, "Documento mismatch")
        self.assertEqual(original.nacimiento, deserialized.nacimiento, "Nacimiento mismatch")
        self.assertEqual(original.numero, deserialized.numero, "Numero mismatch")


if __name__ == '__main__':
    unittest.main()
