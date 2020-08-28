# Objects

Objects are the building blocks of nimona.
They make up mostly everything, both permanent and ephemeral.

The are based on the work of a number of projects that have been adapted to
work together and fit the requirements of nimona.
[JSON], [JSON-LD], [JSON-Schema], [Tagged JSON], and [Object hash].

They do not define how data should be encoded or decoded when being transported.
We can use any data format such as [JSON], [CBOR], [MsgPack], [Cap-n-proto],
and others.

Their purpose is to be able to define structured, (relatively) strongly-typed
and self-describing data that can be transported through any means and encodings
while still being able to consistently produce the same cryptographic hash.

Objects have its own list of basic types and structures, more closely resembling
the ones of [JSON], other than the separation of number to int/uint/float, and
the addition of the data and reference types.

* Map - A collection of key/value pairs
* Array - An ordered list of values
* String
* Integer
* Unsigned Integer
* Float
* Data (binary data)
* Boolean
* Reference - A hash that replaces a value, used for redacting values.

Continuing its comparison with JSON, there are some restrictions that are not
present in the JSON format.

* The top-level Object must always be a map.
* All values of an array must be of the same type.
* A number of key names are reserved.
* All keys must end with the type hint of their value.

## Structure

The top level of each object defines its metadata.
Metadata are fixed and application developers should not be making up new ones.
All type-specific data are grouped under the `data:m` attribute.

```json
{
  "type:s": "type",
  "stream:s": "x0...",
  "owner:s": "x0...",
  "parents:as": [
    "x0...",
    "x1..."
  ],
  "data:m": {
    "foo:s": "bar"
  },
  "_signature:m": {
    "alg:s": "hashing-algorithm",
    "signer:s": "x0...",
    "x:d": "x0..."
  }
}
```

### Metadata

* `type:s` Object type
* `owner:s` (optional) Public keys of the owner of the object.  
* `stream:s` (optional) Root hash of the stream the object is part of.  
* `parents:as` (optional) Array of hashes of parent objects, this is used
  for streams
* `data:m` Map of arbitrary data.  
  Currently this can only be a map, but we're considering allowing any type.
* `_signature:m` (optional) Cryptographic signature by the owner.

Additional metadata will be added in regards to access control and schema
specification.

## Features

Objects provide a number of interesting features:

### [Hinted types](concepts-objects-hinting.md)

Keys/value pairs in object maps include a hint in the key about the type of
the value.

This helps in maintaining the expected type of values through different
transport encodings and programming languages.

### [Consistent hashing](concepts-objects-hashing.md)

Instead of defining a consistent way of encoding an object into bytes and
hashing the result, we are instead hashing each attribute of an object
independently, collect their hashes and hash them again.

This allows decoupling the transport encoding of an object from its hashing.

In addition it allows for providing additional attributes in the object that
do not have to be included in the hash, such attributes are prefixed with an
underscore.

Finally it allows for the redaction of attribute values from the object without
changing the hash of the complete object.
Since each value is hashed individually, the sender of an object can instead
of sending the value of an attribute send its hash.
The recipient will still be able to generate the hash for the object even
without having access to the underlying value.

### Self describing

All objects include a type that hints at what the object is about.
The type consists of an namespace and type name.
A number of "well known" types will exist that have no namespace.

_Note: Currently well known types have the `nimona.io` namespace but that will
be removed soon enough._

#### Well known types

* `nimona.io/crypto.PublicKey`
* `nimona.io/crypto.PrivateKey`
* `nimona.io/object.Hash`
* `nimona.io/peer.PeerInfo`
* `nimona.io/peer.PeerRequest`
* `nimona.io/peer.PeerResponse`
* `nimona.io/peer.Certificate`
* `nimona.io/peer.CertificateRequest`
* `nimona.io/exchange.ObjectRequest`
* `nimona.io/exchange.ObjectResponse`
* ...

## References

* [JSON]
* [CBOR]
* [MsgPack]
* [Cap-n-proto]
* [JSON-LD]
* [JSON-Schema]
* [Tagged JSON]
* [Ben Laurie]
* [Object hash]

[JSON]: https://www.json.org
[CBOR]: http://cbor.io
[MsgPack]: https://msgpack.org
[Cap-n-proto]: https://capnproto.org
[JSON-LD]: https://json-ld.org
[JSON-Schema]: https://json-schema.org
[Tagged JSON]: https://tjson.org
[Ben Laurie]: https://github.com/benlaurie
[Object hash]: https://github.com/benlaurie/objecthash
