# Respuesta a preguntas del oral
1. ReadFull vs read_full custom. Ciertamente ReadFull en caso de que no lea los bytes esperados, falla y no reintenta. Se reemplazó por read_full el cual está implementado de una forma similar que write_full. Se encuentra en packet.go en todas las ramas.
2. SIGTERM en multihilo: En caso de que se reciba un SIGTERM, ahora registra los sockets abiertos en un set. Luego al finalizar exitosamente lo saca del conjunto. En caso de que se reciba un sigterm, la funcion exit_gracefully (que se ejecuta cuando se recibe dicha señal), cierra todos los sockets abiertos.
    
    Esto hace que, si un cliente estaba leyendo un mensaje, falle y por lo tanto termina de procesar el mensaje y por lo tanto, se cierra el hilo.
Puede haber problemas de concurrencia para el set, así que se le puso un lock a ese set en particular.
3. Justificación del uso de Threads pese a las restricciones del GIL:  Lo que controla el GIL es que no se ejecute bytecode de python al mismo tiempo, es decir, las instrucciones del programa que procesa el CPU. Pero para las operaciones de llamada de red y IO, el GIL se libera.

    Como el servidor, lo que principalmente hace son lecturas del socket, entonces en general el servidor va a ejecutarse de forma paralela.
En las partes donde no se ejecuta de forma paralela (calcular ganadores por ejemplo) tenemos locks así que tampoco podrían ser paralelas.
O que respecto a las llamadas de red o IO, son despreciables en cuanto a tiempo de espera.

# Documentación de los ejercicios (5 al 8)
Documento los protocolos de cada ejercicio. Es el mismo contenido que hay en cada rama pero unificado.

## Documentación Protocolo Ej 5
Para esta versión inicial se implementaron 2 estructuras: Packet y Apuesta.

* Packet: Es un wrapper de un arreglo de bytes. Contiene un header y un payload. El header simplemente indica el tamaño en bytes del payload. El header siempre mide 4 bytes.
Entonces para la comunicación siempre se leen/escriben 4 bytes primero y luego la cantidad de bytes que indica dicho header. Esto lo hace tanto el cliente como el servidor para comunicarse.
* Apuesta: Para poder interpretar los bytes dentro del Packet, se utilizó el protocolo TLV. Se establecieron los campos que tiene una apuesta (Type: 1 byte), la longitud del valor (2 bytes) y el valor. Para ello se implementaron los metodos Serialize y Deserialize.

Ejemplo de comunicación:
1. El cliente primero serializa la apuesta (TLV).
2. Una vez serializado, lo envuelve en un Packet. Ahora tiene una cabecera con el largo de toda la Apuesta serializada.
3. Envía el paquete por medio del socket al servidor
4. El servidor lee 4 bytes, y con ello puede identificar el largo del payload.
5. Deserializa el payload interpretandolo como una Apuesta, en base al TLV.
6. Responde con un Packet con mensaje "OK".

## Documentación Protocolo Ej 6
Para esta versión se implementaron 3 estructuras: Packet, Apuesta y Batch.

* Packet: Es un wrapper de un arreglo de bytes. Contiene un header y un payload. El header simplemente indica el tamaño en bytes del payload. El header siempre mide 4 bytes.
Entonces para la comunicación siempre se leen/escriben 4 bytes primero y luego la cantidad de bytes que indica dicho header. Esto lo hace tanto el cliente como el servidor para comunicarse.
* Apuesta: Para poder interpretar los bytes dentro del Packet, se utilizó el protocolo TLV. Se establecieron los campos que tiene una apuesta (Type: 1 byte), la longitud del valor (2 bytes) y el valor. Para ello se implementaron los metodos Serialize y Deserialize.
* Batch: Un Batch contiene el AgencyID y un payload. En el payload van los Packet que contienen las Apuesta serializadas.

Ejemplo de comunicación:
1. El cliente primero lee Apuestas del CSV y las serializa (TLV).
2. Una vez serializado, lo envuelve en un Packet. Ahora tiene una cabecera con el largo de toda la Apuesta serializada.
3. Acumula en un arreglo, los Packet serializados a enviar en forma de bytes.
4. A dicho arreglo lo junta con 4 bytes de AgencyID. Y a todo eso lo envuelve en un Packet y se envía al servidor.
5. El servidor lee el Packet, y con ello puede identificar el largo del payload.
6. Deserializa el payload interpretandolo como un Batch (AgencyID + payload), por lo que lee 4 bytes para el AgencyID, y luego al payload empieza a Deserializarlo como un array de Packets.
7. A cada Packet lo deserializa como una Apuesta (TLV), y al finalizar de procesar el batch responde con un Packet("OK")
8. El cliente lee el OK y continua.


## Documentación Protocolo Ej 7
Para esta versión se implementaron las siguientes estructuras: Packet, Apuesta, Batch y Message.

* Packet: Es un wrapper de un arreglo de bytes. Contiene un header y un payload. El header simplemente indica el tamaño en bytes del payload. El header siempre mide 4 bytes.
Entonces para la comunicación siempre se leen/escriben 4 bytes primero y luego la cantidad de bytes que indica dicho header. Esto lo hace tanto el cliente como el servidor para comunicarse.
* Apuesta: Para poder interpretar los bytes dentro del Packet, se utilizó el protocolo TLV. Se establecieron los campos que tiene una apuesta (Type: 1 byte), la longitud del valor (2 bytes) y el valor. Para ello se implementaron los metodos Serialize y Deserialize.
* Batch: Un Batch contiene el AgencyID y un payload. En el payload van los Packet que contienen las Apuesta serializadas.
* Message: Está compuesto por un Byte indicando el tipo de mensaje (Batch, Consulta, RespuestaWin, RespuestaWait) y su Payload. Al serializarse, es 1 byte para el tipo de mensaje, 4 para el largo del Payload y el Payload.

Ejemplo de comunicación:
1. El cliente primero lee Apuestas del CSV y las serializa (TLV).
2. Una vez serializado, lo envuelve en un Packet. Ahora tiene una cabecera con el largo de toda la Apuesta serializada.
3. Acumula en un arreglo, los Packet serializados a enviar en forma de bytes.
4. A dicho arreglo lo junta con 4 bytes de AgencyID. Y a todo eso lo envuelve en un Message, del tipo MsgTypeBatchBet.
5. El servidor lee el Message, el cual es de tipo MsgTypeBatchBet, y con ello puede identificar el tipo y el largo del payload.
6. Deserializa el payload interpretandolo como un Batch (AgencyID + payload), por lo que lee 4 bytes para el AgencyID, y luego al payload empieza a Deserializarlo como un array de Packets.
7. A cada Packet lo deserializa como una Apuesta (TLV), y al finalizar de procesar el batch responde con un Packet("OK")
8. El cliente lee el OK y continua hasta finalizar todos las apuestas. Una vez finalizado, empieza con el envío de consultas del ganador. Arma un Message del tipo MsgTypeConsulta con el AgencyID como payload (4 bytes). Y lo envía al servidor
9. El servidor, recibe esas consultas. En caso de que sea la primera vez que se recibe esa consulta, se agenda como que dicha agencia terminó de procesar las apuestas. Responde con un Message del tipo MsgTypeRespuestaWait en caso de que no se haya alcanzado el threshold de agencias finalizadas, o MsgTypeRespuestaWinner en caso de que ya se hayan sorteado los ganadores, con el payload como un array de documentos (4 bytes) de los ganadores de la agencia consultada.
10. El cliente lee ese payload, imprime la cantidad de ganadores y finaliza el loop.


## EJ8: Soporte Multi-threading y paralelismo
En el contexto Multi-Threading, se deben identificar los recursos compartidos para poder garantizar que se están utilizando de forma exclusiva.
En el servidor, son los siguientes:
- el archivo que utiliza como persistencia `store_bets` y `load_bets`
- _agencies_that_finished
- _sorteo_done

Se utiliza un Lock, para garantizar que:
- Al recibir las apuestas y se intenten guardar en el archivo con la función `store_bets`, entonces solo 1 thread podrá hacerlo a la vez.
- Al finalizar el sorteo, solo 1 thread podrá cargar las apuestas en el diccionario `results` con la funcion `load_bets`.
- Cuando se llega al threshold de agencias necesarias para finalizar el sorteo, entonces se modifican `_agencies_that_finished` y `_sorteo_done` (hay otros threads leyendo)

En cuanto al Paralelismo, se utiizó un ThreadPool para ejecutar la función `__handle_client_connection`, el cual se encarga de crear un thread para cada nuevo cliente y despacharlo para que un hilo lo procese. 

- Cada hilo recibe el socket abierto y lo cierra cuando termina.
- Procesa los mensajes que se reciben de ese socket.

Notar que se está procesando los mensajes en paralelo, como lo pedía la consigna pues cada thread tiene su socket donde lee y escribe los mensajes, no necesita un lock para utilizarlo. Sin embargo no hay paralelismo en otras acciones durante el procesamiento, es allí donde se usan los locks para garantizar la concurrencia.
