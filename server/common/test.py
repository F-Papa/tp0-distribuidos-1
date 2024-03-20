# hex_str = "21011d9601d790910101000153616e746961676f204c696f6e656c7c4c6f726361"
# hex_str = "21011d9601d790910101000153616e746961676f204c696f6e656c7c4c6f726361"
# data = bytes.fromhex(hex_str)

# for a_byte in data:
#     #print byte in hex
#     print(hex(a_byte), end=' ')
# # exit()

# agency = int.from_bytes(data[1:2], byteorder='big')
# chosen_number = int.from_bytes(data[2:4], byteorder='big')
# document = int.from_bytes(data[4:8], byteorder='big')
# birth_day = int.from_bytes(data[8:9], byteorder='big')
# birth_month = int.from_bytes(data[9:10], byteorder='big')
# birth_year = int.from_bytes(data[10:12], byteorder='big')
# decoded_string = data[12:].decode('utf-8')

# bettor_info = decoded_string.split('|')
# name = bettor_info[0]
# lastname = bettor_info[1]

# birthdate = f"{birth_day}/{birth_month}/{birth_year}"

# #Print everything
# print(agency)
# print(chosen_number)
# print(document)
# print(birth_day)
# print(birth_month)
# print(birth_year)
# print(name)
# print(lastname)
# print(birthdate)

dni = int(43243730).to_bytes(4, byteorder='big')
number_as_bytes = int(2103).to_bytes(2, byteorder='big')
CONFIRMATION_MESSAGE = dni + number_as_bytes + b"\n"
print(len(CONFIRMATION_MESSAGE))