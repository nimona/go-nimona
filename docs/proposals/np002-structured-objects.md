---
np: 002
title: Objects
author: George Antoniadis (george@noodles.gr)
status: Draft
category: Objects
created: 2018-12-27
---

# Structured Objects

## Simple Summary

Objects expand on the work on [hinted object notation] and further define
the top levels of their structure in order to add some commonly used attributes
that applications can leverage.

## Problem Statement

In order for application and service developers to be able to identify, use, and
create data structures compatible with other applications we need to define a
basic list of well known attributed, some required some optional.

## Proposal

The top level of each object consists of three main attributes.

- `type:s` is an arbitrary string defining the type of the object's content.
- `metadata:m` are a fixed set of attributes that add extra info to the object.
- `data:m` is the content of the object itself.

```json
{
  "type:s": "type",
  "metadata:m": {
    "stream:s": "bah...",
    "owner:s": "bah...",
    "parents:m": [
      "*:as": ["bah..."],
      "some-type:as": ["bah..."],
    ],
    "_signature:m": {
      "alg:s": "hashing-algorithm",
      "signer:s": "bah...",
      "x:d": "bah..."
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
- `nimona.io/object.CID`
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
- `parents:as` (optional) Array of cids of parent objects, this is used
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

[hinted object notation]: ./np001-hinted-object-notation.md
[JSON]: https://www.json.org
[CBOR]: http://cbor.io
[MsgPack]: https://msgpack.org
[Cap-n-proto]: https://capnproto.org
[JSON-LD]: https://json-ld.org
[JSON-Schema]: https://json-schema.org
[Tagged JSON]: https://tjson.org
[Ben Laurie]: https://github.com/benlaurie
[Object hash]: https://github.com/benlaurie/objecthash
