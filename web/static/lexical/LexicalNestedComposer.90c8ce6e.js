import{O as e,Q as t,r}from"./main.00ecd330.js";var a=e,s=t,o=r.exports;var n=function({initialEditor:e,children:t,initialNodes:r,initialTheme:n,skipCollabChecks:l}){let i=o.useRef(!1),c=o.useContext(s.LexicalComposerContext);null==c&&function(e){let t=new URLSearchParams;t.append("code",e);for(let r=1;r<arguments.length;r++)t.append("v",arguments[r]);throw Error(`Minified Lexical error #${e}; visit https://lexical.dev/docs/error?${t} for the full message or use the non-minified dev environment for full errors and additional helpful warnings.`)}(9);let[d,{getTheme:f}]=c,p=o.useMemo((()=>{var t=n||f()||void 0;const a=s.createLexicalComposerContext(c,t);if(void 0!==t&&(e._config.theme=t),e._parentEditor=d,r)for(var o of r)t=o.getType(),e._nodes.set(t,{klass:o,replace:null,replaceWithKlass:null,transforms:new Set});else{o=e._nodes=new Map(d._nodes);for(const[t,r]of o)e._nodes.set(t,{klass:r.klass,replace:r.replace,replaceWithKlass:r.replaceWithKlass,transforms:new Set})}return e._config.namespace=d._config.namespace,e._editable=d._editable,[e,a]}),[]),{isCollabActive:u,yjsDocMap:m}=a.useCollaborationContext(),h=l||i.current||m.has(e.getKey());return o.useEffect((()=>{h&&(i.current=!0)}),[h]),o.useEffect((()=>d.registerEditableListener((t=>{e.setEditable(t)}))),[e,d]),o.createElement(s.LexicalComposerContext.Provider,{value:p},!u||h?t:null)};export{n as L};
