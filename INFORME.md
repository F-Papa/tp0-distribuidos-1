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
La clase `server` ahora tiene un atributo privado (simbolizado por el prefijo `__`) `__terminated` que se usa como condición para el loop de su método `run`. Esto permite que mediante un llamado al método `stop`, el servidor deje de loopear.

El handler de `SIGTERM` recibe dos parámetros que ya están definidos (signum y frame), así que para se pudiera llamar a `stop` el server pasó a ser una variable global. Otra solución podría haber sido un booleano o un pipe, pero no parecía posible escapar de tener una variable global.

### Cliente
Ahora la clase `client` tiene un atributo `terminated`, cuando este se setea en true, mediante el nuevo método `Terminate`, el loop del cliente no continúa y la función `StartClientLoop` retorna.

Para manejar la señal del SO, tuve que crear un `channel` que escuchara las señales. Este channel y un puntero al puntero del cliente son pasados como parámetros al handler, que al recibir una señal `SIGTERM`, llama a `client.Terminate` en caso de que `client` ya existiera.
