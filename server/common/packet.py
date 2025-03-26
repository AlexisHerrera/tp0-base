import socket
from io import BytesIO
from typing import Union
from .communication import read_exact


class Packet:
    HEADER_SIZE = 4
    def __init__(self, data: bytes):
        self.data = data

    @classmethod
    def read_packet(cls, stream: Union[socket.socket, BytesIO]) -> 'Packet':
        header = read_exact(stream, Packet.HEADER_SIZE)
        total_length = int.from_bytes(header, byteorder="big")
        data = read_exact(stream, total_length)
        return cls(data)
    
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
    