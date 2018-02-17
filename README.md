[![CircleCI](https://img.shields.io/circleci/project/github/nimona/go-nimona-fabric.svg?style=flat-square)](https://circleci.com/gh/nimona/go-nimona-fabric)
[![Coveralls github](https://img.shields.io/coveralls/github/nimona/go-nimona-fabric.svg?style=flat-square)](https://coveralls.io/github/nimona/go-nimona-fabric)
[![license](https://img.shields.io/github/license/nimona/go-nimona-fabric.svg?style=flat-square)](https://github.com/nimona/go-nimona-fabric/blob/master/LICENSE)
[![GitHub issues](https://img.shields.io/github/issues/nimona/go-nimona-fabric.svg?style=flat-square)](https://github.com/nimona/go-nimona-fabric/issues)
[![Waffle.io](https://img.shields.io/waffle/label/nimona/go-nimona-fabric/in%20progress.svg?style=flat-square)](https://waffle.io/nimona/go-nimona-fabric)

# Nimona Fabric

Fabric is an implementation of Nimona's network stack that provides some very
opinionated features targeting mainly peer to peer and decentralized systems.  

- Verbose network address notations that expose tranports, protocols, etc
- Protocols multiplexing and negotiation over the same transport layer
- Optional peer and service discovery
- Optional routing connections through proxy peers

That being said, there is nothing prohibiting its use in other applications, 
eg microservices.  

For a rational as well as more information you might want to check out the
[design document](https://github.com/nimona/nimona/blob/master/network.md).
