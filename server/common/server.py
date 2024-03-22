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

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while not self._terminated:
            try:
                self.__accept_new_connection()
                self.__handle_client_connection()
            except OSError as e:
                if not self._terminated:
                    raise e
        
        self._server_socket.close()
        if self._conn is not None:
            self._conn.close()

        logging.info('action: stop_server | result: success')
        

    def __handle_client_connection(self):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            bets = communication.recv_bet_batch(self._conn)
            if bets is not None:
                store_bets(bets)
                communication.send_confirmation(self._conn)
                logging.info(
                    f"action: batch_apuestas_almacenado | result: success | cantidad: {len(bets)}"
                )     
        except OSError as e:
            if e.errno in [errno.EBADF,errno.EINTR] and self._terminated:
                logging.info("action: stop_server | result: success")
            else:
                logging.error("action: receive_message | result: fail | error: {e}")
        finally:
            self._conn.close()

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
