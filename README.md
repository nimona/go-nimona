# Nimona

**WARNING**: Nimona is still in its very early stages of design and development
and will stay like this for a while.

# Protocol Overview

Nimona enables users to own their data and relationships, while at the same time allowing them to consume services without being restricted to a single provider.

# Concepts

## **Document**

Documents are the basic building blocks for our data needs.
Their root node is a key-value map where keys are always strings, and values can be any of the following types: `map`, `string`, `uint64`, `int64`, `bool`, `bytes`, and lists of the same values.

They contain a `$type`, `$context`, `$schema`, and `$metadata` on their top level, as well as any additional type-specific attributes.

```yaml
$type: 'test/fixture'
$context: 'nimona'
$metadata:
  owner: { ... }
  root: { ... }
  parents:
    head: [{ ... }]
  sequense: 0
  timestamp: '<time-created>'
  _signature: { ... }
foo: 'bar'
```

Documents are individually immutable, but can be linked together to form document streams. In streams, the root object defines the schema, policy, and initial state. Subsequent objects in the stream can mutate the state as long as these changes adhere to the schema and policy.

### Types and hints

In order to work around issues with encoding and decoding documents, a hinting system in introduced for maps. Type hints are appended to the map key in the form of `<key>:<hint>` and help disambiguate the type of the value. For example when encoding a byte array in JSON, the resulting value will be a string representation bytes encoded as base64. When trying to decode this into a document it’s unclear if these are bytes or a string. Similar issues exist with numbers and other types that are all serialised into the same format.

### **Document Hash**

The identifier for a document, created using a variation of Ben Laurie’s [object hash](https://github.com/benlaurie/objecthash/).
The hash itself is a document of type `core/document/hash`.
Its string representation is `nimona://document:doc+sha256:<base58(hash)>`.

```yaml
$type: 'core/document/hash'
alg: 'doc+sha256'
dig: [...]
```

### **Document Patch**

While documents are immutable, patch documents can link back to  used to form document streams. In streams, the root object defines the type, schema, permissions, and initial state. Subsequent patches in the stream can mutate the state as long as they adhere to the schema and policy.

Document patches are of type `core/document/patch` and point to either a root document or previous patch, and contain a list of operations that modify the root document.

```yaml
$type: 'core/document/patch'
$metadata:
  parents: ['nimona://doc:<root-document-id>']
op: 'replace'
path: '/someString'
value: 'foobaz'
```

### **Document Graph**

A directed acyclic graph starting with a (non-patch) document as a root and zero or more document patches as children. The patches can be topologically sorted, and their operations applied to the root document in order to get its final state.

They are identified by the hash of their root document.
Their string representation is `nimona://graph:<base58(root-document-hash)>`

## **Peer**

Any instance of application or service that wishes to use nimona.

### Peer Key

Each peer must have a unique public key that other use to connect to them and identify them. Currently only ed25519 keys are supported. They are documents of type `core/peer/key` but can also be printed in string form as `nimona://peer:key:<base58([keyType,kty,d,x])>`.

```yaml
$type: 'core/peer/key'
publicKey:
  $type: 'core/crypto/ed25519.public'
  d: [...]
  kty: 'OKP'
  x: [...]
```

### Peer Address

Peers can accept incoming connections can advertise their peer addresses in order for others to be able to connect to the. In order for a peer to be dial-able, they must advertise their public key, protocol, ip/hostname, and port. Peer addresses can be also printed in string form as `nimona://peer:addr:<publicKey>@<transport>:<address>`.

```yaml
$type: 'core/peer/address'
address: 'localhost:3000'
publicKey:
  $type: 'core/crypto/ed25519.public'
  d: [...]
  kty: 'OKP'
  x: [...]
transport: 'utp'
```

## Identity

Each individuals, company, or service taking part in nimona has their own long-term identity they can use for identification across all application.

Shorthand   `nimona://id:<base58(root-document-id)>`

## Identity Document

An identity is a document stream that enabled users to rotate their keys as needed, and delegate or revoke access to third party applications and other identities.

```yaml
$type: 'core/identity'
use: 'provider' # provider, user
keys: [{$type: 'core/crypto/ed25519.public', ... }]
next: [{$type: 'core/document/hash', ... }]
witnesses: [{$type: 'core/identity', ... }]
delegates: [{$type: 'core/identity', ... }]
providers: [{$type: 'core/identity', ... }]
```

## Identity Alias

Aliases allow linking to identities using more user friendly names such as hostnames.

Shorthand   `nimona://id:alias:<hostname>/<handle>`

## Network

Networks provide users with services such as discovery, storage, and other services. They can be both public as well as private networks, requiring users to either register, or be invited in order to join. Users can be part of more than one network a given time, and they can select which networks their data live on and which network to use for other services.

```yaml
$type: 'core/network'
$metadata:
  owner: { $type: 'core/identity', ... }
hostname: 'romdo.io'
requiresRegistration: true
requiresToken: true
```

Networks can be also be formatted as `nimona://network:<base58(root-document-id)>`.

### Network Alias

In order to make networks more humanly identifiable we’re using domain names for identification, and use DNS Link to resolve their domain to a number of peers we can use to connect to.

`nimona://network:alias:romdo.io`

### Identity Alias

Some networks offer the option to users to create human readable aliases for their identities. These take the form `nimona://identity:alias:<network-hostname>/<handle>` where handle is a unique (to the network) name.

When applications are provided with an identity alias, they can ask the network to resolve the handle and get back the underlying identity.

# Protocol Specification

## Networking

### Transport

The primary protocol currently considered is **μTP** ([Micro Transport Protocol](https://en.wikipedia.org/wiki/Micro_Transport_Protocol)), a UDP-based variant of BitTorrent’s protocol intended to mitigate poor latency and other congestion control problems found in conventional BitTorrent over TCP, while proving reliable, ordered delivery.

### Encryption

In order to connect to a target peer, a number of things are required:

1. Their supported protocols (currently only µTP is supported)
2. Their addresses (host and port)
3. Their ed25519 public key

Once a connection has been established, the following handshake needs to take place. Note that even all peers are equal we are using the terms client/server for denote the peer which initiated the connection (client) and the peer who accepted the connection (server).

1. Client calculates the and sends the x25519 curve point (32 bytes) to server
2. Server receives curve from client
3. Both derive a shared key using BLAKE-2b as a key derivation function
4. Both encrypt further communication with AES 256-bit GCM using the shared key, with a nonce counter increasing for every incoming/outgoing message

### mRPC: Low-level wire protocol

Once an encrypted connection has been established, a low level RPC protocol can be used to send and receive messages and replies.

1. Encrypted messages are prefixed with an uint32 denoting the message's length.
2. The decoded message content is prefixed with an uint32 designating a sequence number
3. The sequence number is used as an identifier to identify requests/responses
4. The sequence number 0 is reserved for requests that do not expect a response

### oRPC: High-level wire protocol

Once a connection has been established, communication happens via a basic RPC using serialised tilde objects.

oRPC supports three options:

1. Request `rpc SayHello(HelloRequest)`
2. Request/response `rpc SayHello(HelloRequest) returns (HelloResponse)`
3. Request/stream-response `rpc SayHello(HelloRequest) returns (stream HelloResponse)`
