# Structured Objects

Objects are expand on the work of the
[hinted object notation](np001-hinted-object-notation.md) and further define
the top levels of their structure in order to add some commonly used attributes
that applications can leverage.

## Structure

The top level of each object consists of three main attributes.

- `type:s` is an arbitrary string defining the type of the object's content.
- `metadata:m` are a fixed set of attributes that add extra info to the object.
- `data:m` is the content of the object itself.

```json
{
  "type:s": "type",
  "metadata:m": {
    "stream:s": "0x...",
    "owner:s": "0x...",
    "parents:as": [
      "0x...",
      "0x..."
    ],
    "_signature:m": {
      "alg:s": "hashing-algorithm",
      "signer:s": "0x...",
      "x:d": "0x..."
    }
  },
  "data:m": {
    "foo:s": "bar"
  }
}
```

### Type

::: danger
Types are currently a way of moving forward, it's quite possible they will be
deprecated in the future in favor once schemes are introduced.
:::

### Well known types

- `nimona.io/crypto.PublicKey`
- `nimona.io/crypto.PrivateKey`
- `nimona.io/object.Hash`
- `nimona.io/peer.ConnectionInfoInfo`
- `nimona.io/peer.ConnectionInfoRequest`
- `nimona.io/peer.ConnectionInfoResponse`
- `nimona.io/object.Certificate`
- `nimona.io/object.CertificateRequest`
- `nimona.io/exchange.ObjectRequest`
- `nimona.io/exchange.ObjectResponse`
- ...

### Metadata

- `owner:s` (optional) Public keys of the owner of the object.  
- `stream:s` (optional) Root hash of the stream the object is part of.  
- `parents:as` (optional) Array of hashes of parent objects, this is used
  for streams
- `_signature:m` (optional) Cryptographic signature by the owner.

Additional metadata will be added in regards to access control and schema
specification.


## References

- [JSON]
- [CBOR]
- [MsgPack]
- [Cap-n-proto]
- [JSON-LD]
- [JSON-Schema]
- [Tagged JSON]
- [Ben Laurie]
- [Object hash]

[JSON]: https://www.json.org
[CBOR]: http://cbor.io
[MsgPack]: https://msgpack.org
[Cap-n-proto]: https://capnproto.org
[JSON-LD]: https://json-ld.org
[JSON-Schema]: https://json-schema.org
[Tagged JSON]: https://tjson.org
[Ben Laurie]: https://github.com/benlaurie
[Object hash]: https://github.com/benlaurie/objecthash
