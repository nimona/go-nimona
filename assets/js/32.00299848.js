(window.webpackJsonp=window.webpackJsonp||[]).push([[32],{423:function(t,e,a){"use strict";a.r(e);var s=a(57),r=Object(s.a)({},(function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("ContentSlotsDistributor",{attrs:{"slot-key":t.$parent.slotKey}},[a("h1",{attrs:{id:"structured-objects"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#structured-objects"}},[t._v("#")]),t._v(" Structured Objects")]),t._v(" "),a("h2",{attrs:{id:"simple-summary"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#simple-summary"}},[t._v("#")]),t._v(" Simple Summary")]),t._v(" "),a("p",[t._v("Objects expand on the work on "),a("RouterLink",{attrs:{to:"/docs/proposals/np001-hinted-object-notation.html"}},[t._v("hinted object notation")]),t._v(" and further define\nthe top levels of their structure in order to add some commonly used attributes\nthat applications can leverage.")],1),t._v(" "),a("h2",{attrs:{id:"problem-statement"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#problem-statement"}},[t._v("#")]),t._v(" Problem Statement")]),t._v(" "),a("p",[t._v("In order for application and service developers to be able to identify, use, and\ncreate data structures compatible with other applications we need to define a\nbasic list of well known attributed, some required some optional.")]),t._v(" "),a("h2",{attrs:{id:"proposal"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#proposal"}},[t._v("#")]),t._v(" Proposal")]),t._v(" "),a("p",[t._v("The top level of each object consists of three main attributes.")]),t._v(" "),a("ul",[a("li",[a("code",[t._v("@type:s")]),t._v(" is an arbitrary string defining the type of the object's content.")]),t._v(" "),a("li",[a("code",[t._v("@metadata:m")]),t._v(" are a fixed set of attributes that add extra info to the object.")])]),t._v(" "),a("div",{staticClass:"language-json extra-class"},[a("pre",{pre:!0,attrs:{class:"language-json"}},[a("code",[a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n  "),a("span",{pre:!0,attrs:{class:"token property"}},[t._v('"@type:s"')]),a("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token string"}},[t._v('"type"')]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n  "),a("span",{pre:!0,attrs:{class:"token property"}},[t._v('"@metadata:m"')]),a("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n    "),a("span",{pre:!0,attrs:{class:"token property"}},[t._v('"root:r"')]),a("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token string"}},[t._v('"bah..."')]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),a("span",{pre:!0,attrs:{class:"token property"}},[t._v('"owner:s"')]),a("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token string"}},[t._v('"did:x:y"')]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),a("span",{pre:!0,attrs:{class:"token property"}},[t._v('"parents:m"')]),a("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("[")]),t._v("\n      "),a("span",{pre:!0,attrs:{class:"token property"}},[t._v('"*:as"')]),a("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("[")]),a("span",{pre:!0,attrs:{class:"token string"}},[t._v('"bah..."')]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("]")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n      "),a("span",{pre:!0,attrs:{class:"token property"}},[t._v('"some-type:as"')]),a("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("[")]),a("span",{pre:!0,attrs:{class:"token string"}},[t._v('"bah..."')]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("]")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("]")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),a("span",{pre:!0,attrs:{class:"token property"}},[t._v('"_signature:m"')]),a("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n      "),a("span",{pre:!0,attrs:{class:"token property"}},[t._v('"alg:s"')]),a("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token string"}},[t._v('"hashing-algorithm"')]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n      "),a("span",{pre:!0,attrs:{class:"token property"}},[t._v('"signer:s"')]),a("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token string"}},[t._v('"bah..."')]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n      "),a("span",{pre:!0,attrs:{class:"token property"}},[t._v('"x:d"')]),a("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token string"}},[t._v('"bah..."')]),t._v("\n    "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n  "),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n  "),a("span",{pre:!0,attrs:{class:"token property"}},[t._v('"foo:s"')]),a("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),a("span",{pre:!0,attrs:{class:"token string"}},[t._v('"bar"')]),t._v("\n"),a("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n")])])]),a("h3",{attrs:{id:"type"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#type"}},[t._v("#")]),t._v(" Type")]),t._v(" "),a("div",{staticClass:"custom-block danger"},[a("p",{staticClass:"custom-block-title"},[t._v("WARNING")]),t._v(" "),a("p",[t._v("Types are currently a way of moving forward, it's quite possible they will be\ndeprecated in the future in favor once schemes are introduced.")])]),t._v(" "),a("h3",{attrs:{id:"well-known-types"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#well-known-types"}},[t._v("#")]),t._v(" Well known types")]),t._v(" "),a("ul",[a("li",[a("code",[t._v("nimona.io/crypto.PublicKey")])]),t._v(" "),a("li",[a("code",[t._v("nimona.io/crypto.PrivateKey")])]),t._v(" "),a("li",[a("code",[t._v("nimona.io/object.CID")])]),t._v(" "),a("li",[a("code",[t._v("nimona.io/peer.ConnectionInfoInfo")])]),t._v(" "),a("li",[a("code",[t._v("nimona.io/peer.ConnectionInfoRequest")])]),t._v(" "),a("li",[a("code",[t._v("nimona.io/peer.ConnectionInfoResponse")])]),t._v(" "),a("li",[a("code",[t._v("nimona.io/object.Certificate")])]),t._v(" "),a("li",[a("code",[t._v("nimona.io/object.CertificateRequest")])]),t._v(" "),a("li",[a("code",[t._v("nimona.io/exchange.ObjectRequest")])]),t._v(" "),a("li",[a("code",[t._v("nimona.io/exchange.ObjectResponse")])]),t._v(" "),a("li",[t._v("...")])]),t._v(" "),a("h3",{attrs:{id:"metadata"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#metadata"}},[t._v("#")]),t._v(" Metadata")]),t._v(" "),a("ul",[a("li",[a("code",[t._v("owner:s")]),t._v(" (optional) Public keys of the owner of the object.")]),t._v(" "),a("li",[a("code",[t._v("root:r")]),t._v(" (optional) Root hash of the stream the object is part of.")]),t._v(" "),a("li",[a("code",[t._v("parents:as")]),t._v(" (optional) Array of cids of parent objects, this is used\nfor streams")]),t._v(" "),a("li",[a("code",[t._v("_signature:m")]),t._v(" (optional) Cryptographic signature by the owner.")])]),t._v(" "),a("p",[t._v("Additional metadata will be added in regards to access control and schema\nspecification.")]),t._v(" "),a("h2",{attrs:{id:"references"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#references"}},[t._v("#")]),t._v(" References")]),t._v(" "),a("ul",[a("li",[a("a",{attrs:{href:"https://www.json.org",target:"_blank",rel:"noopener noreferrer"}},[t._v("JSON"),a("OutboundLink")],1)]),t._v(" "),a("li",[a("a",{attrs:{href:"http://cbor.io",target:"_blank",rel:"noopener noreferrer"}},[t._v("CBOR"),a("OutboundLink")],1)]),t._v(" "),a("li",[a("a",{attrs:{href:"https://msgpack.org",target:"_blank",rel:"noopener noreferrer"}},[t._v("MsgPack"),a("OutboundLink")],1)]),t._v(" "),a("li",[a("a",{attrs:{href:"https://capnproto.org",target:"_blank",rel:"noopener noreferrer"}},[t._v("Cap-n-proto"),a("OutboundLink")],1)]),t._v(" "),a("li",[a("a",{attrs:{href:"https://json-ld.org",target:"_blank",rel:"noopener noreferrer"}},[t._v("JSON-LD"),a("OutboundLink")],1)]),t._v(" "),a("li",[a("a",{attrs:{href:"https://json-schema.org",target:"_blank",rel:"noopener noreferrer"}},[t._v("JSON-Schema"),a("OutboundLink")],1)]),t._v(" "),a("li",[a("a",{attrs:{href:"https://tjson.org",target:"_blank",rel:"noopener noreferrer"}},[t._v("Tagged JSON"),a("OutboundLink")],1)]),t._v(" "),a("li",[a("a",{attrs:{href:"https://github.com/benlaurie",target:"_blank",rel:"noopener noreferrer"}},[t._v("Ben Laurie"),a("OutboundLink")],1)]),t._v(" "),a("li",[a("a",{attrs:{href:"https://github.com/benlaurie/objecthash",target:"_blank",rel:"noopener noreferrer"}},[t._v("Object hash"),a("OutboundLink")],1)])])])}),[],!1,null,null,null);e.default=r.exports}}]);