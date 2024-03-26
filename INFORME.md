# Informe TP0 - Franco Papa (106249)

## Ejercicio 1
Para este ejercicio la solucion fue modificar el archivo `docker-compose-dev.yaml`. En él ruve que duplicar las lineas correspondientes al servicio **'client1'** y reemplazar los **1** por **2** en la versión duplicada.

## Ejercicio 1.1
Para realizar este ejercicio tuve que dividir el archivo de **docker compose** en secciones. A cada una le dediqué una funcion que retornase el texto correspondiente.
De todas ellas la única que recibe parámetros es la del servicio **client_n** que automatiza lo hecho en el ejercicio anterior. 

Finalmente escribo el texto generado por mis funciones en el archivo `docker-compose-dev.yaml`

Para ejecutarlo basta con el comando:
> python compose.py \<num_de_clientes>

## Ejercicio 2
En este ejercicio aprendí que no era apropiado usar tabs para identar las líneas en el archivo de docker compose, por lo que tuve que recurrir al doble espacio. 

Para lograr lo pedido me pareció que la mejor opción era un `host volume` que simplemente montase el archivo del host original dentro del contenedor. Para ello modifiqué el archivo `compose.py` para que agregase la clave `volumes`
con el respectivo **source:target** a cada servicio del archivo `docker-compose-dev.yaml`.

## Ejercicio 3
Para este ejercicio tuve crear un script the python que ejecutara en un **proceso hijo** la aplicacion `netcat`.

Para la conexion, es importante agregar el contenedor donde se ejecute el script a la misma red que el server, cuyo nombre se puede encontrar con  `docker network ls` y es `tp0_testing_net`. Si esto se cumple, se puede acceder al server simplemente por el hostname: `server`

Para automatizar la prueba, en proceso hijo el `stdin` y `stdout` son reemplazados por unos `os.pipe`, de modo que el proceso padre puede escribir el input y leer el output. 

El script `test_server.sh` se ocupa de crear la imagen y el contenedor para correr la prueba y mostrar el resultado. Puede ejecutarse mediante:

> ./test_server.sh

## Ejercicio 4

### Server
La clase `Server` ahora tiene dos atributos privados (simbolizado por el prefijo `_`) `_terminated` y `_conn` que se usa como utilizados para el graceful shutdown del mismo. El primero es un flag para que el loop de su método `run`, deje de loopear. El segundo es el socket obtenido mediante `_server_socket.accept`.


Una vez creado el server en `main`, el handler de `SIGTERM` es asignado a una función anónima que llama al método `stop` del server. Este método hace un `RDWR_SHUTDOWN` del los dos sockets del server y setea `_terminated = True`. Cuando termina el loop, se llama a `socket.close` para liberar los recursos.



### Cliente
Ahora la clase `Client` tiene un atributo `terminated`, cuando este se setea en true, mediante el nuevo método `Terminate`, el loop del cliente no continúa y la función `StartClientLoop` retorna. 

Otra cambio es que `client.createClientSocket` chequea el valor de `terminated` antes de retornar. Esto es porque al salir del *signal handler*, el *instruction pointer* podría encontrarse dentro del loop y realizar una iteración extra, cosa que no es deseado.

Dentro del loop cuando se chequea si hubieron errores en las operaciones de red, no se loggean errores si `terminated == True`, ya que es esperable que ocurran dado que se está intentando usar un socket cerrado. 

Para manejar la señal del SO, tuve que crear un `channel` que escuchara las señales. Este channel y un puntero al puntero del cliente son pasados como parámetros al handler, que al recibir una señal `SIGTERM`, llama a `client.Terminate` en caso de que `client` ya existiera.

## Ejercicio 5

### Serialización
#### Apuestas
Para la implementación de los nuevos requisitos del cliente, tuve que crear una clase `Bet` y una clase `BettorInfo` (información de quien apuesta). Junto con ellas pensé en una forma de serializar las apuestas y resultó así:

    <tamaño><agencia><número><dni><dia><mes><año><nombre>|<apellido>

Donde los primeros 7 campos no necesitan delimitador porque tienen tamaño fijo:
- `tamaño`: 1 Byte (Serializaciones de a lo sumo 255 bytes)
- `id agencia`: 1 Byte (A lo sumo 255 agencias, y solo hay 5)
- `número`: 2 Bytes (Numeros del 0 al 65535)
- `dni`: 4 Bytes Números hasta 100M+
- `día`: 1 Bytes (Día de nacimiento)
- `mes`: 1 Bytes (Mes de nacimiento)
- `año`: 2 Bytes (Año de nacimiento)

Y los 2 campos siguientes siguientes están delimitados por el caracter `|`, ya que su tamaño es variable.

#### Confirmación
Los mensajes de confirmacion tienen el siguiente formato y son siempre de 6 Bytes:

    <codigo msj><dni><número>
- `codigo msj`: 1 Bytes (El número 21, que en mi protocolo corresponde a confirmación, dándole al receptor la certeza de que tendrá 6 bytes de longitud).
- `dni`: 4 Bytes Números hasta 100M+
- `número`: 2 Bytes (Numeros del 0 al 65535)

#### Short-Read y Short-Write

El short-write es solucionado mediante un ciclo que sigue enviando los bytes que no se enviaron en caso de que los bytes escritos sean menores a la longitude de la serialización en bytes.

Para el short-read de la confirmación, se hace lo mismo, solo que con lectura en vez de de escritura. Esto es posible porque se sabe que el mensaje de confirmación va a ser de 6 bytes.

### Servidor
#### Recepción de Apuestas
Para el lado del servidor, implementé un módulo de `communication` que mediante el método `recv_bet` es capaz de recibir una apuesta por medio de un socket, protegiéndose del **short-read** utilizando  el primer Byte de la serializacion (que indica su longitud) y comparándolo con los Bytes leídos hasta ese momento. Además, devuelve una instancia de la clase `Bet` provista por la cátedra. 

#### Confirmación
El módulo `communication` también provee una función `send_confirmation` que envía a través del socket la confirmación que fue descripta anteriormente para una instancia de `Bet`


### Cliente
#### Módulo Communication
En el lado del cliente, el módulo de comunicaciones provee las funciones `SendBet` y `RecieveConfirmation`. La Primera se encarga de serializar una apuesta (si no supera los 255 bytes) y enviarla por un socket, asegurándose de no hacer un *short-write*. La segunda recibe un mensaje por el socket y si es un mensaje de confirmación válido, devuelve los campos `número` y `dni` del mismo.

## Ejercicio 6
Para este ejercicio el módulo `communication` de ambas partes se vio modificado para trabajar con arreglos (o listas) de `Bet` en vez de una sola. 

### Serialización
#### Paquete
    <tamaño batch><id agencia><apuestas>

Donde `tamaño` ocupa 2 bytes y es el tamaño del paquete en bytes. `id agencia` ocupa 1 byte, y `apuestas` es la informacion de las apuestas serializadas una tras otra, cada una con el formato:

#### Apuesta
    <número><dni><dia><mes><año><nombre>|<apellido>|

#### Configurabilidad
El tamaño de paquetes por lote puede especificarse dentro del archivo `config.yaml` bajo `protocol.bets_per_batch`. En caso de no estarlo, se toma el valor por defecto `250`, que mantiene los paquetes de un tamaño  menor al pedido. Se toma como hipótesis que el programa se usará con archivos donde los nombres no sean demasiado más largos que los que hay en `dataset.zip`

### Confirmación
Este aspecto se simplificó y el servidor solo responde el mensaje de confirmación de 1 byte `21 (base 10)`.

### Manejo de Archivos
#### Ubicación
Los clientes esperan encontrar los archivos en la ubicación especificada por la *variable de entorno* `BETS_FILE`. Para su prueba rápida, `docker-compose-dev.yaml` está configurado para setearla a "data.csv" y a bindear esa ubicación para *i-ésimo* cliente a la ruta relative del host `./data/agency-i.csv`


#### Lectura
La clase `CSVFile` se construye con el path a un archivo. Mediante su método `ReadBetsFromCSVFile` permite leer n lineas del archivo y devolver m cantidad de apuestas. (m < n sii se llega al final del archivo o hay algún error).

## Ejercicio 7

### Nuevos Mensajes
Se rediseñó el protocolo de comunicación para posibilitar el envío de distintos tipos de mensajes. Ahora cada los mensajes entre partes tienen el siguiente formato:
    
    <tamaño><código><agencia><cuerpo>

Donde los campos `tamaño`, `código`, y `agencia` ocupan 2 bytes, 1 byte y 1 byte respectivamente.

#### Bet Message (Código 14)
Se utiliza para enviar apuestas del cliente al servidor. Se mantiene la serialización de lotes de apuestas del ejercicio anterior una tras otra. Resultado el cuerpo del mensaje:

    <número><dni><dia><mes><año><nombre>|<apellido>|...    

#### Finished Message (Código 20)
Se utiliza para enviar la finalización del envío de apuestas por parte de una agencia. No tiene cuerpo.

#### Consult Message (Código 23)
El cliente lo envía al servidor para consultar los resultados del sorteo. No tiene cuerpo.

#### Results Message (Código 22)
El servidor lo envía al cliente cuando recibe un mensaje `Consult (23)` y ya recibió un mensaje `Finished (20)` de los 5 clientes.

#### Wait Message (Código 25)
El servidor lo envía al cliente para cuando recibe un mensaje `Consult (23)` y aún no recibió `Finished (20)` de los 5 clientes.

### Cliente
#### Fases o Etapas
Para la lograr una implementación de la lógica del cliente, este ahora tiene en su estado la `etapa` en la que se encuentra. Este estado dicta su comportamiento iteración a iteración dentro del método `StartClientLoop`.

##### Fase de Envío de Apuestas
Durante esta etapa. El cliente lee de su almacenamiento un conjunto de apuestas y las envía. Si se llega al final del archivo, entonces envía el mensaje `Finished (20)` y avanza a la etapa de Consulta de Resultados.

##### Fase de Consulta de Resultados
El cliente envía el mensaje `Consult (23)` al servidor. Si recibe una respuesta de tipo  `Wait(25)`, se suspende por `loop.interval`. Si la respuesta es en cambio de tipo `Results (22)`, el cliente guarda los ganadores en su estado y avanza a la siguiente etapa.

##### Fase de Anuncio de Ganadores
En esta etapa el cliente rompe el loop, anuncia los ganadores y continúa con su flujo hasta terminar su ejecución.

### Servidor
#### Polimorfismo de Mensaje
La clase abstracta `Message` facilita el procesamiento de mensajes por parte del servidor, hay una clase hija para cada mensaje que pueda recibir por parte del cliente:
- FinishedMessage
- BetMessage
- ConsultWinnersMessage

Cada uno de ellos tiene los atributos necesarios para que el servidor pueda ejecutar la acción correspondiente.

#### Cache de Ganadores
El servidor tiene el método `_winning_bets` que si es su primera vez ejecutándose lee los ganadores del disco y los guarda en su estado interno para futuros llamados. Si no es el primer llamado, devuelve los ganadores cacheados anteriormente.


