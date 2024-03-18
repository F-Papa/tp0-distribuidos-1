import logging
import socket
from .utils import Bet


def _bet_from_bytes(data: bytes):
    packet_size = int.from_bytes(data[:1], byteorder='big')
    agency = int.from_bytes(data[1:2], byteorder='big')
    chosen_number = int.from_bytes(data[2:4], byteorder='big')
    decoded_string = data[4:].decode('utf-8')

    bettor_info = decoded_string.split('|')
    name = bettor_info[0]
    lastname = bettor_info[1]
    document = bettor_info[2]
    birthdate = bettor_info[3]
    
    return Bet(agency=agency, first_name=name, last_name=lastname, document=document, birthdate=birthdate, number=chosen_number)

def recv_bet(sock: socket.socket) ->  Bet:
    """
    Receive a serialized bet through a socket
    """
    msg = sock.recv(1024)
    if not msg:
        return None

    expected_length = int.from_bytes(msg[:1], byteorder='big')
    
    while len(msg) < expected_length:
        msg += sock.recv(1024)

    return _bet_from_bytes(msg.rstrip())

def send_confirmation(sock: socket.socket) -> None:
    """
    Send a message through a socket
    """
    CONFIRMATION_MESSAGE = "1\n"

    bytes_sent = sock.send(CONFIRMATION_MESSAGE.encode('utf-8'))
    total_bytes_sent = bytes_sent
    while total_bytes_sent < len(CONFIRMATION_MESSAGE):
        bytes_sent = sock.send(CONFIRMATION_MESSAGE[total_bytes_sent:].encode('utf-8'))
        total_bytes_sent += bytes_sent
    