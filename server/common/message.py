import socket

from .communication import read_exact

class Message:
    MSG_TYPE_BATCH_BET = 1
    MSG_TYPE_CONSULTA = 2
    MSG_TYPE_RESPUESTA = 3

    def __init__(self, msg_type: int, payload: bytes):
        self.msg_type = msg_type
        self.payload = payload

    @classmethod
    def read_message(cls, stream: socket.socket) -> 'Message':
        """
        Reads a message from the stream, the message is expected to have the following format:
        - 1 byte representing the message type
        - 4 bytes representing the payload length
        - The payload itself
        """
        header = read_exact(stream, 5)
        msg_type = header[0]
        payload_len = int.from_bytes(header[1:5], byteorder="big")
        payload = read_exact(stream, payload_len)
        return cls(msg_type, payload)
    
    def write_message(self, conn: socket.socket) -> None:
        """Writes the message to the socket"""
        header = bytes([self.msg_type]) + len(self.payload).to_bytes(4, byteorder="big")
        conn.sendall(header + self.payload)
        
    