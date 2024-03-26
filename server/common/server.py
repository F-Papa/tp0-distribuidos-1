import os
import socket
import logging
import errno
import threading
from . import communication
from .utils import *

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.registed_connections = {}
        self.unregistered_connections = {}
        self.handles = []
        self._terminated = False
        self._clients_finished = {}
        self._winning_bets_list = []
        for i in range(1,2):
            self._clients_finished[i] = False
        self._results_condition = threading.Condition()
        self._bets_lock = threading.Lock()
        self._connections_lock = threading.Lock()


    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while not self._terminated:
            try:
                sock, addr = self.__accept_new_connection()
                for thread in self.handles:
                    if not thread.is_alive():
                        thread.join()
                        self.handles.remove(thread)

                thread = threading.Thread(target=self.__handle_new_connection, args=(sock, addr))
                thread.start()
                self.handles.append(thread)
            except OSError as e:
                if not self._terminated:
                    raise e

        self._server_socket.close()
        
        logging.info('action: stop_server | result: finishing')

        if not self._terminated:
            with self._connections_lock:
                for sock in self.unregistered_connections.values():
                    sock.close()
                
                for sock in self.registed_connections.values():
                    sock.close()
        
        for handle in self.handles:
            handle.join()
    
        logging.info('action: stop_server | result: success')

    def __results_ready(self) -> bool:
        with self._results_condition:
            return all(self._clients_finished.values())    

    def _winning_bets(self) -> list[Bet]:
        with self._bets_lock:
            if self._winning_bets_list:
                return self._winning_bets_list
            self._winning_bets_list = [bet for bet in load_bets() if has_won(bet)]
            return self._winning_bets_list

    def __handle_new_connection(self, sock, addr):
        self.unregistered_connections[addr] = sock

        message = communication.recv_message(sock)
        if not message or not message.is_connect():
            logging.info(f"action: connect | result: failure | ip: {addr[0]}")
            sock.close()
            return

        agency = message.agency()
        with self._connections_lock:
            self.registed_connections[agency] = self.unregistered_connections.pop(addr)
        logging.info(f"action: connect | result: success | ip: {addr[0]} | agency: {agency}")
        self.__handle_client_connection(sock, agency)


    def __handle_client_connection(self, sock, agency):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """

        results_sent = False
        while not self._terminated and not results_sent:
            try:
                message: communication.Message = communication.recv_message(sock)
                if not message:
                    logging.info(f"action: recv message | result: failure | agency: {agency} | error: {e}")
                    break
                results_sent = self.__process_message(message)
            except OSError as e:
                if not self._terminated:
                    logging.error(f"action: recv message | result: failure | agency: {agency} | error: {e}")
                break

        with self._connections_lock:
            self.registed_connections[agency].close()
            del self.registed_connections[agency]
        logging.info(f"action: stop thread | result: success | agency: {agency}")

    def __process_message(self, message: communication.Message):
        # Bet message
        if message.is_bet():
            logging.debug(f"action: processing_message | agency: {message.agency()} | result: in_progress | type: bet")
            bets = message.bets()
            with self._bets_lock:
                store_bets(bets)
            communication.send_confirmation(self.registed_connections[message.agency()])
            logging.info(
                f"action: batch_apuestas_almacenado | agency: {message.agency()} | result: success | cantidad: {len(bets)}"
            )
            return False

        # Finished message
        elif message.is_finished():
            logging.debug(f"action: processing_message | agency: {message.agency()} | result: in_progress | type: finished")
            with self._results_condition:
                self._clients_finished[message.agency_id] = True
            if self.__results_ready():
                with self._results_condition:
                    self._results_condition.notify_all()
                logging.info(
                    f"action: sorteo | result: success | agency: {message.agency()} | cant_ganadores: {len(self._winning_bets())}"
                ) 
            return False

        # Consult winners message
        elif message.is_consult_winners():
            logging.debug(f"action: processing_message | agency: {message.agency()} | result: in_progress | type: consult_winners")
            if not self.__results_ready():
                with self._results_condition:
                    logging.info(f"action: wait for winners | agency: {message.agency()} | result: in_progress")
                    self._results_condition.wait()
                    if self._terminated:
                        return False
    
            winning_bets = self._winning_bets()
            winning_documents_for_agency = [bet.document for bet in winning_bets if bet.agency == message.agency()]
            communication.send_winners(self.registed_connections[message.agency()], winning_documents_for_agency)
            logging.info(
                f"action: winners_sent | agency: {message.agency()} | result: success | cantidad: {len(winning_documents_for_agency)}"
            )
            return True

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
        return c, addr

    def stop(self):
        # Set the server to terminated state, so it won't keep looping
        logging.info('action: stop_server | result: started')
        self._terminated = True
        self._server_socket.shutdown(socket.SHUT_RDWR)
 
        with self._connections_lock:
            for sock in self.unregistered_connections.values():
                sock.shutdown(socket.SHUT_RDWR)
                sock.close()
            
            for sock in self.registed_connections.values():
                sock.shutdown(socket.SHUT_RDWR)
                sock.close()
            
        with self._results_condition:
            self._results_condition.notify_all()

        logging.info('action: stop_server | result: in progress')