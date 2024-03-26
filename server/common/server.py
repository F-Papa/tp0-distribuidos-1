import os
import socket
import logging
import errno
from . import communication
from .utils import *

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._conn = None
        self._terminated = False
        self._clients_finished = {}
        self._winning_bets_list = []
        for i in range(1,6):
            self._clients_finished[i] = False

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        self.__accept_new_connection()
        while not self._terminated:
            try:
                self.__handle_client_connection()
            except OSError as e:
                if not self._terminated:
                    raise e
        
        self._server_socket.close()
        if self._conn is not None:
            self._conn.close()

        logging.info('action: stop_server | result: success')

    def __results_ready(self) -> bool:
        return all(self._clients_finished.values())    

    def _winning_bets(self) -> list[Bet]:
        if self._winning_bets_list:
            return self._winning_bets_list
        else:
            self._winning_bets_list = [bet for bet in load_bets() if has_won(bet)]
            return self._winning_bets_list

    def __handle_client_connection(self):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            message: communication.Message = communication.recv_message(self._conn)
            if not message:
                self._conn.close()
                return
            self.__process_message(message)
        except OSError as e:
            if e.errno in [errno.EBADF,errno.EINTR] and self._terminated:
                logging.info("action: stop_server | result: success")
            else:
                raise e

    def __process_message(self, message: communication.Message):
        # Bet message
        if message.is_bet():
            logging.debug(f"action: processing_message | agency: {message.agency()} | result: in_progress | type: bet")
            bets = message.bets()
            store_bets(bets)
            communication.send_confirmation(self._conn)
            logging.info(
                f"action: batch_apuestas_almacenado | agency: {message.agency()} | result: success | cantidad: {len(bets)}"
            )

        # Finished message
        elif message.is_finished():
            logging.debug(f"action: processing_message | agency: {message.agency()} | result: in_progress | type: finished")
            self._clients_finished[message.agency_id] = True
            if self.__results_ready():
                logging.info(
                    f"action: sorteo | result: success | agency: {message.agency()} | cant_ganadores: {len(self._winning_bets())}"
                ) 

        # Consult winners message
        elif message.is_consult_winners():
            logging.debug(f"action: processing_message | agency: {message.agency()} | result: in_progress | type: consult_winners")
            if self.__results_ready():
                winning_bets = self._winning_bets()
                winning_documents_for_agency = [bet.document for bet in winning_bets if bet.agency == message.agency()]
                communication.send_winners(self._conn, winning_documents_for_agency)
                logging.info(
                    f"action: winners_sent | agency: {message.agency()} | result: success | cantidad: {len(winning_documents_for_agency)}"
                )
            else:
                communication.send_wait(self._conn)

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
        self._conn = c

    def stop(self):
        # Set the server to terminated state, so it won't keep looping
        logging.info('action: stop_server | result: in_progress')
        self._terminated = True
        self._server_socket.shutdown(socket.SHUT_RDWR)
        if self._conn is not None:
            self._conn.shutdown(socket.SHUT_RDWR)
