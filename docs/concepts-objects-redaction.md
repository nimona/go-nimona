# Referencing and Redaction

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

We can replace the nested object with its hash, and still get the same hash as
when we had the full object in its place; `454fc56c03071d9`.

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

* [Ben Laurie]
* [Object hash]

[Ben Laurie]: https://github.com/benlaurie
[Object hash]: https://github.com/benlaurie/objecthash
