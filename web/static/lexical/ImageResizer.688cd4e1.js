import{r as e,j as t,a as r}from"./main.2a655075.js";function i(e,t,r){return Math.min(Math.max(e,t),r)}const n=1,s=8,o=2,a=4;function c({onResizeStart:c,onResizeEnd:l,buttonRef:u,imageRef:d,maxWidth:m,editor:g,showCaption:h,setShowCaption:p,captionsEnabled:y}){const w=e.exports.useRef(null),z=e.exports.useRef({priority:"",value:"default"}),f=e.exports.useRef({currentHeight:0,currentWidth:0,direction:0,isResizing:!1,ratio:0,startHeight:0,startWidth:0,startX:0,startY:0}),v=g.getRootElement(),b=m||(null!==v?v.getBoundingClientRect().width-20:100),P=null!==v?v.getBoundingClientRect().height-20:100,R=(e,t)=>{if(!g.isEditable())return;const r=d.current,i=w.current;if(null!==r&&null!==i){e.preventDefault();const{width:l,height:u}=r.getBoundingClientRect(),d=f.current;d.startWidth=l,d.startHeight=u,d.ratio=l/u,d.currentWidth=l,d.currentHeight=u,d.startX=e.clientX,d.startY=e.clientY,d.isResizing=!0,d.direction=t,(e=>{const t=e===n||e===a?"ew":e===s||e===o?"ns":e&s&&e&a||e&o&&e&n?"nwse":"nesw";null!==v&&v.style.setProperty("cursor",`${t}-resize`,"important"),null!==document.body&&(document.body.style.setProperty("cursor",`${t}-resize`,"important"),z.current.value=document.body.style.getPropertyValue("-webkit-user-select"),z.current.priority=document.body.style.getPropertyPriority("-webkit-user-select"),document.body.style.setProperty("-webkit-user-select","none","important"))})(t),c(),i.classList.add("image-control-wrapper--resizing"),r.style.height=`${u}px`,r.style.width=`${l}px`,document.addEventListener("pointermove",x),document.addEventListener("pointerup",W)}},x=e=>{const t=d.current,r=f.current,c=r.direction&(n|a),l=r.direction&(o|s);if(null!==t&&r.isResizing)if(c&&l){let s=Math.floor(r.startX-e.clientX);s=r.direction&n?-s:s;const o=i(r.startWidth+s,100,b),a=o/r.ratio;t.style.width=`${o}px`,t.style.height=`${a}px`,r.currentHeight=a,r.currentWidth=o}else if(l){let n=Math.floor(r.startY-e.clientY);n=r.direction&o?-n:n;const s=i(r.startHeight+n,100,P);t.style.height=`${s}px`,r.currentHeight=s}else{let s=Math.floor(r.startX-e.clientX);s=r.direction&n?-s:s;const o=i(r.startWidth+s,100,b);t.style.width=`${o}px`,r.currentWidth=o}},W=()=>{const e=d.current,t=f.current,r=w.current;if(null!==e&&null!==r&&t.isResizing){const e=t.currentWidth,i=t.currentHeight;t.startWidth=0,t.startHeight=0,t.ratio=0,t.startX=0,t.startY=0,t.currentWidth=0,t.currentHeight=0,t.isResizing=!1,r.classList.remove("image-control-wrapper--resizing"),null!==v&&v.style.setProperty("cursor","text"),null!==document.body&&(document.body.style.setProperty("cursor","default"),document.body.style.setProperty("-webkit-user-select",z.current.value,z.current.priority)),l(e,i),document.removeEventListener("pointermove",x),document.removeEventListener("pointerup",W)}};return t("div",{ref:w,children:[!h&&y&&r("button",{className:"image-caption-button",ref:u,onClick:()=>{p(!h)},children:"Add Caption"}),r("div",{className:"image-resizer image-resizer-n",onPointerDown:e=>{R(e,s)}}),r("div",{className:"image-resizer image-resizer-ne",onPointerDown:e=>{R(e,s|n)}}),r("div",{className:"image-resizer image-resizer-e",onPointerDown:e=>{R(e,n)}}),r("div",{className:"image-resizer image-resizer-se",onPointerDown:e=>{R(e,o|n)}}),r("div",{className:"image-resizer image-resizer-s",onPointerDown:e=>{R(e,o)}}),r("div",{className:"image-resizer image-resizer-sw",onPointerDown:e=>{R(e,o|a)}}),r("div",{className:"image-resizer image-resizer-w",onPointerDown:e=>{R(e,a)}}),r("div",{className:"image-resizer image-resizer-nw",onPointerDown:e=>{R(e,s|a)}})]})}export{c as I};
