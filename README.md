# Nimona Fabric

Fabric is a networking library that provides some very opinionated features
targeting mainly peer to peer and decentralized systems.  

- Verbose network address notations that expose tranports, protocols, middleware, etc
- Protocols multiplexing and negotiation over the same transport layer
- Optional peer and service discovery
- Optional routing connections through proxy peers

That being said, there is nothing prohibiting its use in other applications, 
eg microservices.  

## Problem statement

In peer to peer networks each peer is more or less unique.  
Peers want or have to use different transport layers, different protocols,
different encoding schemes etc.  

The reasoning behind this is that different peers have different ways of being
accessed. This usually ends up requiring both the clients and servers to have 
complex logic for handling connections, routing, etc.

## Addressing and Connecting

Fabric's addresses are a series of `/` delimited protocols and parameters that 
the client and server need to go through until connection is established.

Let's take for example a very simple ping-pong protocol. Peer A connects to 
peer B, sends a `ping` message, peer B responds with `pong` and the connection 
is then closed.

Assuming we only know the peer's id, the address peer A will have to connect 
to peer B will be something like `peer:B/ping`.

Peer B will try to dial `peer:B/ping` with Fabric.  
Since Fabric won't have transport for `peer:B` it will try to resolve this
using a number of resolvers. (Initially a registry and a Kademlia DHT resolver
will be available but more can be added in the future).

Let's say that the DHT resolver finds a couple of addresses that peer B has been
advertising for itself:

* `tcp:172.16.0.1:3000/tls/yamux`
* `ws:172.16.0.1:3001/tls`
* `peer:C/tls/yamux/relay:peer:B/tls/yamux`

The addresses show exactly what needs to happen, to these addresses the original
address will appened so they'll look something like:

* `tcp:172.16.0.1:3000/tls/yamux/peer:B/ping`
* `ws:172.16.0.1:3001/tls/yamux/peer:B/ping`
* `peer:C/tls/yamux/relay:peer:B/tls/yamux/peer:B/ping`

In the first two, connect to a peer with either TCP (`tcp`) or WebSockets 
(`ws`), negotiate TLS (`tls`), negotiate the yamux multiplexer (`yamux`) so we
can open new bi-directional connections on the same socket, make sure we are 
talking to the corrrect peer (`peer`), and finaly negotiate the protocol we
need to reach (`ping`).

The third address has a bit more information but the gist is the same. It will
connect to a third, relay peer, do some stuff, and eventually create a relay
to the original target. This will allow peer B to be accessible even if it is 
not publicly accessible or properly NATed.

This allows building a number of "middleware" protocols that will wrap the 
primary connection as well as add contextual information to it.

The `peer` middleware will for example add to the connection information on
the local and remote peers, fingerprints, keys, etc. Other middleware/protocols
can use this information or add their own.

This is useful in cases where one server is for example hosting more than one
peers and the middleware down the chain need to know which peer this connection
is for.

### Address examples

* Simple connections: `tcp:127.0.0.1:3000/ping`
* Connections with identity negotiation/selection: `tcp:127.0.0.1:3000/peer:4D309D0C/ping`
* Multiplexing connections: `tcp:127.0.0.1:3000/yamux/peer:4D309D0C/ping`
* Relaying connections: `tcp:127.0.0.1:3000/yamux/relay:peer:4D309D0C/peer:4D309D0C/ping`
* Custom transports: `wss:foo.bar.com/peer:4D309D0C/ping`

## Name resolution

There will be times when we will need to connect to peer whose transport address
we don't have. eg `peer:4D309D0C/ping`.

A number of resolvers will be available to our Fabric that will resolve this 
address into a something we can actuall dial. eg `tcp:127.0.0.1:3000/peer:4D309D0C/ping`.

Currently only a DHT resolver is in the works that will resolve a peer's ID into
one or more addresses that can be tried.