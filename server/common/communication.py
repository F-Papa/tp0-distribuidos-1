import socket
from .utils import Bet


CONFIRMATION_MESSAGE = 21

def _bet_from_bytes(data: bytes):
    # packet_size = int.from_bytes(data[:1], byteorder='big')
    agency = int.from_bytes(data[1:2], byteorder='big')
    chosen_number = int.from_bytes(data[2:4], byteorder='big')
    document = int.from_bytes(data[4:8], byteorder='big')
    birth_day = int.from_bytes(data[8:9], byteorder='big')
    birth_month = int.from_bytes(data[9:10], byteorder='big')
    birth_year = int.from_bytes(data[10:12], byteorder='big')
    decoded_string = data[12:].decode('utf-8')

    if birth_day < 10:
        birth_day = f"0{birth_day}"
    if birth_month < 10:
        birth_month = f"0{birth_month}"
    birthdate = f"{birth_year}-{birth_month}-{birth_day}"

    name, lastname = decoded_string.split('|')
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

def send_confirmation(bet: Bet, sock: socket.socket) -> None:
    """
    Send a message through a socket
    """

    confirmation_message = CONFIRMATION_MESSAGE.to_bytes(1, byteorder='big')
    confirmation_message += bet.document.to_bytes(4, byteorder='big')
    confirmation_message += bet.number.to_bytes(2, byteorder='big')
    confirmation_message += b"\n"

    total_bytes_sent = 0
    while total_bytes_sent < len(confirmation_message):
        bytes_sent = sock.send(confirmation_message[total_bytes_sent:])
        total_bytes_sent += bytes_sent
    