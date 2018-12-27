# Nimona: Fluffy-Unicorn
Fluffy-Unicorn is a self-describing, type-hinted, encoding-
independent, language-agnostic, and consistently hash-able data format.
(_It sounds fancier than it really is, I promise._)

It does not define how data should be encoded or decoded when being transported. It can use any data format such as [JSON], [CBOR], [MsgPack], [Cap-n-proto], and others.

Its purpose is to be able to define (relatively) strongly-typed and self-defined objects that can be transported through any means and encoding while still being able to produce the same cryptographic hash.

## Overview
__First things first:__ Fluffy-Unicorn is based on the work of a number of projects that have been adapted to work together and fit the requirements of nimona. [JSON], [JSON-LD], [JSON-Schema], [Tagged JSON], and [Object hash].

## Types
Fluffy-Unicorn has its own list of basic types and structures, more closely resembling the ones of [JSON], other than the split of number to int/uint/float, and the addition of a binary type.

* Object - A collection of key/value pairs
* Array - An ordered list of values
* String
* Integer
* Unsigned Integer
* Float
* Data (binary data)
* Boolean

Continuing its comparison with JSON, there are some restrictions that are not present in the JSON format.

* The top-level Fluffy-Unicorn-Object must always be an object
* All values of an array must be of the same type.
* A number of key names are reserved.

## Type Hinting
In order to allow Fluffy-Unicorn-Objects to be encoding and language agnostic, the types of all values have to be able to survive multiple re-encodings between incompatible formats, languages, and libraries.

Since we can’t depend on each encoding format to keep the correct type, we provide a hint for the value’s type as part of the key. 

```json
{
  "some-string:s": "bar",
  "nested-object:o": {
    "unsigned-number-one:u": 1,
    "array-of-ints:a<i>": [-1, 0, 1]
  }
}
```

Hints for all of the supported types and structures:

* `s` string
* `i` int
* `u` uint
* `f` float
* `d` data (binary)
* `b` bool
* `a` array
  * `a<x>` array of type `x`
	  * `a<a<x>>` nested array of type `x`
* `o` object (map, dict)
* `o*` object reference _We’ll come back to this one later_

_Note: Type hinting is based on the [Tagged JSON] micro-format with a couple of minor changes; mainly terminology, the hints/tags used, the removal of sets and binary data formats, and the addition of the object reference type._

## Self-description & Schemas
Fluffy-Unicorn takes advantage of the reserved key `$schema` to allow objects to define their schemas. Defining a schema is not required.

The value of `$schema` is an object that defines and documents the structure of the object.

```json
{
  "$schema:o": {
    "description:s": "Hello world object",
    "properties:a<s>": [
      "body:s"
    ]
  },
  "body:s": "Hello world!"
}
```

_Note: The schema definition is still work in progress and will remain like this for a while._

## Object Hashing
Fluffy-Unicorn-Object-Hashing is based off the JSON version of [Ben Laurie]’s [Object hash] and provides a way to consistently hash objects that works across languages and encodings.

Given an object, it will sort its key/value pairs based on their key, loop through them, hash each key, and then using the type hint in the key, normalise and hash its value (after prefixing them with the type’s hint). 
It will finally gather all key and value hashes, and hash them all together to produce the object’s hash.

### Referencing and Redaction
If while looping through the key/value pairs we bump into an object (type hint of `o`), the object hashing function will be called for that whole object and the whole object’s hash will be used.

This allows replacing nested objects with their object hash without affecting the result of the root object’s hash.

Let’s say that the following object results in a hash of `454fc56c03071d9`.
```json
{
  "some-string:s": "bar",
  "nested-object:o": {
    "unsigned-number-one:u": 1,
    "array-of-ints:a<i>": [-1, 0, 1]
  }
}
```

While the hash of its nested object is `3b4ba8e4fd82231`.
```json
{
  "unsigned-number-one:u": 1,
  "array-of-ints:a<i>": [-1, 0, 1]
}
```

We can replace the nested object with its hash, and still get the same hash as when we had the full object in its place; `454fc56c03071d9`.
```json
{
  "some-string:s": "bar",
  "nested-object:o*": "3b4ba8e4fd82231"
}
```

Notice that now the type hint has been changed from `o` (object) to `o*` (object reference). This is what allows the hashing function to correctly use the value as a hash.

_Note: As we start using Fluffy-Unicorn-Objects to create more complex data structures and graphs, object references will be used to reduce the size of repeated information between different objects._

_Note: Using references for example, the schema example could be simply pointing to the schema instead of including it; and it would still produce the same hash, even if we didn’t have access to the schema itself._

```json
{
  "$schema:o*": "c3c7820210c5810d",
  "body:s": "Hello world!"
}
```

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
