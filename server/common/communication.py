import logging
import socket
from .utils import Bet

SIZE_FIELD_LENGTH = 2      # Size of the length field in bytes
NUMBER_LENGTH_IN_BYTES = 2 # Size of the number field in bytes
AGENCY_LENGTH_IN_BYTES = 1 # Size of the agency field in bytes
DNI_LENGTH_IN_BYTES = 4    # Size of the DNI field in bytes
YEAR_LENGTH_IN_BYTES = 2   # Size of the year field in bytes
MONTH_LENGTH_IN_BYTES = 1  # Size of the month field in bytes
DAY_LENGTH_IN_BYTES = 1    # Size of the day field in bytes
CONFIRMATION_CODE = 21     # What the server sends to confirm a packet

def _bets_from_bytes(data: bytes) -> list[Bet]:
    # packet_size = int.from_bytes(data[:1], byteorder='big')
    
    agency = int.from_bytes(data[SIZE_FIELD_LENGTH:SIZE_FIELD_LENGTH+AGENCY_LENGTH_IN_BYTES], byteorder='big')

    offset = 3
    bets = []

    #Print data as hex
    # logging.debug(f"Data Received: {data.hex()}")

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
            logging.info(f"Data: {data.hex()}")
            logging.info(f"offset: {offset} | first_delim: {first_delim} | second_delim: {second_delim} | len(data): {len(data)}")
            break
        
        decoded_string = data[offset+10:second_delim].decode()
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

    # logging.debug(f"Received {len(msg)} bytes: {msg.hex()}")

    return _bets_from_bytes(msg.rstrip())

def send_confirmation(sock: socket.socket) -> None:
    """
    Send a message through a socket
    """
    confirmation_message = CONFIRMATION_CODE.to_bytes(1, byteorder='big') + b"\n"
    total_bytes_sent = 0
    while total_bytes_sent < len(confirmation_message):
        bytes_sent = sock.send(confirmation_message[total_bytes_sent:])
        total_bytes_sent += bytes_sent
    