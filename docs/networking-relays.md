# Networking - Relays

## Problem statement

Peers are not always able to listen for incoming network connections for a number of reasons. Local private networks without or with misconfigured UPNP, public or mobile networks, corporate or other governmental networks behind firewalls etc.

These peers can still participate in the network but to a limited extend. By using any TCP based transport layer they will be able to send and receive objects to/from peers they can make connections to. This also leaves them unable to talk to peers who share the same limitations.

## Proposal: Relays

Peers can chose to act as relays and advertise the fact on the mesh. Peers unable to accept incoming connections can use those peers to receive objects from others. Such peers will look for available relays and advertise them as their relays in their connection infos. Relays peers should be used after dialing all the peer's addresses has failed.

In order for relays not to be able to read the objects being given to them, the payload objects themselves will have to be encrypted. To remove the need for the peers to perform a handshake a shared key will be [derived](https://blog.filippo.io/using-ed25519-keys-for-encryption) from the sender's and recipient's keys pairs allowing both peers to construct the encryption and decryption secret without any previous communication.

There are two main object types used for relaying messages.

- `DataForwardRequest` which is sent to relay peers
- `DataForwardEnvelope` which is sent by relay peers to recipients. The envelope contains an encrypted payload. It can also include a data forward request itself to enable chaining.

The private key that is used to derive the shared secret can be an ephemeral key created only for that specific recipient or even object. For now we are using the sender's actual key.

Multiple relays can be chained together in a later version of the protocol that will enable an onion-like routing mechanism. This is already possible with the current version of the relay protocol but figuring out which relays to use and make sure it's secure will need more work.

### Example

Let's say that **Alice** (sender) wants to send a `Whatever` object to **Bob** (recipient) through a relay, in this case **Roger**:

1. **Alice** has to get the nested objects ready, and so...
    - Constructs a `Whatever` object and encrypts it with `shared_key(Alice, Bob)`
    - Constructs a `DataForwardEnvelope`
        - Sender: `Alice`
        - Payload: encrypted `Whatever` from previous step
    - Constructs a `DataForwardRequest`
        - Recipient: `Bob`
        - Payload: plaintext `DataForwardEnvelope` from previous step
2. **Alice** sends the `DataForwardRequest` object to **Roger**
3. **Roger** receives the `DataForwardRequest` 
4. **Roger** sends the the nested `DataForwardEnvelope` object its recipient, **Bob**
5. **Bob** receives the `DataForwardEnvelope` and decrypts the `Whaterver` object from its payload using the sender attribute to derive the shared key
6. **Bob** now has the decrypted `Whatever` object
7. **Roger** sends back to **Alice** a `DataForwardResponse` informing them of whether they were able to deliver the object

### Example data forward request

![images/mesh-relays-object.svg](images/mesh-relays-object.svg)

## Messages

### Connection Info

```ndl
object nimona.io/peer.ConnectionInfo {
    publicKey nimona.io/crypto.PublicKey
    addresses repeated string
    relays repeated nimona.io/peer.ConnectionInfo
}
```

### Data Forward Request

```ndl
signed object nimona.io/mesh.DataForwardRequest {
    requestID string
    recipient nimona.io/crypto.PublicKey
    envelope nimona.io/object.Object
}
```

### Data Forward Envelope

```ndl
signed object nimona.io/mesh.DataForwardEnvelope {
        requestID string
    sender nimona.io/crypto.PublicKey
    data data
}
```

### Data Forward Response

```ndl
signed object nimona.io/mesh.DataForwardResponse {
        requestID string
        success bool
        error string
}
```
