(window.webpackJsonp=window.webpackJsonp||[]).push([[23],{396:function(e,t,s){"use strict";s.r(t);var i=s(48),n=Object(i.a)({},(function(){var e=this,t=e.$createElement,s=e._self._c||t;return s("ContentSlotsDistributor",{attrs:{"slot-key":e.$parent.slotKey}},[s("h1",{attrs:{id:"design-decisions"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#design-decisions"}},[e._v("#")]),e._v(" Design Decisions")]),e._v(" "),s("ul",[s("li",[e._v("Full ed25519 public keys for peer identifiers (instead of fingerprints)")]),e._v(" "),s("li",[e._v("Peer's public key must be known in order to be able to connect or exchange messages\n"),s("ul",[s("li",[e._v("Mutual TLS used for establishing TCP connections")]),e._v(" "),s("li",[e._v("Derived shared secret used for encrypting relayed messages")])])]),e._v(" "),s("li",[e._v("Transport agnostic object based communication")]),e._v(" "),s("li",[e._v("Encoding agnostic objects\n"),s("ul",[s("li",[e._v("Object keys include a hint that defines the expected type of their value")]),e._v(" "),s("li",[e._v("Object attributes have a limited set of types (int, float, string, bytes, bool, map, array)\n"),s("ul",[s("li",[e._v("No null type (not final)")])])])])]),e._v(" "),s("li",[e._v("Encoding agnostic object hashes\n"),s("ul",[s("li",[e._v("Null values are ignored")]),e._v(" "),s("li",[e._v("Empty arrays and maps are ignored")]),e._v(" "),s("li",[e._v("Maps keys are sorted")]),e._v(" "),s("li",[e._v("Floats are normalized as IEEE float (not final)")]),e._v(" "),s("li",[e._v("Integers are normalized as strings")])])]),e._v(" "),s("li",[e._v("Discovery is delegated to a sub-set of the network's peers")])])])}),[],!1,null,null,null);t.default=n.exports}}]);