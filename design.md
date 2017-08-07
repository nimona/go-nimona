## Design Document - v0

This is an oversimplified version of a kademlia/mainline DHT.

For at least the v0 of our DHT we only care about storing and retrieving information about other DHT clients/servers; which we will call Peers.

Mainline separates peers from Peers as one is the DHT client/server and the other is a BitTorrent provider.  
We do not care about anything more than the DHT clients/servers for now, so this should simplify things a bit.

## Terminology

* `peer` - A participating client/server in the DHT network.
* `k-buckets` - The routing table keeps a list of known GOOD peers.
* `alpha` - A small number representing the degree of parallelism in network calls, usually 3.
* `B` - The size in bits of the keys used to identify peers; in Kademlia this is 160, the length of a SHA1 digest.
* `k` - The maximum number of contacts stored in a bucket; this is 20 in kademlia and 8 in mainline.
* `tRefresh` - The time after which an otherwise unaccessed bucket must be refreshed; 3600s in Kademlia.
* `refresh` - The process of getting fresh peers into unaccessed buckets.

## Peer

Each peer has an ID, for the purpose of this document we assume that the each ID is unique in the network.

Kademlia uses 160-bit identifiers (SHA1) for peer IDs, we could explore the possibility of moving to some form of multihash at some point if we decide that this DHT does not have to be compatible with Kad.

### XOR Metric & Distance Calculation

Kademlia's operations are based upon the use of exclusive OR, XOR, as a metric. The distance between any two IDs x and y is defined as

    distance(x, y) = x XOR y

### The K-Bucket List (Routing Table)

A Kademlia peer organizes its contacts (other peers known to it) in buckets which hold a maximum of k peers. These are known as k-buckets.

The buckets are organized by the distance between the peer and the contacts in the bucket. Specifically, for bucket j, where 0 <= j < k, we are guaranteed that

      2^j <= distance(peer, contact) < 2^(j+1)

Given the very large address space, this means that bucket zero has only one possible member, the key which differs from the peerID only in the high order bit, and for all practical purposes is never populated, except perhaps in testing. On other hand, if peerIDs are evenly distributed, it is very likely that half of all peers will lie in the range of bucket B-1 = 159.

Within buckets contacts are sorted by the time of the most recent communication, with those which have most recently communicated at the end of the list and those which have least recently communicated at the front, regardless of whether the peer or the contact initiated the sequence of messages.

Whenever a peer receives a communication from another, it updates the corresponding bucket. If the contact already exists, it is moved to the end of the bucket. Otherwise, if the bucket is not full, the new contact is added at the end. If the bucket is full, the peer pings the contact at the head of the bucket's list. If that least recently seen contact fails to respond in an (unspecified) reasonable time, it is dropped from the list, and the new contact is added at the tail. Otherwise the new contact is ignored for bucket updating purposes.

## Protocol

### Data structures

This is the minimum requirements for the peer structure:

* `Peer`
  * `ID` - Unique identifier for the Peer.
  * `Addresses` - Should provide one or more addresses by which to contact the peer.

## Messaging

Protocol design is up for debate at this point.  

* `PING` - Verify that a peer is still alive.  
  The recipient of the PING must update the bucket corresponding to the sender; and, if there is a reply, the sender must update the bucket appropriate to the recipient.  
* `FIND_PEER` - The recipient of the request will return the k peers in his own buckets that are the closest ones to the requested ID.

Each message includes a random value from the initiator (of length B in kad) which the recipient must include in their response.  
This ensures that when the response is received it corresponds to the request previously sent.

## Joining the Network

A peer joins the network as follows:

* It inserts the value of some known (bootstrap) peer `A` into the appropriate bucket as its first contact.
* It does a `FIND_PEER` for their peer ID.
* It refreshes all buckets.

### Peer lookup (`FIND_PEER`)

The lookup begins by creating a shortlist of an alpha number of contacts closest to the ID being searched for. Resulting contacts can be from one or more k-buckets. The contact closest to the target peer, `closestPeer`, is noted.

The peer then sends parallel, asynchronous `FIND_PEER` requests to the alpha contacts in the shortlist. Each contact, if it is live, should normally return k peers. If any of the alpha contacts fails to reply, it is removed from the shortlist.

The peer then fills the shortlist with contacts from the replies received. These should be the ones that are those closest to the target. From the shortlist it selects another alpha contacts filtering out the ones that have already been contacted. Once again a `FIND_PEER` is sent to each in parallel.

Each such parallel search updates `closestPeer`, the closest peer seen so far.

The sequence of parallel lookups is continued until either no peer in the sets returned is closer than the closest peer already seen or the initiating peer has accumulated k probed and known to be active contacts.

If a cycle doesn't find a closer peer, if closestPeer is unchanged, then the initiating peer sends a FIND_PEER to each of the k closest peers that it has not already queried.

At the end of this process, the peer will have accumulated a set of k active contacts.

### Refresh

K-buckets should be generally kept fresh by normal traffic. To cases where there are no lookups (in our case in low traffic networks), each peer refreshes all buckets to which it has not performed a peer lookup in the past `tRefresh`. Refreshing a bucket is done by creating a synthetic ID in the bucket's range and then performing a `FIND_PEER` for that ID.

## Literature

* [Kademlia: A Peer-to-peer Information System Based on the XOR Metric](https://pdos.csail.mit.edu/~petar/papers/maymounkov-kademlia-lncs.pdf)
* [Chord: A Scalable Peer-to-peer Lookup Protocol for Internet Applications](https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf)
* [Pastry: Scalable, distributed object location and routing for large-scale peer-to-peer systems](https://www.cs.rice.edu/~druschel/publications/Pastry.pdf)
* [Tapestry: A resilient global-scale overlay for service deployment](http://cs.brown.edu/courses/cs138/s12/doc/papers/tapestry_jsac03.pdf)
* [BEP-5: DHT Protocol](http://www.bittorrent.org/beps/bep_0005.html)
* [Kademlia Xlaticce Specification](http://xlattice.sourceforge.net/components/protocol/kademlia/specs.html)
* [A Common Protocol for Implementing Various DHT Algorithms](https://mice.cs.columbia.edu/getTechreport.php?techreportID=425&format=pdf&)
* [Coral: Sloppy hashing and self-organizing clusters](http://iptps03.cs.berkeley.edu/final-papers/coral.pdf)
* [Improving Lookup Performance over a Widely-Deployed DHT](https://www.cs.uoregon.edu/Reports/TR-2005-005.pdf)
* [ReDS: A Framework for Reputation-Enhanced DHTs](https://www.cs.indiana.edu/~kapadia/papers/reds-tpds-preprint-2013.pdf)
* [A Kademlia-based DHT for Resource Lookup in P2PSIP](https://tools.ietf.org/html/draft-cirani-p2psip-dsip-dhtkademlia-00)
