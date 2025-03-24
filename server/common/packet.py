import logging
import socket


class Packet:
    def __init__(self, data: bytes):
        self.data = data

    # Creates a Packet from a socket
    @staticmethod
    def from_socket(conn: socket.socket) -> 'Packet':
        """Expects a header of 4 bytes, then reads the expected bytes"""
        header = Packet.read_exact(conn, 4)
        total_length = int.from_bytes(header, byteorder="big")
        data = Packet.read_exact(conn, total_length)
        return Packet(data)
    
    def read_exact(conn: socket.socket, n: int) -> bytes:
        """Read exactly n bytes from a socket, this avoids short reads"""
        data = bytearray()
        while len(data) < n:
            chunk = conn.recv(n - len(data))
            if not chunk:
                raise RuntimeError("Connection closed before reading expected bytes")
            data.extend(chunk)
        return bytes(data)
    
    def write(self, conn: socket.socket) -> None:
        """Writes the packet to the socket, this avoids short writes"""
        total_sent = 0
        while total_sent < len(self.data):
            try:
                sent = conn.send(self.data[total_sent:])
                if sent == 0:
                    raise RuntimeError("Connection closed")
                total_sent += sent
            except socket.error as e:
                logging.error("Socket write error: %s", e)
                raise
