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
- [ ] El nombre del repo (`go-l4-load-balancer`) ya no dice lo que hay dentro
- [x] README actualizado: el datapath de iptables, y la sección de alcance y limitaciones
      (Linux only, VMs en vez de contenedores, CRI/CNI hardcodeados, sin estado declarativo)

## Agent — bloqueantes

- [ ] **El módulo no compila**: `reconcile.go:49` llama a `CreateService` con 3 args y pide 5.
      El payload llega como `name+ip+port`, así que hay que decidir si `nodePort` y `podPort`
      los manda el control plane o van por defecto en el agente.
- [ ] **`KUBE-FORWARD` está vacía**: se salta a ella desde `FORWARD` pero no tiene reglas, así
      que todo depende de la policy del kernel. Con `FORWARD DROP` (Docker la pone así) no
      funciona ni pod↔pod, ni pod→fuera, ni NodePort→pod. Faltan los ACCEPT de
      `ESTABLISHED,RELATED`, `-s PodCIDR` y `-d PodCIDR`, con un `-F` delante.

## Agent — no corre nada dentro del pod

El envoltorio de red está completo (netns + veth + IP + ruta), pero ningún proceso entra
nunca en ese namespace. Consecuencias: `Status` se queda en `PENDING` para siempre, y el
DNAT apunta a un puerto donde no escucha nadie.

- [ ] Añadir a `Pod` qué ejecutar y el proceso resultante (`Command []string`, `*exec.Cmd`)
- [ ] Lanzar el binario con `ip netns exec <netns> <cmd>` al final de `CreatePod`
- [ ] **`run()` no vale para esto**: usa `CombinedOutput()`, que bloquea hasta que el comando
      termina. Hace falta otra función con `cmd.Start()` que guarde el proceso
- [ ] Pasar `Status` a `RUNNING` cuando el proceso arranca, y a `FAILED` si muere
- [ ] `RemovePod`: el SIGTERM/SIGKILL necesita ese proceso guardado

## Agent — rollback y limpieza

- [ ] Rollback de `CreatePod` (ver `TODO(rollback)` en `agent/internal/agent/pod.go`): la IP y
      el ID ya se liberan, falta deshacer el netns y el veth, y restaurar la cadena del service
      si el fallo llega después del `-F`
- [ ] `RemovePod` completo (ver `TODO(removepod)`): `ReleaseIP`, `ReleaseID`, borrar netns y
      veth, tirar la cadena `KUBE-SEP-*` y recalcular probabilidades
- [ ] **Trampa**: al reutilizar un índice liberado, el `-N pod.ID` peta con "Chain already
      exists" si no se borró antes la cadena SEP vieja
- [ ] **Al borrar el último pod** hay que volver a poner la regla `REJECT` en la cadena del
      service: el `-F` de `SetupPodIptables` se la lleva, y una cadena vacía deja pasar el
      tráfico en vez de rechazarlo

## Agent — idempotencia y calidad

- [ ] Los `-A` de `CreateServiceSetup` se duplican si el agente reinicia y recrea un service.
      El `-N` ya lo cubre `runIgnoreExists`, los `-A` no (mismo problema que los jumps, misma
      solución: `-F` o comprobar con `-C`)
- [ ] Cadenas `KUBE-SEP-*` para backends remotos: hoy el salto va a `backend.ID`, que para un
      remoto es un ID de nodo y no una cadena. Necesitan su propia cadena con DNAT a
      `nodeIP:NodePort`
- [ ] `func (s *Service) Chain() string` — el nombre de la cadena se construye a mano en dos
      sitios
- [ ] `CheckPodHealth` sigue siendo un stub
- [ ] Comentario obsoleto en `node.go`: "len(AllocatedIPs) is used as pointer to the next free IP"

## Limitaciones conocidas de networking

- [ ] NodePort solo responde en la IP del nodo (`-d nodeIP/32`), no en `127.0.0.1` ni en otra
      NIC. Lo correcto sería `-m addrtype --dst-type LOCAL`
- [ ] Hairpin sin resolver: un pod que llama al ClusterIP de su propio service y se toca a sí
      mismo en el sorteo. kube-proxy lo arregla marcando esos paquetes para MASQUERADE

## Control plane — hacia un API server de verdad

Ahora mismo el control plane despacha eventos imperativos y el reconciliador los ejecuta en
orden. Eso es un orquestador imperativo. Lo que falta para que sea declarativo:

- [ ] Un almacén key-value (embebido o en memoria, tipo etcd de juguete) donde persistir los
      objetos del clúster: nodos, servicios, pods, endpoints
- [ ] Modelar esos objetos con estado **deseado** y estado **observado**, no solo eventos
- [ ] Un segundo bucle de reconciliación, separado del de eventos: un ticker que cada N
      minutos compare lo que el control plane cree que hay con lo que el nodo reporta, y
      emita las acciones que hagan falta
- [ ] Decidir qué manda cuando los dos bucles discrepan
- [ ] Sin esto, un nodo que se pierde eventos (desconexión, reinicio) se queda desincronizado
      para siempre

## Data plane

- [ ] Queda por programar buena parte del data plane del `lb/`

## Networking general

- [ ] Repasar el tema de networking de punta a punta una vez el build compile y haya un proceso
      dentro del pod (hasta entonces no se puede probar nada de verdad)

## Preguntas abiertas
- [ ] ¿Qué pasa si el broadcast falla en algún nodo? (reintentos, consistencia)
- [ ] ¿El resync periódico lo pide el agente al control plane, o el control plane lo empuja?
- [ ] ¿Qué pasa si un nodo remoto recibe `ActionAddRemoteEndpoint` de un servicio que aún no tiene creado?
- [ ] Futuro: evento inverso (`ActionRemoveRemoteEndpoint`) cuando un nodo se queda sin pods de ese servicio
- [ ] Futuro: scheduler en vez de selección aleatoria de nodo
