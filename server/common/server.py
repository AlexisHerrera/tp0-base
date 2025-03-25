import socket
import logging
import signal

from .utils import store_bets, Bet
from .packet import *
from .protocol import deserialize_apuesta

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._alive = True
        signal.signal(signal.SIGTERM, self.exit_gracefully)

    def exit_gracefully(self, signum, _frame):
        logging.info(f'action: SIGTERM_received | result: success | signum: {signum}')
        self._alive = False
        try:
            # Close server socket blocked in accept
            # This may leave client connections alive, so closing those file descriptors is needed
            self._server_socket.close()
            logging.info("action: socket_close | result: success")
        except OSError as e:
            logging.error("action: socket close | result: fail | error: {e}", e)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        while self._alive:
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except OSError as e:
                if not self._alive:
                    logging.info("action: loop stopped | result: success | reason: socket closed gracefully")
                    break
                else:
                    logging.error("action: loop stopped | result: fail | error: %s", e)
                    break

    @staticmethod
    def __handle_client_connection(client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            batch_packet: Packet = Packet.read_packet(client_sock)
            addr = client_sock.getpeername()
            logging.info(f"action: receive_message | result: success | ip: {addr[0]} | data received: {len(batch_packet.data)} bytes")
            agency_number, packets = batch_packet.deserialize_batch()
            logging.info(f"action: deserialize_batch | result: success | agency_number: {agency_number} | packets: {len(packets)}")
            bets = packets_to_bets(agency_number, packets)
            if len(bets) == len(packets):
                logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
            else:
                logging.error(f"action: apuesta_recibida | result: fail | cantidad: ${len(bets)}")
            store_bets(bets)
            confirmationPacket = Packet("OK\n".encode("utf-8"))
            confirmationPacket.write(client_sock)
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()
            logging.info('action: client_sock close | result: success')
    
    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c

def packets_to_bets(agency: int, packets: list[Packet]) -> list[Bet]:
    bets: list[Bet] = []
    for packet in packets:
        try:
            apuesta = deserialize_apuesta(packet.data)
            bet = Bet(str(agency), apuesta.nombre, apuesta.apellido, apuesta.documento, apuesta.nacimiento, apuesta.numero)
            bets.append(bet)
        except ValueError as e:
            logging.error(f"action: deserialize_apuesta | result: fail | error: {e}")
    return bets
