from .message import Message

DEFAULT_AGENCY_NUMBER_SIZE = 4

def parse_agency_number(payload: bytes, size: int = DEFAULT_AGENCY_NUMBER_SIZE):
    if len(payload) < size:
        return None
    return int.from_bytes(payload[:size], byteorder="big")

def build_respuesta_message(winners: list[int]) -> Message:
    """ Si no finaliza el sorteo se envía un payload vacío, sino con la lista de ganadores """
    payload = b""
    if winners:
        payload = b"".join([w.to_bytes(4, byteorder="big") for w in winners])
    return Message(Message.MSG_TYPE_RESPUESTA, payload)
