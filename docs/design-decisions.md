# Design Decisions

- Full ed25519 public keys for peer identifiers (instead of fingerprints)
- Peer's public key must be known in order to be able to connect or exchange messages
  - Mutual TLS used for establishing TCP connections
  - Derived shared secret used for encrypting relayed messages
- Transport agnostic object based communication
- Encoding agnostic objects
  - Object keys include a hint that defines the expected type of their value
  - Object attributes have a limited set of types (int, float, string, bytes, bool, map, array)
    - No null type (not final)
- Encoding agnostic object hashes
  - Null values are ignored
  - Empty arrays and maps are ignored
  - Maps keys are sorted
  - Floats are normalized as IEEE float (not final)
  - Integers are normalized as strings
- Discovery is delegated to a sub-set of the network's peers
  