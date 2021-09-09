(window.webpackJsonp=window.webpackJsonp||[]).push([[34],{410:function(a,s,t){"use strict";t.r(s);var e=t(48),r=Object(e.a)({},(function(){var a=this,s=a.$createElement,t=a._self._c||s;return t("ContentSlotsDistributor",{attrs:{"slot-key":a.$parent.slotKey}},[t("h1",{attrs:{id:"chat"}},[t("a",{staticClass:"header-anchor",attrs:{href:"#chat"}},[a._v("#")]),a._v(" Chat")]),a._v(" "),t("p",[a._v("Chat is a proof of concept app that allows peers to join public conversations\nand send messages.")]),a._v(" "),t("h2",{attrs:{id:"env-vars"}},[t("a",{staticClass:"header-anchor",attrs:{href:"#env-vars"}},[a._v("#")]),a._v(" Env vars")]),a._v(" "),t("ul",[t("li",[t("code",[a._v("NIMONA_PEER_PRIVATE_KEY")]),a._v(" - Private key for peer. (optional)")]),a._v(" "),t("li",[t("code",[a._v("NIMONA_PEER_BIND_ADDRESS")]),a._v(" - Address (in the "),t("code",[a._v("ip:port")]),a._v(" format) to bind sonar\nto. (optional)")]),a._v(" "),t("li",[t("code",[a._v("NIMONA_PEER_BOOTSTRAPS")]),a._v(" - Bootstrap peers to use (in the\n"),t("code",[a._v("publicKey@tcps:ip:port")]),a._v("  shorthand format). (optional)")]),a._v(" "),t("li",[t("code",[a._v("NIMONA_CHAT_CHANNEL_HASH")]),a._v(" - Channel to join (optional)")])]),a._v(" "),t("h2",{attrs:{id:"example"}},[t("a",{staticClass:"header-anchor",attrs:{href:"#example"}},[a._v("#")]),a._v(" Example")]),a._v(" "),t("p",[a._v("Create some keys for your peers.")]),a._v(" "),t("div",{staticClass:"language-sh extra-class"},[t("pre",{pre:!0,attrs:{class:"language-sh"}},[t("code",[a._v("go run ./cmd/keygen/main.go\n")])])]),t("div",{staticClass:"language-sh extra-class"},[t("pre",{pre:!0,attrs:{class:"language-sh"}},[t("code",[t("span",{pre:!0,attrs:{class:"token assign-left variable"}},[a._v("LOG_LEVEL")]),t("span",{pre:!0,attrs:{class:"token operator"}},[a._v("=")]),a._v("fatal "),t("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v("\\")]),a._v("\n"),t("span",{pre:!0,attrs:{class:"token assign-left variable"}},[a._v("NIMONA_BIND_ADDRESS")]),t("span",{pre:!0,attrs:{class:"token operator"}},[a._v("=")]),t("span",{pre:!0,attrs:{class:"token number"}},[a._v("0.0")]),a._v(".0.0:18000 "),t("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v("\\")]),a._v("\n"),t("span",{pre:!0,attrs:{class:"token assign-left variable"}},[a._v("NIMONA_PEER_PRIVATE_KEY")]),t("span",{pre:!0,attrs:{class:"token operator"}},[a._v("=")]),t("span",{pre:!0,attrs:{class:"token operator"}},[a._v("<")]),a._v("key"),t("span",{pre:!0,attrs:{class:"token operator"}},[t("span",{pre:!0,attrs:{class:"token file-descriptor important"}},[a._v("1")]),a._v(">")]),a._v(" "),t("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v("\\")]),a._v("\ngo run ./examples/chat/*.go\n")])])]),t("div",{staticClass:"language-sh extra-class"},[t("pre",{pre:!0,attrs:{class:"language-sh"}},[t("code",[t("span",{pre:!0,attrs:{class:"token assign-left variable"}},[a._v("LOG_LEVEL")]),t("span",{pre:!0,attrs:{class:"token operator"}},[a._v("=")]),a._v("fatal "),t("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v("\\")]),a._v("\n"),t("span",{pre:!0,attrs:{class:"token assign-left variable"}},[a._v("NIMONA_BIND_ADDRESS")]),t("span",{pre:!0,attrs:{class:"token operator"}},[a._v("=")]),t("span",{pre:!0,attrs:{class:"token number"}},[a._v("0.0")]),a._v(".0.0:18001 "),t("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v("\\")]),a._v("\n"),t("span",{pre:!0,attrs:{class:"token assign-left variable"}},[a._v("NIMONA_PEER_PRIVATE_KEY")]),t("span",{pre:!0,attrs:{class:"token operator"}},[a._v("=")]),t("span",{pre:!0,attrs:{class:"token operator"}},[a._v("<")]),a._v("key"),t("span",{pre:!0,attrs:{class:"token operator"}},[t("span",{pre:!0,attrs:{class:"token file-descriptor important"}},[a._v("2")]),a._v(">")]),a._v(" "),t("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v("\\")]),a._v("\ngo run "),t("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v("..")]),a._v("/examples/chat/*.go\n")])])]),t("div",{staticClass:"language-sh extra-class"},[t("pre",{pre:!0,attrs:{class:"language-sh"}},[t("code",[t("span",{pre:!0,attrs:{class:"token assign-left variable"}},[a._v("LOG_LEVEL")]),t("span",{pre:!0,attrs:{class:"token operator"}},[a._v("=")]),a._v("fatal "),t("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v("\\")]),a._v("\n"),t("span",{pre:!0,attrs:{class:"token assign-left variable"}},[a._v("NIMONA_BIND_ADDRESS")]),t("span",{pre:!0,attrs:{class:"token operator"}},[a._v("=")]),t("span",{pre:!0,attrs:{class:"token number"}},[a._v("0.0")]),a._v(".0.0:18002 "),t("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v("\\")]),a._v("\n"),t("span",{pre:!0,attrs:{class:"token assign-left variable"}},[a._v("NIMONA_PEER_PRIVATE_KEY")]),t("span",{pre:!0,attrs:{class:"token operator"}},[a._v("=")]),t("span",{pre:!0,attrs:{class:"token operator"}},[a._v("<")]),a._v("key"),t("span",{pre:!0,attrs:{class:"token operator"}},[t("span",{pre:!0,attrs:{class:"token file-descriptor important"}},[a._v("3")]),a._v(">")]),a._v(" "),t("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v("\\")]),a._v("\ngo run ./examples/chat/*.go\n")])])])])}),[],!1,null,null,null);s.default=r.exports}}]);