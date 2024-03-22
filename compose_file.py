import sys

def main(num_of_clients: int):
    print(f"Creating docker-compose-dev.yaml with {num_of_clients} clients")
    with open("docker-compose-dev.yaml", "w") as f:
        f.write(general_config())
        f.write("services:\n")
        f.write(server_config())
        for i in range(1, num_of_clients+1):
            f.write(client_config(i))
        f.write(network_config())

def general_config():
    return """
version: '3.9'
name: 'tp0'
"""

def server_config():
    return """
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net
    volumes:
      - ./server/config.ini:/config.ini
"""

def client_config(client_number: int):
    return f"""
  client{client_number}:
    container_name: client{client_number}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID={client_number}
      - CLI_LOG_LEVEL=DEBUG
      - BETS_FILE=/data.csv
    networks:
      - testing_net
    depends_on:
      - server
    volumes:
      - ./client/config.yaml:/config.yaml
      - .data/agency-{client_number}.csv:/data.csv
"""

def network_config():
    return """
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: compose_file.py <num_of_clients>")
        sys.exit(1)
    if not sys.argv[1].isdigit():
        print("num_of_clients must be an integer")
        sys.exit(1)
    
    num_of_clients = int(sys.argv[1])
    main(num_of_clients)


