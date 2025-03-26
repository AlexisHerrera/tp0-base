from io import BytesIO
import logging
import socket

from .apuesta import deserialize_apuesta

from .utils import Bet
from .packet import Packet


class Batch:
    AGENCY_NUMBER_SIZE = 4
    def __init__(self, agency_number, packets, payload_size):
        self.agency_number: int = agency_number
        self.packets: list['Packet'] = packets
        self.payload_size: int = payload_size

    def from_message_payload(payload: bytes) -> 'Batch':
        """
        Transforms a payload into a Batch object
        """
        agency_number, packets = Batch.__deserialize_batch(payload)
        return Batch(agency_number, packets, len(payload))

    def __deserialize_batch(payload: bytes) -> tuple[int, list['Packet']]:
        """
        Deserializes a batch of packets from the current packet data
        """
        stream = BytesIO(payload)
        agency_bytes = stream.read(Batch.AGENCY_NUMBER_SIZE)
        if len(agency_bytes) < Batch.AGENCY_NUMBER_SIZE:
            raise ValueError("Not enough bytes to read agency number")
        agency_number = int.from_bytes(agency_bytes, byteorder="big")

        bytesList: list['Packet'] = []
        total_length = len(payload)

        while stream.tell() < total_length:
            if total_length - stream.tell() < Packet.HEADER_SIZE:
                break
            sub_packet = Packet.read_packet(stream)
            bytesList.append(sub_packet)

        return agency_number,bytesList

    def packets_to_bets(self) -> list[Bet]:
        bets: list[Bet] = []
        for packet in self.packets:
            try:
                apuesta = deserialize_apuesta(packet.data)
                bet = Bet(str(self.agency_number), apuesta.nombre, apuesta.apellido, apuesta.documento, apuesta.nacimiento, apuesta.numero)
                bets.append(bet)
            except ValueError as e:
                logging.error(f"action: deserialize_apuesta | result: fail | error: {e}")
        return bets
