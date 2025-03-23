import sys

SERVICES_TEXT="""name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    volumes:
      - type: bind
        source: ./server/config.ini
        target: /config.ini
    environment:
      - PYTHONUNBUFFERED=1
    networks:
      - testing_net
"""

CLIENT_TEMPLATE="""  client{client_id}:
    container_name: client{client_id}
    image: client:latest
    entrypoint: /client
    volumes:
      - type: bind
        source: ./client/config.yaml
        target: /config.yaml
    environment:
      - CLI_ID={client_id}
      - NACIMIENTO=1999-03-17
    networks:
      - testing_net
    depends_on:
      - server
"""

NETWORKS_TEXT="""networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""

def main():
    filename, number_of_clients = sys.argv[1], sys.argv[2]

    clients_text = ""
    for client_id in range(1, int(number_of_clients) + 1):
        template_completed = CLIENT_TEMPLATE.format(client_id=client_id)
        clients_text += template_completed + '\n'

    with open(filename, "w") as file:
        final_text = SERVICES_TEXT + '\n' + clients_text + NETWORKS_TEXT
        file.write(final_text)

if __name__ == '__main__':
    main()
