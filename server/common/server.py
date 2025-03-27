import socket
import logging
import signal

from .consulta import build_respuesta_message, parse_agency_number
from .message import Message
from .packet import *
from .utils import has_won, load_bets, store_bets
from .batch import Batch

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._alive = True
        signal.signal(signal.SIGTERM, self.exit_gracefully)
        self._threshold = 2
        self._agencies_that_finished = set()
        self._sorteo_done = False

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

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            message: Message = Message.read_message(client_sock)
            self.process(message, client_sock)
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()
            logging.info('action: client_sock close | result: success')

    def process_batch_bet(self, message: Message, client_sock):
        addr = client_sock.getpeername()
        batch: Batch = Batch.from_message_payload(message.payload)
        logging.info(f"action: receive_message | result: success | ip: {addr[0]} | payload size: {batch.payload_size} bytes")
        logging.info(f"action: deserialize_batch | result: success | agency_number: {batch.agency_number} | packets: {len(batch.packets)}")
        bets = batch.packets_to_bets()
        if len(bets) == len(batch.packets):
            logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
        else:
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: ${len(bets)}")
        store_bets(bets)
        confirmationPacket = Packet("OK\n".encode("utf-8"))
        confirmationPacket.write(client_sock)

    def process_consulta(self, message: Message, client_sock):
        addr = client_sock.getpeername()
        agency_number = parse_agency_number(message.payload)
        if agency_number is None:
            logging.error(f"action: consulta | result: fail | error: invalid agency number")
            return
        logging.info(f"action: consulta | result: success | agency_number: {agency_number} | ip: {addr[0]}")
        
        if agency_number not in self._agencies_that_finished:
            self._agencies_that_finished.add(agency_number)
        
        # Si finalizó el sorteo, hacer el sorteo
        if len(self._agencies_that_finished) >= self._threshold and not self._sorteo_done:
            logging.info("action: sorteo | result: success")
            bets = list(load_bets())
            results = {}
            for bet in bets:
                if has_won(bet):
                    agency = int(bet.agency)
                    if agency not in results:
                        results[agency] = []
                    try:
                        doc = int(bet.document)
                    except Exception as e:
                        logging.error(f"action: sorteo | result: fail | error converting dni: {e}")
                        continue
                    results[agency].append(doc)
            self._results = results
            self._sorteo_done = True
        # Si ya termino el sorteo, se envian los ganadores de la agencia, sino se envia el payload vacio
        winners = self._results.get(agency_number, []) if self._sorteo_done else []
        response = build_respuesta_message(winners)
        response.write_message(client_sock)


    def process(self, message: Message, client_sock):
        addr = client_sock.getpeername()
        if message.msg_type == Message.MSG_TYPE_BATCH_BET:
            self.process_batch_bet(message, client_sock)
        elif message.msg_type == Message.MSG_TYPE_CONSULTA:
            self.process_consulta(message, client_sock)
        else:
            logging.warning("Tipo de mensaje desconocido")
    
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
