# Networking

Nimona is at its core a peer to peer network where each peer represents an
application, service, bot, etc that can send and receive objects.

## Establishing connections

Currently the only communication protocol supported is TCP with mTLS.
In order for a peer to be able to connect to another, they need to first know
an IP address, and their public key.

A number of bootstrap peers will have to be provided to new peers so they are
able to start discovering others.
