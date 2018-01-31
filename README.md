[![CircleCI](https://circleci.com/gh/nimona/go-nimona-fabric.svg?style=svg)](https://circleci.com/gh/nimona/go-nimona-fabric)
[![Maintainability](https://api.codeclimate.com/v1/badges/96aaae63697e543b9bc9/maintainability)](https://codeclimate.com/github/nimona/go-nimona-fabric/maintainability)

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
