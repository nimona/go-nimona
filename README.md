# Nimona Fabric

Fabric is an implementation of Nimona's network stack that provides some very
opinionated features targeting mainly peer to peer and decentralized systems.  

- Verbose network address notations that expose tranports, protocols, middleware, etc
- Protocols multiplexing and negotiation over the same transport layer
- Optional peer and service discovery
- Optional routing connections through proxy peers

That being said, there is nothing prohibiting its use in other applications, 
eg microservices.  

For a rational as well as more information you might want to check out the
[design document](https://github.com/nimona/nimona/blob/master/fabric.md).
