import socket
from io import BytesIO
from typing import Union


class Packet:
    HEADER_SIZE = 4
    def __init__(self, data: bytes):
        self.data = data

    @classmethod
    def read_packet(cls, stream: Union[socket.socket, BytesIO]) -> 'Packet':
        header = cls.read_exact(stream, Packet.HEADER_SIZE)
        total_length = int.from_bytes(header, byteorder="big")
        data = cls.read_exact(stream, total_length)
        return cls(data)

    @staticmethod
    def read_exact(source: Union[socket.socket, BytesIO], n: int) -> bytes:
        data = bytearray()
        if hasattr(source, "recv"):
            while len(data) < n:
                chunk = source.recv(n - len(data))
                if not chunk:
                    raise RuntimeError("Connection closed before reading expected bytes")
                data.extend(chunk)
        else:
            while len(data) < n:
                chunk = source.read(n - len(data))
                if not chunk:
                    raise RuntimeError("EOF reached before reading expected bytes")
                data.extend(chunk)
        return bytes(data)
    
    def write(self, conn: socket.socket) -> None:
        """Writes the packet to the socket, this avoids short writes"""
        header = len(self.data).to_bytes(Packet.HEADER_SIZE, byteorder="big")
        full_data = header + self.data
        total_sent = 0
        while total_sent < len(full_data):
            sent = conn.send(full_data[total_sent:])
            if sent == 0:
                raise RuntimeError("Connection closed")
            total_sent += sent
    