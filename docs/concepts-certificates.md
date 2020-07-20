# Certificates

_Note: Work in progress._

Certificates enable users to authorize applications to act on their behalf
in the network while limiting the data they can interact with and the actions
they can perform.

A certificate is a signed object of type `nimona.io/peer.Certificate` and
consists of the following required attributes:

* `subject` (string): The public part of the application's key pair.
* `resources` (repeated string): Resources the application has been permitted
  to access. Can either be an exact match for an object type, or a glob.
* `actions` (repeated string): The type of action permitted for each resource.
  This is a one-to-one matching with the resources mentioned above. The action
  at index 0 is related to the resources at index 0. Actions currently defined
  are only `read` and `create`.
* `created` (string, ISO-8601): The timestamp the certificate was created.
* `expires` (string, ISO-8601): The timestamp the certificate expires.

An application must include a certificate in every object it sends on behalf of
a user.

If a peer receives an object with a certificate that has either expired, or
does not grant the requester the permissions it needs to perform an action,
the remote party simply ignores the object or request.

```json
{
  "type:s": "nimona.io/peer.Certificate",
  "data:m": {
    "subject:s": "ed25519.2h8Qu2TJCpnwv7jUaQLpazsxMW4iCaTAFgxoi5crsEAs",
    "resources:as": [
      "nimona.io/profile.Profile",
      "nimona.io/profile.ProfileRequest",
      "mochi.io/conversation.*"
    ],
    "actions:as": [
      "create",
      "create",
      "read"
    ],
    "created:s": "2020-06-25T19:16:43Z",
    "expires:s": "2021-06-25T19:16:43Z",
  },
  "_signatures:am": [{
      "signer:o": "ed25519.8mE4CeLLCwpyfqyNFkT6gV32ZYcYP6jt1yzMDmzbxxRL",
      "alg:o": "OH_ES256",
      "x:d": "x0..."
  }]
}
```

## Certificate Requests

Applications need to request a certification from the user.
To do that they create, sign, and give to the user a certificate request.

The user needs to load the certificate request using an identity application
that manages their identity keys, verify that they are happy with the
permissions the application is asking for, create a certificate and sign it.

The certificate request can optionally include a request for a profile from
the user.

```json
{
  "type:s": "nimona.io/peer.CertificateRequest",
  "data:m": {
    "applicationName:s": "Foobar",
    "applicationDescription:s": "An app that does nothing",
    "applicationURL:s": "https://github.com/nimona",
    "applicationIcon:d": "x0...",
    "applicationBanner:d": "x0...",
    "requestProfile:b": true,
    "certificateSubject:s": "ed25519.2h8Qu2TJCpnwv7jUaQLpazsxMW4iCaTAFgxoi5crsEAs",
    "certificateResources:as": [
      "nimona.io/profile.Profile",
      "nimona.io/profile.ProfileRequest",
      "mochi.io/conversation.*"
    ],
    "certificateActions:as": [
      "create",
      "create",
      "read"
    ]
  }
}
```

The certificate request can be currently provided in two ways:

a. Through the use of a QR code that the user must scan using their identity
application.
This is mainly used for when the application is running on a device where the
user does not have an identity application installed.

Once the certificate has been created and signed, the identity application
will lookup the certificate request's signer in the network and directly
send them the certificate.

b. Through a link with a custom URL using the `nimona://` scheme and the
`certificate-request` host.
The certificate request itself must be provided via a query parameter named
`certificateRequest`.

An optional query param `returnPath` can be set in order to define how the
identity application is expected to return the certificate back to the
requester.

If not set, the identity app will lookup the requester on the network and send
them the certificate.

If the `returnPath` is set to a URL with either HTTPS or custom scheme, the
identity application will append `&certificate=xxx` to the return path and
redirect the user to it.

## Revocation

Currently certificate revocation is not supported, we are working hard on that.
