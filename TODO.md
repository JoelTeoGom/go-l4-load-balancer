# TODO

## Crear pod de un servicio

### Control plane (lb)
- [ ] Recibir la petición "crear pod del servicio X"
- [ ] Comprobar si el servicio X existe → si no existe, devolver error
- [ ] Seleccionar un nodo aleatoriamente (de momento no hay scheduler)
- [ ] Enviar `ActionCreatePod` al nodo seleccionado
- [ ] Esperar confirmación de que el pod se ha creado correctamente
  - [ ] Si falla → devolver error y NO hacer broadcast
  - [ ] Decidir cómo el agent reporta el éxito/fallo al control plane
- [ ] Si se creó bien → broadcast de `ActionAddRemoteEndpoint` (nodo + IP + NodePort + servicio) a los demás nodos
  - [ ] Excluir al nodo que ha creado el pod (ya tiene su regla de local pod)

### Agent — nodo seleccionado (`ActionCreatePod`)
- [x] Crear el pod (ns + reglas iptables)
- [x] Añadir la regla de local pod
- [x] Recalcular posibilidades/backends
- [ ] Pulir y arreglar la creación de pods

### Agent — nodos restantes (`ActionAddRemoteEndpoint`)
- [ ] Añadir el nuevo tipo de evento en [agent/internal/event/event.go](agent/internal/event/event.go)
- [ ] Manejar el evento en el agent
- [ ] Comprobar si ya existe la regla "el nodo Y tiene al menos un pod del servicio X"
  - [ ] Si existe → no hacer nada (ya se había enterado antes)
  - [ ] Si no existe → añadir la regla del remote endpoint (IP del nodo + NodePort)
- [ ] Si se añadió una regla nueva → reasignar todos los backends / recalcular posibilidades del servicio

## Estado en el control plane
- [ ] Crear un registro/estado del clúster en el control plane:
  - [ ] Servicios existentes
  - [ ] Qué nodos tienen al menos un pod de cada servicio (IP + NodePort)
- [ ] Mantener este estado actualizado al crear pods (y en el futuro al borrarlos)

## Registro de un nodo nuevo
- [ ] Al registrarse un nodo, el control plane le envía el estado actual:
  - [ ] Todos los servicios existentes
  - [ ] Todos los remote endpoints (en qué nodos hay pods de cada servicio)
- [ ] El nodo nuevo crea las reglas correspondientes
  - Sin esto el nodo es inútil: si sale del nodo, tiene que saber dónde están disponibles los servicios en los otros nodos

## Refactor / naming
- [ ] Renombrar `lb` / "load balancer" / "control plane" a orquestador (o similar) — en realidad es un mini Kubernetes

## Networking
- [ ] Arreglar todo el tema de networking (ahora mismo está medio roto)

## Preguntas abiertas
- [ ] ¿Qué pasa si el broadcast falla en algún nodo? (reintentos, consistencia)
- [ ] ¿Qué pasa si un nodo remoto recibe `ActionAddRemoteEndpoint` de un servicio que aún no tiene creado?
- [ ] Futuro: evento inverso (`ActionRemoveRemoteEndpoint`) cuando un nodo se queda sin pods de ese servicio
- [ ] Futuro: scheduler en vez de selección aleatoria de nodo
