# Hinting

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

## Hashing

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

_Note: when hashing map keys, the type should be always changed to `r`.
So that `foo:s` would become `foo:r` before hashing.`

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

We can replace the nested object with its hash, and still get the same hash as when we had the full object in its place; `454fc56c03071d9`.

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

_Note: Using references for example, the schema example could be simply
pointing to the schema instead of including it; and it would still produce
the same hash, even if we didn’t have access to the schema itself._

```json
{
  "_schema:r": "c3c7820210c5810d",
  "body:s": "Hello world!"
}
```

## References

* [Tagged JSON]

[Tagged JSON]: https://tjson.org
