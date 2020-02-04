# Streams

While objects on their own are useful for creating permanent content-addressable
data, there are very few applications where data never get updated.
Streams allow developers to create complex applications by applying event driven
and event sourcing patterns using graphs of immutable objects.

Objects in a stream form a directed acyclic graph (DAG) by allowing each of the
objects to reference other objects it depends or knows of.
This graph can then be serialized into linear series of objects that can be
replayed consistently by all everyone that has the same representation of the
graph.

Streams are identified by the hash of their root object.
This means that even though each of their objects is content-addressable; the
stream as a whole is not, as its root hash (and thus identifier) does not
change when more objects are added to the graph.
The benefit of this is that there is no need to find a way to reference the
stream as it changes.
The downside is that you do not really know if you have actually  received the
whole stream and if people are not holding back at you.

## Relationships

TODO

## Access control

TODO

## Hypothetical roots

Since streams are identified by the hash of their root object, in order for
someone to retrieve a stream they must at the very least that hash.

This is mostly fine when you are trying to retrieve a blog post or a
conversation, but makes it complicated if you're looking for a stream for a
person's public profile for example.

To allow for this an addition type of root objects exist, defined as
`hypothetical`.
This are object that peers can assume exist and only consist of a type, and the
key of the identity that would have generated them.
On their own they do do not hold additional information, but are expected to
provide it via the related objects that much be signed by identity that was
assumed.

```ndl
stream nimona.io/profile {
    hypothetical root object { }
    signed object NameUpdated {
        nameFirst string
        nameLast string
        dependsOn repeated relationship
    }
}
```

The following would be the hypothetical root for a profile stream for the
identity with the public key `foo`.

```json
{
  "@type:s": "nimona.io/profile.Created",
  "@identity:s": "foo"
}
```

A peer looking for the profile of this identity, it can generate this object on
its own, calculate its hash, and find providers for it on the network and
retrieve the stream from them.

## Synchronization

TODO

## Literature

* <https://docs.textile.io/concepts/threads>
* <https://www.streamr.com/docs/streams>
* <https://holochain.org>
* <https://github.com/textileio/go-textile/issues/694>
* <https://tuhrig.de/messages-vs-events-vs-commands>
* <https://www.swirlds.com/downloads/SWIRLDS-TR-2016-01.pdf>
* <https://arxiv.org/pdf/1710.04469.pdf>
* <http://archagon.net/blog/2018/03/24/data-laced-with-history>
