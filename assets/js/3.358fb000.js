(window.webpackJsonp=window.webpackJsonp||[]).push([[3],{335:function(t,n,e){"use strict";e.d(n,"d",(function(){return r})),e.d(n,"a",(function(){return u})),e.d(n,"i",(function(){return a})),e.d(n,"f",(function(){return s})),e.d(n,"g",(function(){return l})),e.d(n,"h",(function(){return c})),e.d(n,"b",(function(){return f})),e.d(n,"e",(function(){return h})),e.d(n,"k",(function(){return p})),e.d(n,"l",(function(){return d})),e.d(n,"c",(function(){return v})),e.d(n,"j",(function(){return m}));e(33),e(87),e(336),e(112),e(338),e(192),e(85),e(115),e(9),e(116),e(37),e(117),e(190);var r=/#.*$/,i=/\.(md|html)$/,u=/\/$/,a=/^[a-z]+:/i;function o(t){return decodeURI(t).replace(r,"").replace(i,"")}function s(t){return a.test(t)}function l(t){return/^mailto:/.test(t)}function c(t){return/^tel:/.test(t)}function f(t){if(s(t))return t;var n=t.match(r),e=n?n[0]:"",i=o(t);return u.test(i)?t:i+".html"+e}function h(t,n){var e=decodeURIComponent(t.hash),i=function(t){var n=t.match(r);if(n)return n[0]}(n);return(!i||e===i)&&o(t.path)===o(n)}function p(t,n,e){if(s(n))return{type:"external",path:n};e&&(n=function(t,n,e){var r=t.charAt(0);if("/"===r)return t;if("?"===r||"#"===r)return n+t;var i=n.split("/");e&&i[i.length-1]||i.pop();for(var u=t.replace(/^\//,"").split("/"),a=0;a<u.length;a++){var o=u[a];".."===o?i.pop():"."!==o&&i.push(o)}""!==i[0]&&i.unshift("");return i.join("/")}(n,e));for(var r=o(n),i=0;i<t.length;i++)if(o(t[i].regularPath)===r)return Object.assign({},t[i],{type:"page",path:f(t[i].path)});return console.error('[vuepress] No matching page found for sidebar item "'.concat(n,'"')),{}}function d(t,n,e,r){var i=e.pages,u=e.themeConfig,a=r&&u.locales&&u.locales[r]||u;if("auto"===(t.frontmatter.sidebar||a.sidebar||u.sidebar))return g(t);var o=a.sidebar||u.sidebar;if(o){var s=function(t,n){if(Array.isArray(n))return{base:"/",config:n};for(var e in n)if(0===(r=t,/(\.html|\/)$/.test(r)?r:r+"/").indexOf(encodeURI(e)))return{base:e,config:n[e]};var r;return{}}(n,o),l=s.base,c=s.config;return"auto"===c?g(t):c?c.map((function(t){return function t(n,e,r){var i=arguments.length>3&&void 0!==arguments[3]?arguments[3]:1;if("string"==typeof n)return p(e,n,r);if(Array.isArray(n))return Object.assign(p(e,n[0],r),{title:n[1]});var u=n.children||[];return 0===u.length&&n.path?Object.assign(p(e,n.path,r),{title:n.title}):{type:"group",path:n.path,title:n.title,sidebarDepth:n.sidebarDepth,initialOpenGroupIndex:n.initialOpenGroupIndex,children:u.map((function(n){return t(n,e,r,i+1)})),collapsable:!1!==n.collapsable}}(t,i,l)})):[]}return[]}function g(t){var n=v(t.headers||[]);return[{type:"group",collapsable:!1,title:t.title,path:null,children:n.map((function(n){return{type:"auto",title:n.title,basePath:t.path,path:t.path+"#"+n.slug,children:n.children||[]}}))}]}function v(t){var n;return(t=t.map((function(t){return Object.assign({},t)}))).forEach((function(t){2===t.level?n=t:n&&(n.children||(n.children=[])).push(t)})),t.filter((function(t){return 2===t.level}))}function m(t){return Object.assign(t,{type:t.items&&t.items.length?"links":"link"})}},336:function(t,n,e){"use strict";var r=e(12),i=e(187),u=e(11),a=e(86),o=e(24),s=e(32),l=e(58),c=e(188),f=e(189);i("match",(function(t,n,e){return[function(n){var e=s(this),i=null==n?void 0:l(n,t);return i?r(i,n,e):new RegExp(n)[t](o(e))},function(t){var r=u(this),i=o(t),s=e(n,r,i);if(s.done)return s.value;if(!r.global)return f(r,i);var l=r.unicode;r.lastIndex=0;for(var h,p=[],d=0;null!==(h=f(r,i));){var g=o(h[0]);p[d]=g,""===g&&(r.lastIndex=c(i,a(r.lastIndex),l)),d++}return 0===d?null:p}]}))},337:function(t,n,e){"use strict";e(339),e(113),e(9),e(114);var r=e(335),i={name:"NavLink",props:{item:{required:!0}},computed:{link:function(){return Object(r.b)(this.item.link)},exact:function(){var t=this;return this.$site.locales?Object.keys(this.$site.locales).some((function(n){return n===t.link})):"/"===this.link},isNonHttpURI:function(){return Object(r.g)(this.link)||Object(r.h)(this.link)},isBlankTarget:function(){return"_blank"===this.target},isInternal:function(){return!Object(r.f)(this.link)&&!this.isBlankTarget},target:function(){return this.isNonHttpURI?null:this.item.target?this.item.target:Object(r.f)(this.link)?"_blank":""},rel:function(){return this.isNonHttpURI||!1===this.item.rel?null:this.item.rel?this.item.rel:this.isBlankTarget?"noopener noreferrer":null}},methods:{focusoutAction:function(){this.$emit("focusout")}}},u=e(57),a=Object(u.a)(i,(function(){var t=this,n=t.$createElement,e=t._self._c||n;return t.isInternal?e("RouterLink",{staticClass:"nav-link",attrs:{to:t.link,exact:t.exact},nativeOn:{focusout:function(n){return t.focusoutAction.apply(null,arguments)}}},[t._v("\n  "+t._s(t.item.text)+"\n")]):e("a",{staticClass:"nav-link external",attrs:{href:t.link,target:t.target,rel:t.rel},on:{focusout:t.focusoutAction}},[t._v("\n  "+t._s(t.item.text)+"\n  "),t.isBlankTarget?e("OutboundLink"):t._e()],1)}),[],!1,null,null,null);n.a=a.exports},338:function(t,n,e){"use strict";var r=e(39),i=e(12),u=e(5),a=e(187),o=e(191),s=e(11),l=e(32),c=e(118),f=e(188),h=e(86),p=e(24),d=e(58),g=e(40),v=e(189),m=e(89),b=e(186),k=e(6),x=b.UNSUPPORTED_Y,_=Math.min,O=[].push,y=u(/./.exec),I=u(O),j=u("".slice);a("split",(function(t,n,e){var u;return u="c"=="abbc".split(/(b)*/)[1]||4!="test".split(/(?:)/,-1).length||2!="ab".split(/(?:ab)*/).length||4!=".".split(/(.?)(.?)/).length||".".split(/()()/).length>1||"".split(/.?/).length?function(t,e){var u=p(l(this)),a=void 0===e?4294967295:e>>>0;if(0===a)return[];if(void 0===t)return[u];if(!o(t))return i(n,u,t,a);for(var s,c,f,h=[],d=(t.ignoreCase?"i":"")+(t.multiline?"m":"")+(t.unicode?"u":"")+(t.sticky?"y":""),v=0,b=new RegExp(t.source,d+"g");(s=i(m,b,u))&&!((c=b.lastIndex)>v&&(I(h,j(u,v,s.index)),s.length>1&&s.index<u.length&&r(O,h,g(s,1)),f=s[0].length,v=c,h.length>=a));)b.lastIndex===s.index&&b.lastIndex++;return v===u.length?!f&&y(b,"")||I(h,""):I(h,j(u,v)),h.length>a?g(h,0,a):h}:"0".split(void 0,0).length?function(t,e){return void 0===t&&0===e?[]:i(n,this,t,e)}:n,[function(n,e){var r=l(this),a=null==n?void 0:d(n,t);return a?i(a,n,r,e):i(u,p(r),n,e)},function(t,r){var i=s(this),a=p(t),o=e(u,i,a,r,u!==n);if(o.done)return o.value;var l=c(i,RegExp),d=i.unicode,g=(i.ignoreCase?"i":"")+(i.multiline?"m":"")+(i.unicode?"u":"")+(x?"g":"y"),m=new l(x?"^(?:"+i.source+")":i,g),b=void 0===r?4294967295:r>>>0;if(0===b)return[];if(0===a.length)return null===v(m,a)?[a]:[];for(var k=0,O=0,y=[];O<a.length;){m.lastIndex=x?0:O;var w,C=v(m,x?j(a,O):a);if(null===C||(w=_(h(m.lastIndex+(x?O:0)),a.length))===k)O=f(a,O,d);else{if(I(y,j(a,k,O)),y.length===b)return y;for(var R=1;R<=C.length-1;R++)if(I(y,C[R]),y.length===b)return y;O=k=w}}return I(y,j(a,k)),y}]}),!!k((function(){var t=/(?:)/,n=t.exec;t.exec=function(){return n.apply(this,arguments)};var e="ab".split(t);return 2!==e.length||"a"!==e[0]||"b"!==e[1]})),x)},339:function(t,n,e){"use strict";var r=e(3),i=e(340);r({target:"String",proto:!0,forced:e(341)("link")},{link:function(t){return i(this,"a","href",t)}})},340:function(t,n,e){var r=e(5),i=e(32),u=e(24),a=/"/g,o=r("".replace);t.exports=function(t,n,e,r){var s=u(i(t)),l="<"+n;return""!==e&&(l+=" "+e+'="'+o(u(r),a,"&quot;")+'"'),l+">"+s+"</"+n+">"}},341:function(t,n,e){var r=e(6);t.exports=function(t){return r((function(){var n=""[t]('"');return n!==n.toLowerCase()||n.split('"').length>3}))}},363:function(t,n,e){},392:function(t,n,e){"use strict";e(363)},398:function(t,n,e){"use strict";e.r(n);var r={name:"Home",components:{NavLink:e(337).a},computed:{data:function(){return this.$page.frontmatter},actionLink:function(){return{}}}},i=(e(392),e(57)),u=Object(i.a)(r,(function(){var t=this.$createElement;this._self._c;return this._m(0)}),[function(){var t=this,n=t.$createElement,e=t._self._c||n;return e("section",{staticClass:"welcome"},[e("script",{attrs:{async:"",defer:"","data-domain":"nimona.io",src:"https://plausible.io/js/plausible.js"}}),t._v(" "),e("div",{staticClass:"container"},[e("h1",{staticClass:"welcome__title"},[t._v("nimona")]),t._v(" "),e("p",{staticClass:"welcome__description"},[t._v("\n      a new internet stack; or something like it\n      "),e("br"),t._v(" "),e("small",[e("a",{attrs:{href:"/docs/"}},[t._v("check out the docs")])]),t._v(",\n      "),e("small",[e("a",{attrs:{href:"https://github.com/nimona/go-nimona"}},[t._v("contribute on github")])])]),t._v(" "),e("figure",{staticClass:"welcome__image"},[e("img",{attrs:{src:"/nimona-logo-icon-white.png"}})])])])}],!1,null,null,null);n.default=u.exports}}]);