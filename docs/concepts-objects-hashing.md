# Consistent Hashing

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

## References

* [Ben Laurie]
* [Object hash]

[Ben Laurie]: https://github.com/benlaurie
[Object hash]: https://github.com/benlaurie/objecthash
