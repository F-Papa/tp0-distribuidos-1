from curses.ascii import SI
import logging
import socket
from .utils import Bet

SIZE_FIELD_LENGTH = 2      # Size of the length field in bytes
TYPE_FIELD_LENGTH = 1      # Size of the type field in bytes
NUMBER_LENGTH_IN_BYTES = 2 # Size of the number field in bytes
AGENCY_LENGTH_IN_BYTES = 1 # Size of the agency field in bytes
DNI_LENGTH_IN_BYTES = 4    # Size of the DNI field in bytes
YEAR_LENGTH_IN_BYTES = 2   # Size of the year field in bytes
MONTH_LENGTH_IN_BYTES = 1  # Size of the month field in bytes
DAY_LENGTH_IN_BYTES = 1    # Size of the day field in bytes

# Client Codes
CONNECT_CODE = 10          # The code the client uses to connect to the server
BET_MSG_CODE = 14          # The code the client uses to send a bet
FINISHED_CODE = 20         # The code the client uses to end betting
CONSULT_CODE = 23          # The code the client uses to request the results

# Server Codes
CONFIRMATION_CODE = 21     # The code the server uses to confirm a bet batch
RESULTS_MSG_CODE = 22      # The code the server uses to send the results
WAIT_MSG_CODE = 25          # The code the server uses to tell the client to wait


class Message():
    def __init__(self):
        self.agency_id = None
        raise Exception("Message is an abstract class")

    def agency(self):
        return self.agency_id

    def is_bet(self):
        return False

    def is_finished(self):
        return False

    def is_consult_winners(self):
        return False

    def is_connect(self):
        return False

class ConsultWinnersMessage(Message):
    def __init__(self, agency: int):
        self.agency_id = agency
    
    def is_consult_winners(self):
        return True

class BetMessage(Message):
    def __init__(self, agency: int, bets: list[Bet]):   
        self.agency_id = agency
        self.bet_list = bets

    def is_bet(self):
        return True
    
    def bets(self):
        return self.bet_list
     
class FinishedMessage(Message):
    def __init__(self, agency: int):
        self.agency_id = agency

    def is_finished(self):
        return True

class ConnectMessage(Message):
    def __init__(self, agency: int):
        self.agency_id = agency

    def is_connect(self):
        return True


def recv_message(sock: socket.socket) -> Message:
    """
    Receive a message through a socket
    """

    msg = sock.recv(SIZE_FIELD_LENGTH)
    if not msg:
        logging.error(f"action: receive_message | result: fail | error: Empty message received")
        return None

    def expected_length(msg: bytes) -> int:
        if len(msg) < SIZE_FIELD_LENGTH:
            return SIZE_FIELD_LENGTH
        
        return int.from_bytes(msg[:SIZE_FIELD_LENGTH], byteorder='big')
    
    while len(msg) < expected_length(msg):
        msg += sock.recv(expected_length(msg) - len(msg))
    
    message_type = int.from_bytes(msg[SIZE_FIELD_LENGTH:SIZE_FIELD_LENGTH+TYPE_FIELD_LENGTH], byteorder='big')
    agency_id = int.from_bytes(msg[SIZE_FIELD_LENGTH+TYPE_FIELD_LENGTH:SIZE_FIELD_LENGTH+TYPE_FIELD_LENGTH+AGENCY_LENGTH_IN_BYTES], byteorder='big')

    if message_type == BET_MSG_CODE:
        bets = _bets_from_bytes(msg[SIZE_FIELD_LENGTH+TYPE_FIELD_LENGTH+AGENCY_LENGTH_IN_BYTES:], agency_id)
        return BetMessage(agency_id, bets)
    elif message_type == FINISHED_CODE:
        return FinishedMessage(agency_id)
    elif message_type == CONSULT_CODE:
        return ConsultWinnersMessage(agency_id)
    else:
        logging.error(f"action: receive_message | result: fail | error: Unknown message received | message: {msg.hex()}")
        logging.error(f"Length: {expected_length(msg)}")
        return None

def _bets_from_bytes(data: bytes, agency: int) -> list[Bet]:
    # packet_size = int.from_bytes(data[:1], byteorder='big')
    offset = 0
    bets = []

    while offset < len(data):
        i = offset
        chosen_number = int.from_bytes(data[i:i+NUMBER_LENGTH_IN_BYTES], byteorder='big')
        i += NUMBER_LENGTH_IN_BYTES
        document = int.from_bytes(data[i:i+DNI_LENGTH_IN_BYTES], byteorder='big')
        i += DNI_LENGTH_IN_BYTES
        birth_day = int.from_bytes(data[i:i+DAY_LENGTH_IN_BYTES], byteorder='big')
        i += DAY_LENGTH_IN_BYTES
        birth_month = int.from_bytes(data[i:i+MONTH_LENGTH_IN_BYTES], byteorder='big')
        i += MONTH_LENGTH_IN_BYTES
        birth_year = int.from_bytes(data[i:i+YEAR_LENGTH_IN_BYTES], byteorder='big')
        i += YEAR_LENGTH_IN_BYTES    

        # Parse name and lastname and increment offset
        first_delim = data.find(b'|', i)
        second_delim = data.find(b'|', first_delim+1)
        
        if first_delim == -1 or second_delim == -1:
            break
        
        decoded_string = data[i:second_delim].decode()
        name, lastname = decoded_string.split('|')
        offset = second_delim + 1

        # Format birthdate
        if birth_day < 10:
            birth_day = f"0{birth_day}"
        if birth_month < 10:
            birth_month = f"0{birth_month}"
        birthdate = f"{birth_year}-{birth_month}-{birth_day}"

        new_bet = Bet(agency=agency, first_name=name, last_name=lastname, document=document, birthdate=birthdate, number=chosen_number)
        bets.append(new_bet)
    
    return bets

def recv_bet_batch(sock: socket.socket) ->  list[Bet]:
    """
    Receive serialized bets through a socket
    """
    msg = sock.recv(1024)
    if not msg:
        return None

    def expected_length(msg: bytes) -> int:
        if len(msg) < SIZE_FIELD_LENGTH:
            return SIZE_FIELD_LENGTH
        return int.from_bytes(msg[:SIZE_FIELD_LENGTH], byteorder='big')
    
    while len(msg) < expected_length(msg):
        msg += sock.recv(expected_length(msg) - len(msg))

    return _bets_from_bytes(msg.rstrip())

def send_winners(sock: socket.socket, winners_documents: list[str]) -> None:
    """
    Send the winners through a socket
    """

    encoded_winners_list = list(map(lambda x: int(x).to_bytes(DNI_LENGTH_IN_BYTES, byteorder='big'), winners_documents))
    encoded_winners = b''.join(encoded_winners_list)
    winners_message = RESULTS_MSG_CODE.to_bytes(1, byteorder='big') + encoded_winners
    _send_aux(sock, winners_message)

def send_confirmation(sock: socket.socket) -> None:
    """
    Send a confirmation message through a socket
    """
    confirmation_message = CONFIRMATION_CODE.to_bytes(1, byteorder='big')
    _send_aux(sock, confirmation_message)
    
def send_wait(sock: socket.socket) -> None:
    """
    Send a wait message through a socket
    """
    wait_message = WAIT_MSG_CODE.to_bytes(1, byteorder='big')
    _send_aux(sock, wait_message)

def _send_aux(sock: socket.socket, message: bytes) -> None:
    """
    Send a message through a socket guaranteeing that all the bytes are sent
    """
    total_bytes_sent = 0
    message += b'\n'
    bytes_to_send = len(message) + SIZE_FIELD_LENGTH
    packet = bytes_to_send.to_bytes(SIZE_FIELD_LENGTH, byteorder='big') + message

    while total_bytes_sent < len(packet):
        bytes_sent = sock.send(packet[total_bytes_sent:])
        total_bytes_sent += bytes_sent
    
    logging.debug(f"Sent {total_bytes_sent} bytes: {packet.hex()}")