(window.webpackJsonp=window.webpackJsonp||[]).push([[36],{444:function(e,n,t){"use strict";t.r(n);var s=t(59),a=Object(s.a)({},(function(){var e=this.$createElement,n=this._self._c||e;return n("ContentSlotsDistributor",{attrs:{"slot-key":this.$parent.slotKey}},[n("h1",{attrs:{id:"nimona-idl-codegen"}},[n("a",{staticClass:"header-anchor",attrs:{href:"#nimona-idl-codegen"}},[this._v("#")]),this._v(" Nimona IDL Codegen")]),this._v(" "),n("div",{staticClass:"language-ndl extra-class"},[n("pre",{pre:!0,attrs:{class:"language-text"}},[n("code",[this._v("package conversation\n\nimport nimona.io/crypto crypto\n\nstream mochi.io/conversation {\n    signed root object Created {\n        name string\n    }\n    signed object TopicUpdated {\n        topic string\n        dependsOn repeated relationship\n    }\n    signed object MessageAdded {\n        body string\n        dependsOn repeated relationship\n    }\n    signed object MessageRemoved {\n        removes relationship\n        dependsOn repeated relationship\n    }\n}\n")])])])])}),[],!1,null,null,null);n.default=a.exports}}]);