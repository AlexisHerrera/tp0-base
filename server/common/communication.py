from io import BytesIO
import socket
from typing import Union


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

