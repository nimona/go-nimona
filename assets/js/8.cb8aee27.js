(window.webpackJsonp=window.webpackJsonp||[]).push([[8],{362:function(t,e,a){t.exports=a.p+"assets/img/np003-streams.drawio.b6b06b4a.svg"},396:function(t,e,a){"use strict";a.r(e);var s=a(48),n=Object(s.a)({},(function(){var t=this,e=t.$createElement,s=t._self._c||e;return s("ContentSlotsDistributor",{attrs:{"slot-key":t.$parent.slotKey}},[s("h1",{attrs:{id:"streams"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#streams"}},[t._v("#")]),t._v(" Streams")]),t._v(" "),s("h2",{attrs:{id:"simple-summary"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#simple-summary"}},[t._v("#")]),t._v(" Simple Summary")]),t._v(" "),s("p",[t._v("Streams provide a way of creating complex mutable data structures using\ndirected acyclic graphs made from objects.")]),t._v(" "),s("h2",{attrs:{id:"problem-statement"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#problem-statement"}},[t._v("#")]),t._v(" Problem Statement")]),t._v(" "),s("p",[t._v("While objects on their own are useful for creating permanent content-addressable\ndata structures, there are very few applications where data never get updated.\nThis is where streams come in, they allow developers to create complex\napplications by applying event driven and event sourcing patterns using graphs\nof individually immutable objects.")]),t._v(" "),s("h2",{attrs:{id:"proposal"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#proposal"}},[t._v("#")]),t._v(" Proposal")]),t._v(" "),s("p",[t._v("Objects in a stream form a directed acyclic graph (DAG) by allowing each of the\nobjects to reference others it depends on or knows of. This graph can then be\nserialized into a linear series of objects that can be replayed consistently by\neveryone that has the same representation of the graph.")]),t._v(" "),s("p",[t._v("Streams are identified by the cid of their root object. This means that even\nthough each of their objects is content-addressable; the stream as a whole is\nnot, as its root cid (and thus identifier) does not change when more objects\nare added to the graph.")]),t._v(" "),s("p",[t._v("The benefit of this is that there is no need to find a way to reference the\nstream as it changes. The downside is that you do not really know if you have\nactually  received the whole stream and whether peers are not holding back on\nyou.")]),t._v(" "),s("p",[s("img",{attrs:{src:a(362),alt:"stream"}})]),t._v(" "),s("h2",{attrs:{id:"structures"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#structures"}},[t._v("#")]),t._v(" Structures")]),t._v(" "),s("h3",{attrs:{id:"root-object"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#root-object"}},[t._v("#")]),t._v(" Root Object")]),t._v(" "),s("div",{staticClass:"language-json extra-class"},[s("pre",{pre:!0,attrs:{class:"language-json"}},[s("code",[s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"type:s"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"stream:nimona.io/kv"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"data:m"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n")])])]),s("h3",{attrs:{id:"child-object"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#child-object"}},[t._v("#")]),t._v(" Child Object")]),t._v(" "),s("div",{staticClass:"language-json extra-class"},[s("pre",{pre:!0,attrs:{class:"language-json"}},[s("code",[s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"type:s"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"<any-type>"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"metadata:m"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"stream:s"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"<stream-root-cid>"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"parents:m"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n      "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"*:as"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("[")]),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"<last-known-leaf-cids>"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("]")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"data:m"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n")])])]),s("h2",{attrs:{id:"access-control-policy"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#access-control-policy"}},[t._v("#")]),t._v(" Access control policy")]),t._v(" "),s("p",[s("em",[t._v("Note: Work in progress.")])]),t._v(" "),s("p",[t._v("Policy parameters:")]),t._v(" "),s("ul",[s("li",[s("code",[t._v("type")]),t._v(" required. ["),s("code",[t._v("signature")]),t._v("].")]),t._v(" "),s("li",[s("code",[t._v("subjects")]),t._v(" optional (public key).")]),t._v(" "),s("li",[s("code",[t._v("actions")]),t._v(" optional ["),s("code",[t._v("read")]),t._v(", "),s("code",[t._v("append")]),t._v("].")]),t._v(" "),s("li",[s("code",[t._v("resources")]),t._v(" optional (only used for stream policies).")]),t._v(" "),s("li",[s("code",[t._v("effect")]),t._v(" required ["),s("code",[t._v("allow")]),t._v(", "),s("code",[t._v("deny")]),t._v("].")])]),t._v(" "),s("h3",{attrs:{id:"evaluation"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#evaluation"}},[t._v("#")]),t._v(" Evaluation")]),t._v(" "),s("p",[t._v("A request is a combination of a "),s("code",[t._v("subject")]),t._v(", "),s("code",[t._v("resource")]),t._v(", and "),s("code",[t._v("action")]),t._v(".")]),t._v(" "),s("p",[t._v("A request "),s("em",[t._v("matches")]),t._v(" a policy if each of the policy's parameters ("),s("code",[t._v("subjects")]),t._v(",\n"),s("code",[t._v("resources")]),t._v(", "),s("code",[t._v("actions")]),t._v(") are either empty or explicitly match the equivalent\nparameter of the request being evaluated.")]),t._v(" "),s("p",[t._v("In order to check whether the given request is allowed or not we evaluate each\npolicy against the request using the following rules.")]),t._v(" "),s("ol",[s("li",[t._v("If there are no policies the action is allowed.")])]),t._v(" "),s("p",[t._v("Else, For each policy:")]),t._v(" "),s("ol",[s("li",[t._v("If the request doesn't match the policy we move on to the next policy.")]),t._v(" "),s("li",[t._v("We count how many parameters were "),s("em",[t._v("explicitly")]),t._v(" matched and if the previous\nmatch had the same or more parameters matched, we consider the policy's\neffect as the latest evaluation result.")]),t._v(" "),s("li",[t._v("When all policies have been evaluated the latest result is used.")])]),t._v(" "),s("h3",{attrs:{id:"policy-example-table-1"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#policy-example-table-1"}},[t._v("#")]),t._v(" Policy example table 1")]),t._v(" "),s("table",[s("thead",[s("tr",[s("th",[t._v("NAME")]),t._v(" "),s("th",[t._v("SUBJECT")]),t._v(" "),s("th",[t._v("RESOURCE")]),t._v(" "),s("th",[t._v("ACTION")]),t._v(" "),s("th",[t._v("S0, R0, A1")]),t._v(" "),s("th",[t._v("S0, R0, A2")]),t._v(" "),s("th",[t._v("S1, R0, A1")]),t._v(" "),s("th",[t._v("S1, R0, A2")])])]),t._v(" "),s("tbody",[s("tr",[s("td",[t._v("0 - allow s* r* a*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("allow")])]),t._v(" "),s("tr",[s("td",[t._v("1 - deny s* r* a1")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("a1")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("allow")])]),t._v(" "),s("tr",[s("td",[t._v("2 - deny s0 r* a*")]),t._v(" "),s("td",[t._v("s0")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("allow")])]),t._v(" "),s("tr",[s("td",[t._v("3 - allow s0 r* a2")]),t._v(" "),s("td",[t._v("s0")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("a2")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("allow")])]),t._v(" "),s("tr",[s("td",[t._v("4 - allow s0 r0 a1")]),t._v(" "),s("td",[t._v("s0")]),t._v(" "),s("td",[t._v("r0")]),t._v(" "),s("td",[t._v("a1")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("allow")])]),t._v(" "),s("tr",[s("td",[t._v("5 - deny s* r0 a2")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("r0")]),t._v(" "),s("td",[t._v("a2")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("deny")])]),t._v(" "),s("tr",[s("td",[t._v("6 - deny s* r* a*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("deny")])])])]),t._v(" "),s("h3",{attrs:{id:"policy-example-table-2"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#policy-example-table-2"}},[t._v("#")]),t._v(" Policy example table 2")]),t._v(" "),s("table",[s("thead",[s("tr",[s("th",[t._v("NAME")]),t._v(" "),s("th",[t._v("SUBJECT")]),t._v(" "),s("th",[t._v("RESOURCE")]),t._v(" "),s("th",[t._v("ACTION")]),t._v(" "),s("th",[t._v("S0, R0, A1")]),t._v(" "),s("th",[t._v("S0, R0, A2")]),t._v(" "),s("th",[t._v("S1, R0, A1")]),t._v(" "),s("th",[t._v("S1, R0, A2")])])]),t._v(" "),s("tbody",[s("tr",[s("td",[t._v("0 - allow s1 r0 a1")]),t._v(" "),s("td",[t._v("s1")]),t._v(" "),s("td",[t._v("r0")]),t._v(" "),s("td",[t._v("a1")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("allow")])]),t._v(" "),s("tr",[s("td",[t._v("1 - deny s* r* a*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("deny")])]),t._v(" "),s("tr",[s("td",[t._v("2 - allow s* r* a*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("allow")])]),t._v(" "),s("tr",[s("td",[t._v("3 - deny s0 r* a*")]),t._v(" "),s("td",[t._v("s0")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("allow")])]),t._v(" "),s("tr",[s("td",[t._v("4 - deny s* r0 a*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("r0")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("deny")])]),t._v(" "),s("tr",[s("td",[t._v("5 - deny s* r0 a2")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("r0")]),t._v(" "),s("td",[t._v("a2")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("deny")])]),t._v(" "),s("tr",[s("td",[t._v("6 - allow s* r* a2")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("a2")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("deny")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("deny")])]),t._v(" "),s("tr",[s("td",[t._v("7 - allow s0 r0 a*")]),t._v(" "),s("td",[t._v("s0")]),t._v(" "),s("td",[t._v("r0")]),t._v(" "),s("td",[t._v("*")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("allow")]),t._v(" "),s("td",[t._v("deny")])])])]),t._v(" "),s("h3",{attrs:{id:"example"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#example"}},[t._v("#")]),t._v(" Example")]),t._v(" "),s("div",{staticClass:"language-json extra-class"},[s("pre",{pre:!0,attrs:{class:"language-json"}},[s("code",[s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"type:s"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"stream:conversation"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"owner:s"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"bah0"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"policies:am"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("[")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token comment"}},[t._v("// allow everyone to read the whole stream")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"type:s"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"signature"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"actions:as"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("[")]),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"read"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("]")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"effect:s"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"allow"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token comment"}},[t._v("// allow a couple of keys to send messages")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"type:s"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"signature"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"subjects:as"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("[")]),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"bah1"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"bah2"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"bah3"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("]")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"actions:as"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("[")]),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"append"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("]")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"resources:as"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("[")]),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"conversation.MessageAdded"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("]")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"effect:s"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"allow"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token comment"}},[t._v("// allow owner to modify conversation topic")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"type:s"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"signature"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"subjects:as"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("[")]),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"bah0"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("]")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"actions:as"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("[")]),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"append"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("]")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"resources:as"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("[")]),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"conversation.TopicUpdated"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("]")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n    "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"effect:s"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"allow"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("]")]),t._v("\n"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n")])])]),s("h2",{attrs:{id:"hypothetical-roots"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#hypothetical-roots"}},[t._v("#")]),t._v(" Hypothetical roots")]),t._v(" "),s("p",[t._v('As mentioned before, streams are identified by the CID of their root object. In\norder for a peer to find the providers of a stream and get its objects, it must\nat the very least know its identifier. This is usually not an issue as most\ntimes a peer will learn about the existence of a stream from somewhere before\ndeciding to request it. There are some cases though where that might not be the\ncase, especially when looking for something that might be considered relatively\n"well known".')]),t._v(" "),s("p",[t._v("An example of this would be the profile stream of an identity. Let's say we are\nlooking at a blog post that a single author. Unless the blog post somehow\ncontains a link to the author's profile stream, there is no other way to easily\nfind the stream's identifier.")]),t._v(" "),s("p",[t._v("This is where hypothetical roots come in.")]),t._v(" "),s("p",[t._v("A hypothetical root is an object that identifies a stream and can be assumed\nexists given the type of stream and the author that would have created it. This\nallows peers to find streams unique to an identity without having to somehow\nlearn of their existence.")]),t._v(" "),s("p",[t._v("Since the hypothetical root does not contain a policy, the stream starts off as\npublicly accessible but writable only by the author. The author can subsequently\ndecide to restrict the rest of the stream by using a more strict policy.")]),t._v(" "),s("hr"),t._v(" "),s("p",[t._v("Let's go back to our original example of profile streams.")]),t._v(" "),s("p",[t._v("Assuming that peer "),s("code",[t._v("a11")]),t._v(" wants the profile stream for the identity "),s("code",[t._v("f00")]),t._v(", all it\nhas to do is construct the hypothetical root, calculate its CID, and find\nproviders for it on the network.")]),t._v(" "),s("div",{staticClass:"language-json extra-class"},[s("pre",{pre:!0,attrs:{class:"language-json"}},[s("code",[s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("{")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"type:s"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"nimona.io/profile.Created"')]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v(",")]),t._v("\n  "),s("span",{pre:!0,attrs:{class:"token property"}},[t._v('"author:s"')]),s("span",{pre:!0,attrs:{class:"token operator"}},[t._v(":")]),t._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[t._v('"f00"')]),t._v("\n"),s("span",{pre:!0,attrs:{class:"token punctuation"}},[t._v("}")]),t._v("\n")])])]),s("p",[t._v("The CID of this object is "),s("code",[t._v("oh1.9KQhQ4UGaQPEyUDAAPDmVJCoHnGtJY7Aun4coFATXCYK")]),t._v("\nand the peer can now lookup the providers for this object, and sync the\nremaining stream.")]),t._v(" "),s("hr"),t._v(" "),s("p",[t._v("The NDL for defining hypothetical roots is as follows. Additional objects can be\ndefined in the stream as needed, but the hypothetical root object itself cannot\nhave additional properties.")]),t._v(" "),s("div",{staticClass:"language-ndl extra-class"},[s("pre",{pre:!0,attrs:{class:"language-text"}},[s("code",[t._v("stream nimona.io/profile {\n    hypothetical root object Created { }\n    signed object NameUpdated {\n        nameFirst string\n        nameLast string\n        dependsOn repeated relationship\n    }\n}\n")])])]),s("h2",{attrs:{id:"synchronization"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#synchronization"}},[t._v("#")]),t._v(" Synchronization")]),t._v(" "),s("p",[s("em",[t._v("Note: Work in progress.")])]),t._v(" "),s("div",{staticClass:"language-ndl extra-class"},[s("pre",{pre:!0,attrs:{class:"language-text"}},[s("code",[t._v("    signed object nimona.io/stream.StreamRequest {\n        nonce string\n        leaves repeated nimona.io/object.CID\n    }\n")])])]),s("div",{staticClass:"language-ndl extra-class"},[s("pre",{pre:!0,attrs:{class:"language-text"}},[s("code",[t._v("    signed object nimona.io/stream.StreamResponse {\n        nonce string\n        children repeated nimona.io/object.CID\n    }\n")])])]),s("div",{staticClass:"language-ndl extra-class"},[s("pre",{pre:!0,attrs:{class:"language-text"}},[s("code",[t._v("    signed object nimona.io/stream.Announcement {\n        nonce string\n        leaves repeated nimona.io/object.CID\n    }\n")])])]),s("h2",{attrs:{id:"subscriptions"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#subscriptions"}},[t._v("#")]),t._v(" Subscriptions")]),t._v(" "),s("p",[t._v('Peers can "subscribe" to stream updates by creating and sending subscriptions to\nother peers. A subscription can be used to subscribe on updates to one or more\nstreams using the streams\' root CID and must also specify an expiration time\nfor the subscription.')]),t._v(" "),s("p",[t._v("When a peer receives or creates an update for a stream, they will go through the\nsubscriptions they have received, and notify the relevant peers about the new\nupdates. If the subscriber does not have have access to the stream, no\nnotification will be sent.")]),t._v(" "),s("div",{staticClass:"language-ndl extra-class"},[s("pre",{pre:!0,attrs:{class:"language-text"}},[s("code",[t._v("signed object nimona.io/stream.Subscription {\n    rootCIDs nimona.io/object.CID\n    expiry nimona.io/object.DateTime\n}\n")])])]),s("p",[t._v("Subscriptions can also be added as stream events. This allows identities and\npeers that have write access to a stream to denote their interest in receiving\nupdates about that stream. In this case "),s("code",[t._v("rootCIDs")]),t._v(" should be empty and the\nexpiry is optional.")]),t._v(" "),s("h2",{attrs:{id:"references"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#references"}},[t._v("#")]),t._v(" References")]),t._v(" "),s("ul",[s("li",[s("a",{attrs:{href:"https://docs.textile.io/threads/#threads",target:"_blank",rel:"noopener noreferrer"}},[t._v("https://docs.textile.io/threads/#threads"),s("OutboundLink")],1)]),t._v(" "),s("li",[s("a",{attrs:{href:"https://www.streamr.com/docs/streams",target:"_blank",rel:"noopener noreferrer"}},[t._v("https://www.streamr.com/docs/streams"),s("OutboundLink")],1)]),t._v(" "),s("li",[s("a",{attrs:{href:"https://holochain.org",target:"_blank",rel:"noopener noreferrer"}},[t._v("https://holochain.org"),s("OutboundLink")],1)]),t._v(" "),s("li",[s("a",{attrs:{href:"https://github.com/textileio/go-textile/issues/694",target:"_blank",rel:"noopener noreferrer"}},[t._v("https://github.com/textileio/go-textile/issues/694"),s("OutboundLink")],1)]),t._v(" "),s("li",[s("a",{attrs:{href:"https://tuhrig.de/messages-vs-events-vs-commands",target:"_blank",rel:"noopener noreferrer"}},[t._v("https://tuhrig.de/messages-vs-events-vs-commands"),s("OutboundLink")],1)]),t._v(" "),s("li",[s("a",{attrs:{href:"https://www.swirlds.com/downloads/SWIRLDS-TR-2016-01.pdf",target:"_blank",rel:"noopener noreferrer"}},[t._v("https://www.swirlds.com/downloads/SWIRLDS-TR-2016-01.pdf"),s("OutboundLink")],1)]),t._v(" "),s("li",[s("a",{attrs:{href:"https://arxiv.org/pdf/1710.04469.pdf",target:"_blank",rel:"noopener noreferrer"}},[t._v("https://arxiv.org/pdf/1710.04469.pdf"),s("OutboundLink")],1)]),t._v(" "),s("li",[s("a",{attrs:{href:"http://archagon.net/blog/2018/03/24/data-laced-with-history",target:"_blank",rel:"noopener noreferrer"}},[t._v("http://archagon.net/blog/2018/03/24/data-laced-with-history"),s("OutboundLink")],1)])])])}),[],!1,null,null,null);e.default=n.exports}}]);