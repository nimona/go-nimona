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
someone to retrieve a stream they must at the very least know that hash.

A hypothetical root is an object that identifies a stream that can be assumed
exists given the type of stream and the identity that would have created it.
This allows peers to find streams unique to an identity without having to
somehow learn of their existence.

In addition to the type and identity's key,s the constructed object will also
need to define a fixed policy that will define the owner as the only one that
has access to modify this stream.
This ensures that hypothetical roots can only be created for initially
private streams.
The owner can decide to update the policy in subsequent updates.

On their own they do do not hold additional information, but are expected to
provide it via the related objects that must be signed by identity that was
assumed.

---

Identity profiles are a good example for such a use-case.

Peer `a11` is wants the profile steam for the identity `f00`.
All it has to do is construct the hypothetical root, calculate its hash,
and find its providers on the network.

```json
{
  "@type:s": "nimona.io/profile.Created",
  "@identity:s": "f00",
  "@policy:o": {
    "subjects:as": ["f00"],
    "resources:as": ["*"],
    "action:s": "ALLOW"
  }
}
```

The hash of this object is `oh1.9KQhQ4UGaQPEyUDAAPDmVJCoHnGtJY7Aun4coFATXCYK`
and the peer can now lookup the providers for this object, and sync the
remaining stream.

---

The NDL for defining hypothetical roots is as follows.
Additional objects can be defined in the stream as needed, but the hypothetical
root cannot have additional properties.

```ndl
stream nimona.io/profile {
    hypothetical root object Created { }
    signed object NameUpdated {
        nameFirst string
        nameLast string
        dependsOn repeated relationship
    }
}
```

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
