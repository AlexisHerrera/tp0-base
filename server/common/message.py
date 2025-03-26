import logging
import socket

from .utils import store_bets
from .batch import Batch
from .communication import read_exact

class Message:
    MSG_TYPE_BATCH_BET = 1
    MSG_TYPE_NOTIFICACION_FIN = 2
    MSG_TYPE_CONSULTA = 3
    MSG_TYPE_RESPUESTA = 4

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

    def process(self, addr: tuple):
        if self.msg_type == Message.MSG_TYPE_BATCH_BET:
            batch: Batch = Batch.from_message_payload(self.payload)
            logging.info(f"action: receive_message | result: success | ip: {addr[0]} | payload size: {batch.payload_size} bytes")
            logging.info(f"action: deserialize_batch | result: success | agency_number: {batch.agency_number} | packets: {len(batch.packets)}")
            bets = batch.packets_to_bets()
            if len(bets) == len(batch.packets):
                logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
            else:
                logging.error(f"action: apuesta_recibida | result: fail | cantidad: ${len(bets)}")
            store_bets(bets)
        elif self.msg_type == Message.MSG_TYPE_NOTIFICACION_FIN:
            print("Fin de notificacion")
        elif self.msg_type == Message.MSG_TYPE_CONSULTA:
            print("Consulta")
        elif self.msg_type == Message.MSG_TYPE_RESPUESTA:
            print("Respuesta")
        else:
            print("Tipo de mensaje desconocido")