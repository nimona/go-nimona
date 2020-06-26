# Schemas

_Note: Schemas are not currently in use, they are still under consideration,
everything in this page should be considered work in progress._

Object takes advantage of the reserved key `_schema` to allow objects to define
their schemas. Defining a schema is not required.

The value of `_schema` defines the structure of the object itself.

```json
{
  "_schema:m": {
    "description:s": "Hello world object",
    "properties:as": [
      "body:s"
    ]
  },
  "data:m": {
    "body:s": "Hello world!"
  }
}
```

## Formats

## Map

```json
{
    "name:m": {
        "first:s": "...",
        "last:s": "..."
    },
    "titles:as": [
        "..."
    ],
    "phonenumbers:am": [{
        "alias:s": "...",
        "phonenumber:s": "..."
    }]
}
```

### String, Verbose

```nsd
name:m (
    first:s
    last:s
)
titles:as
phonenumbers:am (
    alias:s
    phonenumber:s
)
```

### String, Concise

```nsd
name:m(first:s,last:s),titles:as,phonenumbers:am(alias:s,phonenumber:s)
```
