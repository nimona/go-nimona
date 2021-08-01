(window.webpackJsonp=window.webpackJsonp||[]).push([[25],{396:function(e,t,a){"use strict";a.r(t);var r=a(48),o=Object(r.a)({},(function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("ContentSlotsDistributor",{attrs:{"slot-key":e.$parent.slotKey}},[a("h1",{attrs:{id:"networking-relays"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#networking-relays"}},[e._v("#")]),e._v(" Networking - Relays")]),e._v(" "),a("h2",{attrs:{id:"problem-statement"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#problem-statement"}},[e._v("#")]),e._v(" Problem statement")]),e._v(" "),a("p",[e._v("Peers are not always able to listen for incoming network connections for a number of reasons. Local private networks without or with misconfigured UPNP, public or mobile networks, corporate or other governmental networks behind firewalls etc.")]),e._v(" "),a("p",[e._v("These peers can still participate in the network but to a limited extend. By using any TCP based transport layer they will be able to send and receive objects to/from peers they can make connections to. This also leaves them unable to talk to peers who share the same limitations.")]),e._v(" "),a("h2",{attrs:{id:"proposal-relays"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#proposal-relays"}},[e._v("#")]),e._v(" Proposal: Relays")]),e._v(" "),a("p",[e._v("Peers can chose to act as relays and advertise the fact on the network. Peers unable to accept incoming connections can use those peers to receive objects from others. Such peers will look for available relays and advertise them as their relays in their connection infos. Relays peers should be used after dialing all the peer's addresses has failed.")]),e._v(" "),a("p",[e._v("In order for relays not to be able to read the objects being given to them, the payload objects themselves will have to be encrypted. To remove the need for the peers to perform a handshake a shared key will be "),a("a",{attrs:{href:"https://blog.filippo.io/using-ed25519-keys-for-encryption",target:"_blank",rel:"noopener noreferrer"}},[e._v("derived"),a("OutboundLink")],1),e._v(" from the sender's and recipient's keys pairs allowing both peers to construct the encryption and decryption secret without any previous communication.")]),e._v(" "),a("p",[e._v("There are two main object types used for relaying messages.")]),e._v(" "),a("ul",[a("li",[a("code",[e._v("DataForwardRequest")]),e._v(" which is sent to relay peers")]),e._v(" "),a("li",[a("code",[e._v("DataForwardEnvelope")]),e._v(" which is sent by relay peers to recipients. The envelope contains an encrypted payload. It can also include a data forward request itself to enable chaining.")])]),e._v(" "),a("p",[e._v("The private key that is used to derive the shared secret can be an ephemeral key created only for that specific recipient or even object. For now we are using the sender's actual key.")]),e._v(" "),a("p",[e._v("Multiple relays can be chained together in a later version of the protocol that will enable an onion-like routing mechanism. This is already possible with the current version of the relay protocol but figuring out which relays to use and make sure it's secure will need more work.")]),e._v(" "),a("h3",{attrs:{id:"example"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#example"}},[e._v("#")]),e._v(" Example")]),e._v(" "),a("p",[e._v("Let's say that "),a("strong",[e._v("Alice")]),e._v(" (sender) wants to send a "),a("code",[e._v("Whatever")]),e._v(" object to "),a("strong",[e._v("Bob")]),e._v(" (recipient) through a relay, in this case "),a("strong",[e._v("Roger")]),e._v(":")]),e._v(" "),a("ol",[a("li",[a("strong",[e._v("Alice")]),e._v(" has to get the nested objects ready, and so...\n"),a("ul",[a("li",[e._v("Constructs a "),a("code",[e._v("Whatever")]),e._v(" object and encrypts it with "),a("code",[e._v("shared_key(Alice, Bob)")])]),e._v(" "),a("li",[e._v("Constructs a "),a("code",[e._v("DataForwardEnvelope")]),e._v(" "),a("ul",[a("li",[e._v("Sender: "),a("code",[e._v("Alice")])]),e._v(" "),a("li",[e._v("Payload: encrypted "),a("code",[e._v("Whatever")]),e._v(" from previous step")])])]),e._v(" "),a("li",[e._v("Constructs a "),a("code",[e._v("DataForwardRequest")]),e._v(" "),a("ul",[a("li",[e._v("Recipient: "),a("code",[e._v("Bob")])]),e._v(" "),a("li",[e._v("Payload: plaintext "),a("code",[e._v("DataForwardEnvelope")]),e._v(" from previous step")])])])])]),e._v(" "),a("li",[a("strong",[e._v("Alice")]),e._v(" sends the "),a("code",[e._v("DataForwardRequest")]),e._v(" object to "),a("strong",[e._v("Roger")])]),e._v(" "),a("li",[a("strong",[e._v("Roger")]),e._v(" receives the "),a("code",[e._v("DataForwardRequest")])]),e._v(" "),a("li",[a("strong",[e._v("Roger")]),e._v(" sends the the nested "),a("code",[e._v("DataForwardEnvelope")]),e._v(" object its recipient, "),a("strong",[e._v("Bob")])]),e._v(" "),a("li",[a("strong",[e._v("Bob")]),e._v(" receives the "),a("code",[e._v("DataForwardEnvelope")]),e._v(" and decrypts the "),a("code",[e._v("Whaterver")]),e._v(" object from its payload using the sender attribute to derive the shared key")]),e._v(" "),a("li",[a("strong",[e._v("Bob")]),e._v(" now has the decrypted "),a("code",[e._v("Whatever")]),e._v(" object")]),e._v(" "),a("li",[a("strong",[e._v("Roger")]),e._v(" sends back to "),a("strong",[e._v("Alice")]),e._v(" a "),a("code",[e._v("DataForwardResponse")]),e._v(" informing them of whether they were able to deliver the object")])]),e._v(" "),a("h3",{attrs:{id:"example-data-forward-request"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#example-data-forward-request"}},[e._v("#")]),e._v(" Example data forward request")]),e._v(" "),a("p",[a("img",{attrs:{src:"images/network-relays-object.svg",alt:"images/network-relays-object.svg"}})]),e._v(" "),a("h2",{attrs:{id:"messages"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#messages"}},[e._v("#")]),e._v(" Messages")]),e._v(" "),a("h3",{attrs:{id:"connection-info"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#connection-info"}},[e._v("#")]),e._v(" Connection Info")]),e._v(" "),a("div",{staticClass:"language-ndl extra-class"},[a("pre",{pre:!0,attrs:{class:"language-text"}},[a("code",[e._v("object nimona.io/peer.ConnectionInfo {\n    publicKey nimona.io/crypto.PublicKey\n    addresses repeated string\n    relays repeated nimona.io/peer.ConnectionInfo\n}\n")])])]),a("h3",{attrs:{id:"data-forward-request"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#data-forward-request"}},[e._v("#")]),e._v(" Data Forward Request")]),e._v(" "),a("div",{staticClass:"language-ndl extra-class"},[a("pre",{pre:!0,attrs:{class:"language-text"}},[a("code",[e._v("signed object nimona.io/network.DataForwardRequest {\n    requestID string\n    recipient nimona.io/crypto.PublicKey\n    envelope nimona.io/object.Object\n}\n")])])]),a("h3",{attrs:{id:"data-forward-envelope"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#data-forward-envelope"}},[e._v("#")]),e._v(" Data Forward Envelope")]),e._v(" "),a("div",{staticClass:"language-ndl extra-class"},[a("pre",{pre:!0,attrs:{class:"language-text"}},[a("code",[e._v("signed object nimona.io/network.DataForwardEnvelope {\n        requestID string\n    sender nimona.io/crypto.PublicKey\n    data data\n}\n")])])]),a("h3",{attrs:{id:"data-forward-response"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#data-forward-response"}},[e._v("#")]),e._v(" Data Forward Response")]),e._v(" "),a("div",{staticClass:"language-ndl extra-class"},[a("pre",{pre:!0,attrs:{class:"language-text"}},[a("code",[e._v("signed object nimona.io/network.DataForwardResponse {\n        requestID string\n        success bool\n        error string\n}\n")])])])])}),[],!1,null,null,null);t.default=o.exports}}]);