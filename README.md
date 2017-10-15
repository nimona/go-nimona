# Nimona Fabric

Fabric is an attempt to make peer to peer networking a bit simpler.  
The reasoning behind this is that different peers have different ways of being
accessed. This usually ends up requiring both the clients and servers to have 
complex logic for handling connections, routing, etc.

## Addressing and Connecting

Fabric's addresses are a series of `/` delimited protocols and parameters that 
the client and server need to go through until connection is established.

Let's take for example a very simple ping-pong protocol. Client connects to 
server, sends a `ping` message, the server responds with `pong` and the 
connection is then closed.

The address for the server's ping handler would look something like 
`tcp:127.0.0.1:3000/tls/nimona:4D309D0C/ping`.

This tells the client that:

* it needs to connect via a TCP transport (`tcp:`)
* negotiate a TLS connection (`/tls`)
* verify that peer with ID `4D309D0C` is on the other side (`/nimona:`)
* and finally reach the `/ping` handler where it can send its `ping` message.

This allows building a number of "middleware" protocols that will wrap the 
primary connection as well as add contextual information to it.

The identity middleware will for example add to the connection information on
the local and remote peers, fingerprints, keys, etc. Other middleware/protocols
can use this information or add their own.

This is useful in cases where one server is for example hosting more than one
peers and the middleware down the chain need to know which peer this connection
is for.

### Address examples

* Simple connections: `tcp:127.0.0.1:3000/ping`
* Connections with identity negotiation/selection: `tcp:127.0.0.1:3000/nimona:4D309D0C/ping`
* Multiplexing connections: `tcp:127.0.0.1:3000/yamux/nimona:4D309D0C/ping`
* Relaying connections: `tcp:127.0.0.1:3000/yamux/relay/tcp:127.0.0.1:4000/nimona:4D309D0C/ping`
* Custom transports: `wss:foo.bar.com/nimona:4D309D0C/ping`

## Name resolution

There will be times when we will need to connect to peer whose transport address
we don't have. eg `nimona:4D309D0C/ping`.

A number of resolvers will be available to our Fabric that will resolve this 
address into a something we can actuall dial. eg `tcp:127.0.0.1:3000/nimona:4D309D0C/ping`.

Currently only a DHT resolver is in the works that will resolve a peer's ID into
one or more addresses that can be tried.