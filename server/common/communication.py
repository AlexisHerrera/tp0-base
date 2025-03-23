import logging
import socket


def read_exact(conn: socket.socket, n: int) -> bytes:
    data = bytearray()
    while len(data) < n:
        chunk = conn.recv(n - len(data))
        if not chunk:
            raise RuntimeError("Connection closed before reading expected bytes")
        data.extend(chunk)
    return bytes(data)


def read_bet_as_bytes(conn: socket.socket) -> bytes:
    """Expects a header of 4 bytes, then reads the expected bytes"""
    header = read_exact(conn, 4)
    total_length = int.from_bytes(header, byteorder="big")
    data = read_exact(conn, total_length)
    return data


def write_full(conn: socket.socket, data: bytes) -> None:
    total_sent = 0
    while total_sent < len(data):
        try:
            sent = conn.send(data[total_sent:])
            if sent == 0:
                raise RuntimeError("Connection closed")
            total_sent += sent
        except socket.error as e:
            logging.error("Socket write error: %s", e)
            raise

