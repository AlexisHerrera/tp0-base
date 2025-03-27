from typing import Union
from .message import Message

DEFAULT_AGENCY_NUMBER_SIZE = 4

def parse_agency_number(payload: bytes, size: int = DEFAULT_AGENCY_NUMBER_SIZE):
    if len(payload) < size:
        return None
    return int.from_bytes(payload[:size], byteorder="big")

def build_respuesta_message(winners: Union[list[int], None]) -> Message:
    """ Si se recibe None, es porque todavía no terminó. Si se recibe una lista es porque son los ganadores """
    if winners is not None:
        payload = b"".join([w.to_bytes(4, byteorder="big") for w in winners])
        return Message(Message.MSG_TYPE_RESPUESTA_GANADOR, payload)
    return Message(Message.MSG_TYPE_RESPUESTA_WAIT, b"")
