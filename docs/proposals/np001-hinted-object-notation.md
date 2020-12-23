---
np: 001
title: Hinted Object Notation
author: George Antoniadis (george@noodles.gr)
status: Draft
category: Objects
created: 2018-12-27
---

## Simple Summary

A relatively strongly typed and self-describing data format that can be
transported through any means and encoding.

## Problem Statement

One of the core design decisions for nimona was to base the protocol around
messages/events whatever you wanna call them.

This means that all clients must be able to produce and consume structured data.
Many of the existing data formats could be used but present some downsides for
our use-cases.

* [JSON], YAML, XML provide human friendly output which we feel is important
  but do not have strong types and programming languages provide inconsistent
  results.
* [CBOR], [MsgPack], [Cap-n-proto] on the other side, provide binary
  serialization that is machine readable, but are strongly typed and will
  provide consistent results in all languages.
* Protobuf provides binary output and can even support self described data
  using Any messages.

Because of the nature of the protocol we're building the format of the data will
be something that will be hard to change in the future mostly due to the fact
that content hashes will depend on the serialized data and thus the format we
chose.

## Proposal

In order to find a format that best suites our needs, we are proposing "Hinted
Object Notation".

This work is based on the amazing work of a number of projects that have been
adapted to work together and fit the requirements of nimona.
[JSON], [JSON-LD], [JSON-Schema], [Tagged JSON], and [Object hash].

It does not define how data should be encoded or decoded when being transported.
We can use any data format such as [JSON], [CBOR], [MsgPack], [Cap-n-proto],
and others.

Instead its purpose is to be able to define structured, (relatively) strongly-
typed and self-describing data that can be transported through any means and
encodings while still being able to consistently produce the same cryptographic
hash.

It has its own list of basic types and structures, more closely resembling
the ones of [JSON], other than the separation of number to int/uint/float, and
the ability to add more types as we move forward.

* Map - A collection of alphabetically sorted key/value pairs
* Array - An ordered list of values
* String
* Integer
* Unsigned Integer
* Float
* Data (binary data)
* Boolean

Continuing its comparison with [JSON], there are some restrictions that are not
present in the JSON format.

* The top-level Object must always be a map.
* All values of an array must be of the same type.
* A number of key names are reserved.
* All keys must end with the type hint of their value.

### Features

Objects provide a number of interesting features:

#### Type hinting

In order to allow Objects to be encoding and language agnostic, the types of
all values have to be able to survive multiple re-encodings between incompatible
formats, languages, and libraries.

Since we can’t depend on each encoding format to keep the correct type, we 
provide a hint for the value’s type as part of the key.

```json
{
  "some-string:s": "bar",
  "nested-object:m": {
    "unsigned-number-one:u": 1,
    "array-of-ints:ai": [-1, 0, 1]
  }
}
```

Hints for all of the supported types and structures:

* `m` map
* `s` string
* `i` int
* `u` uint
* `f` float
* `d` data (binary)
* `b` bool
* `a` array
  * `ax` array of type `x`
    * `aax` nested array of type `x`
* `r` reference _We’ll come back to this one later_

_Note: Type hinting is based on the [Tagged JSON] micro-format with a couple
of minor changes; mainly terminology, the hints/tags used, the removal of sets
and binary data formats, and the addition of the object reference type._

#### Consistent hashing

Instead of defining a consistent way of serializing a map into bytes and
hashing the result, we are instead hashing each attribute of the map
independently, collect their hashes and hash them again.

This allows decoupling the transport encoding of a map from its hash.

In addition it allows for providing additional attributes in the object that
do not have to be included in the hash, such attributes are prefixed with an
underscore.

Finally it allows for the redaction of attribute values from the object without
changing the hash of the complete object.
Since each value is hashed individually, the sender of an object can instead
of sending the value of an attribute send its hash.
The recipient will still be able to generate the hash for the object even
without having access to the underlying value.

Object-Hashing is based off the JSON version of [Ben Laurie]’s [Object hash]
and provides a way to consistently hash objects that works across languages
and encodings.

* `s` strings get converted into their UTF8 byte equivalent and hashed.
* `i` and `u` integers get converted into strings and then hashed as strings.
* `f` floats get converted into their IEEE 754 representation and hashed.
* `b` bools get converted to single bytes 0/1 and hashed.
* `d` data are simply hashed as they are.
* `r` references do not get hashed again, they are returned untouched.
* `a` for arrays, we go through all their values, hash them according to their
  hint, concat all the hashes, and hash the result.
  `h( h(v0) + h(v1) + ...)`
* `m` for maps, we loop through the sorted key/value pairs, hash each key and
  value, and append them to the previous pair. Finally hash the result.
  `h( h(k0) + h(v0) + h(k1) + h(v1) + ...)`  

_Note: keys starting with an underscore (`_`) should be ignored and not be part
of the hash._

_Note: at this moment only maps can be redacted._

_Note: when hashing maps, their type should be always changed to `r`.
So that `foo:m` would become `foo:r` before hashing.`_

### Referencing and Redaction

References enable redacting values of an object without affecting its hash.

Let’s say that the following object results in a hash of `454fc56c03071d9`.

```json
{
  "some-string:s": "bar",
  "nested-object:m": {
    "unsigned-number-one:u": 1,
    "array-of-ints:ai": [-1, 0, 1]
  }
}
```

While the hash of its nested object is `3b4ba8e4fd82231`.

```json
{
  "unsigned-number-one:u": 1,
  "array-of-ints:ai": [-1, 0, 1]
}
```

We can replace the nested object with its hash, and still get the same hash
as when we had the full object in its place; `454fc56c03071d9`.

```json
{
  "some-string:s": "bar",
  "nested-object:r": "3b4ba8e4fd82231"
}
```

Notice that now the type hint has been changed from `m` (object) to `r`.
While we are normally hashing the map, we are also changing the hint to `r`.

_Note: As we start using Objects to create more complex data structures and
graphs, object references will be used to reduce the size of repeated
information between different objects._

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
