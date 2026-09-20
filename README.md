# orchard

> A toy Kubernetes, written by hand in Go: a control plane, node agents that build pod
> network namespaces and program iptables, and an L4 load balancer in front of it all.

**Status:** work in progress, and a learning project. Not something to run anywhere that
matters. See [Scope and honest limits](#scope-and-honest-limits).

---

## What this is

A miniature orchestrator, built by hand to understand how the real ones work. Nothing here
wraps an existing tool: the load balancer moves the bytes itself, the agent creates the
network namespaces itself, writes its own iptables rules, and the two talk over a protocol
of their own.

It started as an L4 load balancer. It grew. What the agent does today is, in Kubernetes
terms, **kubelet + kube-proxy**: pod lifecycle, IP allocation, veth pairs, a node bridge,
and a set of `KUBE-*` iptables chains that DNAT a virtual service address to a real pod.
The L4 load balancer survives as the component in front, the equivalent of a cloud
LoadBalancer sitting ahead of a cluster's NodePorts.

Two components, developed as if they were separate repositories, kept in one for
convenience:

- **`lb/`** — the control plane and the L4 data plane. Tracks which nodes exist and how
  they are doing, dispatches events to agents, and forwards client TCP connections to a
  node.
- **`agent/`** — the node agent. Registers with the control plane, then creates services
  and pods locally: namespaces, veth pairs, IP allocation and iptables rules.

---

## How a packet gets to a pod

The interesting part of the project, and the reason it exists.

```
client ──▶ lb data plane ──▶ nodeIP:NodePort
                                   │
                            PREROUTING
                                   │
                            KUBE-SERVICES          -d clusterIP --dport → service chain
                                   │               -d nodeIP   --dport → service chain
                                   ▼
                            KUBE-SVC-<service>     -m statistic --probability 1/n
                                   │
                    ┌──────────────┼──────────────┐
                    ▼              ▼              ▼
            KUBE-SEP-<svc>-0   ...-1      (remote node backend)
                    │
              DNAT podIP:podPort
                    │
                    ▼
              br0 ──▶ veth ──▶ pod netns eth0
```

Each pod lives in its own network namespace, joined to a node bridge by a veth pair, with
an address from the node's pod CIDR and a default route through the bridge. Each pod gets
an endpoint chain (`KUBE-SEP-*`) whose only job is the DNAT to that pod. The service chain
(`KUBE-SVC-*`) holds the load balancing itself: one `-m statistic` rule per backend with
probability `1/n`, `1/(n-1)`, …, and the last one unconditional.

Which means the actual packet-by-packet balancing inside a node is done by **the kernel**.
The agent's job is to compute the probabilities and keep the rules correct as pods come and
go. That is a control plane, not a data plane — the same split as the real thing.

### Balancing happens twice

The load balancer picks a **node**. That node's iptables rules pick a **pod**. Two
independent decisions. This mirrors a real cluster: an external load balancer spreads
traffic across nodes, and kube-proxy spreads it across pods inside each one.

### Two listeners, never one

The data plane and the control plane listen on different ports, and the distinction is
absolute: anything arriving on `:8080` is a client, anything on `:9000` is an agent. The
kernel demultiplexes by destination port for free, so the balancer never has to inspect a
payload to work out who is talking to it — which would break protocol-agnosticism and
deadlock against any protocol where the server speaks first.

### Pod addresses repeat across nodes

Every node hands out addresses from the same pod CIDR. Those addresses never leave their
node: a pod on another node is reached as `nodeIP:NodePort`, never by its pod IP. So there
is no conflict and no address coordination to do. Kubernetes allocates a distinct range per
node because it needs real pod-to-pod routing across nodes; this project does not, and gets
to skip the problem.

---

## Scope and honest limits

This is a toy. It is meant to teach me Linux networking and how Kubernetes is put together,
and it is written by hand on purpose. Which means:

- **Linux only, and root only.** Network namespaces, veth pairs, bridges and iptables are
  Linux. The agent needs root or `CAP_NET_ADMIN`. The load balancer runs anywhere.
- **Virtual machines, not containers.** A "pod" here is a network namespace with a process
  in it. There is no image, no filesystem isolation, no cgroups.
- **The CRI and the CNI are hardcoded.** Kubernetes has pluggable interfaces precisely so
  that Docker, containerd, Calico or Cilium can each do this differently. Here there is one
  way of creating a namespace and one way of writing an iptables rule, both baked into the
  agent. That is the opposite of what a real orchestrator does, and it is deliberate: the
  point is to write the thing an interface would normally hide.
- **iptables rules are written directly**, by shelling out to `iptables` and `ip`. Real
  implementations use netlink. Running this alongside Docker on the same host is asking for
  trouble, because Docker has its own opinions about the same chains.
- **No declarative state yet.** The control plane dispatches imperative events. There is no
  desired state to reconcile against, which is the single biggest thing separating this
  from an orchestrator. It is on the roadmap.

---

## Layout

```
orchard/
├── lb/        — control plane, L4 data plane, node registry
├── agent/     — node agent: registration, IPAM, pod netns + veth, iptables rules
├── docs/      — notes and design write-ups
└── TODO.md    — roadmap and open questions
```

---

## Setting it up

Each piece runs on its own machine, all on the same LAN. A home router is enough.

### Requirements

- Go 1.25 or newer on every machine.
- Linux on every node that runs an agent. Pods live in network namespaces, which only
  exist on Linux, and the agent needs root (or `CAP_NET_ADMIN`) to create them.
- `iptables` and `iproute2` on every node, plus a kernel with the `br_netfilter` module.
  The agent loads it and sets `net.bridge.bridge-nf-call-iptables=1`, so that traffic
  crossing the node bridge is seen by iptables at all.
- **A node with Docker installed needs care.** Docker sets the `FORWARD` policy to `DROP`
  and manages its own chains in the same tables. Use a clean VM.
- The load balancer itself runs on any OS.

### 1. Give the load balancer a fixed IP

Agents dial the control plane on `:9000` and clients dial the data plane on `:8080`, so the
load balancer's address must not change. The simplest way is a **DHCP reservation** on the
router:

1. Open the router's admin page (usually `http://192.168.1.1`) and go to the DHCP / LAN
   settings. Some routers call it "static IP" or "address reservation".
2. Find the load balancer machine in the list of connected devices and note its MAC address.
3. Add a reservation that ties that MAC to an IP, for example `192.168.1.50`.
4. Renew the lease on the machine so it picks up the address:
   - macOS: `sudo ipconfig set en0 DHCP`
   - Linux (NetworkManager): `nmcli connection down <name> && nmcli connection up <name>`
5. Check it: `ipconfig getifaddr en0` on macOS, `ip -4 addr` on Linux.

A reservation beats setting a static address on the machine itself: the router knows about
it and will never lend that IP to another device. If you set it by hand anyway, pick an
address outside the router's DHCP range.

> **Watch out for private MAC addresses.** macOS, iOS, Android and Windows can use a random
> Wi-Fi MAC per network, and some rotate it over time. If the MAC changes, the reservation
> stops matching and the machine gets a different IP. You can spot one because the second
> hex digit is `2`, `6`, `A` or `E` (e.g. `32:15:…`). On macOS: System Settings → Wi-Fi →
> Details on your network → Private Wi-Fi address → **Fixed** or **Off**. An Ethernet cable
> avoids the problem entirely.

Nodes don't need a reservation: they find the load balancer, not the other way round.

### 2. Bind to an address other machines can reach

The addresses the control plane and data plane listen on (in
[`lb/loadBalancer/lb.go`](lb/loadBalancer/lb.go)) must be reachable from the rest of the
LAN: either the reserved IP (`<LB_ADDRESS>:9000`) or all interfaces (`:9000`, `:8080`).
Never `localhost` or `127.0.0.1`, which only accept connections from the same machine.
Binding to all interfaces saves you from keeping the code and the router in sync.

### 3. Open the ports

Only the load balancer needs inbound ports, `8080/tcp` and `9000/tcp`.

- macOS: the first run asks whether to allow incoming connections — allow it. If you missed
  the prompt: System Settings → Network → Firewall → Options.
- Linux with ufw: `sudo ufw allow 8080/tcp && sudo ufw allow 9000/tcp`

### 4. Check it from another machine

Put the reserved IP in `.env.local` (git-ignored) and start the load balancer:

```sh
cp .env.example .env.local   # then set LB_ADDRESS
set -a; source .env.local; set +a
cd lb && go run ./cmd/lb
```

Then, from any other machine on the LAN:

```sh
nc -vz <LB_ADDRESS> 9000   # control plane
nc -vz <LB_ADDRESS> 8080   # data plane
```

Both should report the connection as succeeded. If they time out, go back to the firewall;
if they are refused, the process is not listening on that address.

---

## Why

Because reading about kube-proxy, veth pairs, conntrack and control loops is not the same
as having had to make them work. Everything here exists to be built by hand at least once.

## License

MIT