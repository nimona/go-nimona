(window.webpackJsonp=window.webpackJsonp||[]).push([[3],{309:function(t,n,e){"use strict";e.d(n,"d",(function(){return r})),e.d(n,"a",(function(){return u})),e.d(n,"i",(function(){return a})),e.d(n,"f",(function(){return s})),e.d(n,"g",(function(){return o})),e.d(n,"h",(function(){return c})),e.d(n,"b",(function(){return f})),e.d(n,"e",(function(){return h})),e.d(n,"k",(function(){return p})),e.d(n,"l",(function(){return g})),e.d(n,"c",(function(){return v})),e.d(n,"j",(function(){return m}));e(70),e(49),e(310),e(312),e(173),e(69),e(98),e(99),e(30),e(100),e(171);var r=/#.*$/,i=/\.(md|html)$/,u=/\/$/,a=/^[a-z]+:/i;function l(t){return decodeURI(t).replace(r,"").replace(i,"")}function s(t){return a.test(t)}function o(t){return/^mailto:/.test(t)}function c(t){return/^tel:/.test(t)}function f(t){if(s(t))return t;var n=t.match(r),e=n?n[0]:"",i=l(t);return u.test(i)?t:i+".html"+e}function h(t,n){var e=decodeURIComponent(t.hash),i=function(t){var n=t.match(r);if(n)return n[0]}(n);return(!i||e===i)&&l(t.path)===l(n)}function p(t,n,e){if(s(n))return{type:"external",path:n};e&&(n=function(t,n,e){var r=t.charAt(0);if("/"===r)return t;if("?"===r||"#"===r)return n+t;var i=n.split("/");e&&i[i.length-1]||i.pop();for(var u=t.replace(/^\//,"").split("/"),a=0;a<u.length;a++){var l=u[a];".."===l?i.pop():"."!==l&&i.push(l)}""!==i[0]&&i.unshift("");return i.join("/")}(n,e));for(var r=l(n),i=0;i<t.length;i++)if(l(t[i].regularPath)===r)return Object.assign({},t[i],{type:"page",path:f(t[i].path)});return console.error('[vuepress] No matching page found for sidebar item "'.concat(n,'"')),{}}function g(t,n,e,r){var i=e.pages,u=e.themeConfig,a=r&&u.locales&&u.locales[r]||u;if("auto"===(t.frontmatter.sidebar||a.sidebar||u.sidebar))return d(t);var l=a.sidebar||u.sidebar;if(l){var s=function(t,n){if(Array.isArray(n))return{base:"/",config:n};for(var e in n)if(0===(r=t,/(\.html|\/)$/.test(r)?r:r+"/").indexOf(encodeURI(e)))return{base:e,config:n[e]};var r;return{}}(n,l),o=s.base,c=s.config;return"auto"===c?d(t):c?c.map((function(t){return function t(n,e,r){var i=arguments.length>3&&void 0!==arguments[3]?arguments[3]:1;if("string"==typeof n)return p(e,n,r);if(Array.isArray(n))return Object.assign(p(e,n[0],r),{title:n[1]});var u=n.children||[];return 0===u.length&&n.path?Object.assign(p(e,n.path,r),{title:n.title}):{type:"group",path:n.path,title:n.title,sidebarDepth:n.sidebarDepth,initialOpenGroupIndex:n.initialOpenGroupIndex,children:u.map((function(n){return t(n,e,r,i+1)})),collapsable:!1!==n.collapsable}}(t,i,o)})):[]}return[]}function d(t){var n=v(t.headers||[]);return[{type:"group",collapsable:!1,title:t.title,path:null,children:n.map((function(n){return{type:"auto",title:n.title,basePath:t.path,path:t.path+"#"+n.slug,children:n.children||[]}}))}]}function v(t){var n;return(t=t.map((function(t){return Object.assign({},t)}))).forEach((function(t){2===t.level?n=t:n&&(n.children||(n.children=[])).push(t)})),t.filter((function(t){return 2===t.level}))}function m(t){return Object.assign(t,{type:t.items&&t.items.length?"links":"link"})}},310:function(t,n,e){"use strict";var r=e(168),i=e(8),u=e(16),a=e(26),l=e(169),s=e(170);r("match",1,(function(t,n,e){return[function(n){var e=a(this),r=null==n?void 0:n[t];return void 0!==r?r.call(n,e):new RegExp(n)[t](String(e))},function(t){var r=e(n,t,this);if(r.done)return r.value;var a=i(t),o=String(this);if(!a.global)return s(a,o);var c=a.unicode;a.lastIndex=0;for(var f,h=[],p=0;null!==(f=s(a,o));){var g=String(f[0]);h[p]=g,""===g&&(a.lastIndex=l(o,u(a.lastIndex),c)),p++}return 0===p?null:h}]}))},311:function(t,n,e){"use strict";e(313),e(96),e(97);var r=e(309),i={name:"NavLink",props:{item:{required:!0}},computed:{link:function(){return Object(r.b)(this.item.link)},exact:function(){var t=this;return this.$site.locales?Object.keys(this.$site.locales).some((function(n){return n===t.link})):"/"===this.link},isNonHttpURI:function(){return Object(r.g)(this.link)||Object(r.h)(this.link)},isBlankTarget:function(){return"_blank"===this.target},isInternal:function(){return!Object(r.f)(this.link)&&!this.isBlankTarget},target:function(){return this.isNonHttpURI?null:this.item.target?this.item.target:Object(r.f)(this.link)?"_blank":""},rel:function(){return this.isNonHttpURI||!1===this.item.rel?null:this.item.rel?this.item.rel:this.isBlankTarget?"noopener noreferrer":null}},methods:{focusoutAction:function(){this.$emit("focusout")}}},u=e(48),a=Object(u.a)(i,(function(){var t=this,n=t.$createElement,e=t._self._c||n;return t.isInternal?e("RouterLink",{staticClass:"nav-link",attrs:{to:t.link,exact:t.exact},nativeOn:{focusout:function(n){return t.focusoutAction(n)}}},[t._v("\n  "+t._s(t.item.text)+"\n")]):e("a",{staticClass:"nav-link external",attrs:{href:t.link,target:t.target,rel:t.rel},on:{focusout:t.focusoutAction}},[t._v("\n  "+t._s(t.item.text)+"\n  "),t.isBlankTarget?e("OutboundLink"):t._e()],1)}),[],!1,null,null,null);n.a=a.exports},312:function(t,n,e){"use strict";var r=e(168),i=e(172),u=e(8),a=e(26),l=e(101),s=e(169),o=e(16),c=e(170),f=e(71),h=e(5),p=[].push,g=Math.min,d=!h((function(){return!RegExp(4294967295,"y")}));r("split",2,(function(t,n,e){var r;return r="c"=="abbc".split(/(b)*/)[1]||4!="test".split(/(?:)/,-1).length||2!="ab".split(/(?:ab)*/).length||4!=".".split(/(.?)(.?)/).length||".".split(/()()/).length>1||"".split(/.?/).length?function(t,e){var r=String(a(this)),u=void 0===e?4294967295:e>>>0;if(0===u)return[];if(void 0===t)return[r];if(!i(t))return n.call(r,t,u);for(var l,s,o,c=[],h=(t.ignoreCase?"i":"")+(t.multiline?"m":"")+(t.unicode?"u":"")+(t.sticky?"y":""),g=0,d=new RegExp(t.source,h+"g");(l=f.call(d,r))&&!((s=d.lastIndex)>g&&(c.push(r.slice(g,l.index)),l.length>1&&l.index<r.length&&p.apply(c,l.slice(1)),o=l[0].length,g=s,c.length>=u));)d.lastIndex===l.index&&d.lastIndex++;return g===r.length?!o&&d.test("")||c.push(""):c.push(r.slice(g)),c.length>u?c.slice(0,u):c}:"0".split(void 0,0).length?function(t,e){return void 0===t&&0===e?[]:n.call(this,t,e)}:n,[function(n,e){var i=a(this),u=null==n?void 0:n[t];return void 0!==u?u.call(n,i,e):r.call(String(i),n,e)},function(t,i){var a=e(r,t,this,i,r!==n);if(a.done)return a.value;var f=u(t),h=String(this),p=l(f,RegExp),v=f.unicode,m=(f.ignoreCase?"i":"")+(f.multiline?"m":"")+(f.unicode?"u":"")+(d?"y":"g"),b=new p(d?f:"^(?:"+f.source+")",m),k=void 0===i?4294967295:i>>>0;if(0===k)return[];if(0===h.length)return null===c(b,h)?[h]:[];for(var x=0,_=0,I=[];_<h.length;){b.lastIndex=d?_:0;var O,y=c(b,d?h:h.slice(_));if(null===y||(O=g(o(b.lastIndex+(d?0:_)),h.length))===x)_=s(h,_,v);else{if(I.push(h.slice(x,_)),I.length===k)return I;for(var w=1;w<=y.length-1;w++)if(I.push(y[w]),I.length===k)return I;_=x=O}}return I.push(h.slice(x)),I}]}),!d)},313:function(t,n,e){"use strict";var r=e(3),i=e(314);r({target:"String",proto:!0,forced:e(315)("link")},{link:function(t){return i(this,"a","href",t)}})},314:function(t,n,e){var r=e(26),i=/"/g;t.exports=function(t,n,e,u){var a=String(r(t)),l="<"+n;return""!==e&&(l+=" "+e+'="'+String(u).replace(i,"&quot;")+'"'),l+">"+a+"</"+n+">"}},315:function(t,n,e){var r=e(5);t.exports=function(t){return r((function(){var n=""[t]('"');return n!==n.toLowerCase()||n.split('"').length>3}))}},335:function(t,n,e){},364:function(t,n,e){"use strict";e(335)},370:function(t,n,e){"use strict";e.r(n);var r={name:"Home",components:{NavLink:e(311).a},computed:{data:function(){return this.$page.frontmatter},actionLink:function(){return{}}}},i=(e(364),e(48)),u=Object(i.a)(r,(function(){var t=this.$createElement;this._self._c;return this._m(0)}),[function(){var t=this,n=t.$createElement,e=t._self._c||n;return e("section",{staticClass:"welcome"},[e("div",{staticClass:"container"},[e("h1",{staticClass:"welcome__title"},[t._v("nimona")]),t._v(" "),e("p",{staticClass:"welcome__description"},[t._v("\n      a new internet stack; or something like it\n      "),e("br"),t._v(" "),e("small",[e("a",{attrs:{href:"/docs/"}},[t._v("check out the docs")])]),t._v(",\n      "),e("small",[e("a",{attrs:{href:"https://github.com/nimona/go-nimona"}},[t._v("contribute on github")])])]),t._v(" "),e("figure",{staticClass:"welcome__image"},[e("img",{attrs:{src:"/nimona-logo-icon-white.png"}})])])])}],!1,null,null,null);n.default=u.exports}}]);