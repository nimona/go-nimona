(window.webpackJsonp=window.webpackJsonp||[]).push([[3],{318:function(t,n,e){"use strict";e.d(n,"d",(function(){return r})),e.d(n,"a",(function(){return a})),e.d(n,"i",(function(){return u})),e.d(n,"f",(function(){return s})),e.d(n,"g",(function(){return o})),e.d(n,"h",(function(){return c})),e.d(n,"b",(function(){return f})),e.d(n,"e",(function(){return h})),e.d(n,"k",(function(){return p})),e.d(n,"l",(function(){return d})),e.d(n,"c",(function(){return g})),e.d(n,"j",(function(){return m}));e(49),e(75),e(319),e(321),e(176),e(74),e(101),e(102),e(33),e(103),e(173);var r=/#.*$/,i=/\.(md|html)$/,a=/\/$/,u=/^[a-z]+:/i;function l(t){return decodeURI(t).replace(r,"").replace(i,"")}function s(t){return u.test(t)}function o(t){return/^mailto:/.test(t)}function c(t){return/^tel:/.test(t)}function f(t){if(s(t))return t;var n=t.match(r),e=n?n[0]:"",i=l(t);return a.test(i)?t:i+".html"+e}function h(t,n){var e=decodeURIComponent(t.hash),i=function(t){var n=t.match(r);if(n)return n[0]}(n);return(!i||e===i)&&l(t.path)===l(n)}function p(t,n,e){if(s(n))return{type:"external",path:n};e&&(n=function(t,n,e){var r=t.charAt(0);if("/"===r)return t;if("?"===r||"#"===r)return n+t;var i=n.split("/");e&&i[i.length-1]||i.pop();for(var a=t.replace(/^\//,"").split("/"),u=0;u<a.length;u++){var l=a[u];".."===l?i.pop():"."!==l&&i.push(l)}""!==i[0]&&i.unshift("");return i.join("/")}(n,e));for(var r=l(n),i=0;i<t.length;i++)if(l(t[i].regularPath)===r)return Object.assign({},t[i],{type:"page",path:f(t[i].path)});return console.error('[vuepress] No matching page found for sidebar item "'.concat(n,'"')),{}}function d(t,n,e,r){var i=e.pages,a=e.themeConfig,u=r&&a.locales&&a.locales[r]||a;if("auto"===(t.frontmatter.sidebar||u.sidebar||a.sidebar))return v(t);var l=u.sidebar||a.sidebar;if(l){var s=function(t,n){if(Array.isArray(n))return{base:"/",config:n};for(var e in n)if(0===(r=t,/(\.html|\/)$/.test(r)?r:r+"/").indexOf(encodeURI(e)))return{base:e,config:n[e]};var r;return{}}(n,l),o=s.base,c=s.config;return"auto"===c?v(t):c?c.map((function(t){return function t(n,e,r){var i=arguments.length>3&&void 0!==arguments[3]?arguments[3]:1;if("string"==typeof n)return p(e,n,r);if(Array.isArray(n))return Object.assign(p(e,n[0],r),{title:n[1]});var a=n.children||[];return 0===a.length&&n.path?Object.assign(p(e,n.path,r),{title:n.title}):{type:"group",path:n.path,title:n.title,sidebarDepth:n.sidebarDepth,initialOpenGroupIndex:n.initialOpenGroupIndex,children:a.map((function(n){return t(n,e,r,i+1)})),collapsable:!1!==n.collapsable}}(t,i,o)})):[]}return[]}function v(t){var n=g(t.headers||[]);return[{type:"group",collapsable:!1,title:t.title,path:null,children:n.map((function(n){return{type:"auto",title:n.title,basePath:t.path,path:t.path+"#"+n.slug,children:n.children||[]}}))}]}function g(t){var n;return(t=t.map((function(t){return Object.assign({},t)}))).forEach((function(t){2===t.level?n=t:n&&(n.children||(n.children=[])).push(t)})),t.filter((function(t){return 2===t.level}))}function m(t){return Object.assign(t,{type:t.items&&t.items.length?"links":"link"})}},319:function(t,n,e){"use strict";var r=e(170),i=e(8),a=e(15),u=e(23),l=e(27),s=e(171),o=e(172);r("match",(function(t,n,e){return[function(n){var e=l(this),r=null==n?void 0:n[t];return void 0!==r?r.call(n,e):new RegExp(n)[t](u(e))},function(t){var r=i(this),l=u(t),c=e(n,r,l);if(c.done)return c.value;if(!r.global)return o(r,l);var f=r.unicode;r.lastIndex=0;for(var h,p=[],d=0;null!==(h=o(r,l));){var v=u(h[0]);p[d]=v,""===v&&(r.lastIndex=s(l,a(r.lastIndex),f)),d++}return 0===d?null:p}]}))},320:function(t,n,e){"use strict";e(322),e(99),e(100);var r=e(318),i={name:"NavLink",props:{item:{required:!0}},computed:{link:function(){return Object(r.b)(this.item.link)},exact:function(){var t=this;return this.$site.locales?Object.keys(this.$site.locales).some((function(n){return n===t.link})):"/"===this.link},isNonHttpURI:function(){return Object(r.g)(this.link)||Object(r.h)(this.link)},isBlankTarget:function(){return"_blank"===this.target},isInternal:function(){return!Object(r.f)(this.link)&&!this.isBlankTarget},target:function(){return this.isNonHttpURI?null:this.item.target?this.item.target:Object(r.f)(this.link)?"_blank":""},rel:function(){return this.isNonHttpURI||!1===this.item.rel?null:this.item.rel?this.item.rel:this.isBlankTarget?"noopener noreferrer":null}},methods:{focusoutAction:function(){this.$emit("focusout")}}},a=e(48),u=Object(a.a)(i,(function(){var t=this,n=t.$createElement,e=t._self._c||n;return t.isInternal?e("RouterLink",{staticClass:"nav-link",attrs:{to:t.link,exact:t.exact},nativeOn:{focusout:function(n){return t.focusoutAction.apply(null,arguments)}}},[t._v("\n  "+t._s(t.item.text)+"\n")]):e("a",{staticClass:"nav-link external",attrs:{href:t.link,target:t.target,rel:t.rel},on:{focusout:t.focusoutAction}},[t._v("\n  "+t._s(t.item.text)+"\n  "),t.isBlankTarget?e("OutboundLink"):t._e()],1)}),[],!1,null,null,null);n.a=u.exports},321:function(t,n,e){"use strict";var r=e(170),i=e(174),a=e(8),u=e(27),l=e(104),s=e(171),o=e(15),c=e(23),f=e(172),h=e(76),p=e(175),d=e(4),v=p.UNSUPPORTED_Y,g=[].push,m=Math.min;r("split",(function(t,n,e){var r;return r="c"=="abbc".split(/(b)*/)[1]||4!="test".split(/(?:)/,-1).length||2!="ab".split(/(?:ab)*/).length||4!=".".split(/(.?)(.?)/).length||".".split(/()()/).length>1||"".split(/.?/).length?function(t,e){var r=c(u(this)),a=void 0===e?4294967295:e>>>0;if(0===a)return[];if(void 0===t)return[r];if(!i(t))return n.call(r,t,a);for(var l,s,o,f=[],p=(t.ignoreCase?"i":"")+(t.multiline?"m":"")+(t.unicode?"u":"")+(t.sticky?"y":""),d=0,v=new RegExp(t.source,p+"g");(l=h.call(v,r))&&!((s=v.lastIndex)>d&&(f.push(r.slice(d,l.index)),l.length>1&&l.index<r.length&&g.apply(f,l.slice(1)),o=l[0].length,d=s,f.length>=a));)v.lastIndex===l.index&&v.lastIndex++;return d===r.length?!o&&v.test("")||f.push(""):f.push(r.slice(d)),f.length>a?f.slice(0,a):f}:"0".split(void 0,0).length?function(t,e){return void 0===t&&0===e?[]:n.call(this,t,e)}:n,[function(n,e){var i=u(this),a=null==n?void 0:n[t];return void 0!==a?a.call(n,i,e):r.call(c(i),n,e)},function(t,i){var u=a(this),h=c(t),p=e(r,u,h,i,r!==n);if(p.done)return p.value;var d=l(u,RegExp),g=u.unicode,b=(u.ignoreCase?"i":"")+(u.multiline?"m":"")+(u.unicode?"u":"")+(v?"g":"y"),k=new d(v?"^(?:"+u.source+")":u,b),x=void 0===i?4294967295:i>>>0;if(0===x)return[];if(0===h.length)return null===f(k,h)?[h]:[];for(var _=0,y=0,O=[];y<h.length;){k.lastIndex=v?0:y;var I,j=f(k,v?h.slice(y):h);if(null===j||(I=m(o(k.lastIndex+(v?y:0)),h.length))===_)y=s(h,y,g);else{if(O.push(h.slice(_,y)),O.length===x)return O;for(var w=1;w<=j.length-1;w++)if(O.push(j[w]),O.length===x)return O;y=_=I}}return O.push(h.slice(_)),O}]}),!!d((function(){var t=/(?:)/,n=t.exec;t.exec=function(){return n.apply(this,arguments)};var e="ab".split(t);return 2!==e.length||"a"!==e[0]||"b"!==e[1]})),v)},322:function(t,n,e){"use strict";var r=e(3),i=e(323);r({target:"String",proto:!0,forced:e(324)("link")},{link:function(t){return i(this,"a","href",t)}})},323:function(t,n,e){var r=e(27),i=e(23),a=/"/g;t.exports=function(t,n,e,u){var l=i(r(t)),s="<"+n;return""!==e&&(s+=" "+e+'="'+i(u).replace(a,"&quot;")+'"'),s+">"+l+"</"+n+">"}},324:function(t,n,e){var r=e(4);t.exports=function(t){return r((function(){var n=""[t]('"');return n!==n.toLowerCase()||n.split('"').length>3}))}},344:function(t,n,e){},373:function(t,n,e){"use strict";e(344)},379:function(t,n,e){"use strict";e.r(n);var r={name:"Home",components:{NavLink:e(320).a},computed:{data:function(){return this.$page.frontmatter},actionLink:function(){return{}}}},i=(e(373),e(48)),a=Object(i.a)(r,(function(){var t=this.$createElement;this._self._c;return this._m(0)}),[function(){var t=this,n=t.$createElement,e=t._self._c||n;return e("section",{staticClass:"welcome"},[e("script",{attrs:{async:"",defer:"","data-domain":"nimona.io",src:"https://plausible.io/js/plausible.js"}}),t._v(" "),e("div",{staticClass:"container"},[e("h1",{staticClass:"welcome__title"},[t._v("nimona")]),t._v(" "),e("p",{staticClass:"welcome__description"},[t._v("\n      a new internet stack; or something like it\n      "),e("br"),t._v(" "),e("small",[e("a",{attrs:{href:"/docs/"}},[t._v("check out the docs")])]),t._v(",\n      "),e("small",[e("a",{attrs:{href:"https://github.com/nimona/go-nimona"}},[t._v("contribute on github")])])]),t._v(" "),e("figure",{staticClass:"welcome__image"},[e("img",{attrs:{src:"/nimona-logo-icon-white.png"}})])])])}],!1,null,null,null);n.default=a.exports}}]);