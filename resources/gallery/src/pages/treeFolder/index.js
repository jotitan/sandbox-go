import React, {useEffect, useState} from 'react'
import {Tree} from 'antd'
import axios from "axios";

export const getBaseUrl = ()=>{
    console.log(window.location)
    switch (window.location.hostname) {
        case 'localhost':
            return 'http://localhost:9006';
        default : return window.location.origin;

    }
}

const adapt = node => {
    console.log(node)
    let data = {title:node.Name.replace(/_/g," "),key:getBaseUrl() + node.Link}
    if(node.Children != null && node.Children.length > 0){
        data.children = node.Children.map(nc=>adapt(nc));
    }else{
        data.isLeaf=true
    }
    return data;
}

export default function TreeFolder({setUrlFolder}) {
    const [tree,setTree] = useState([])
    const { DirectoryTree } = Tree;

    useEffect(()=>{
         axios({
            method:'GET',
            url:getBaseUrl() + '/rootFolders',
        }).then(d=>{
            setTree([adapt(d.data)] );
         })
    },[])

    const onSelect = (e,f)=>{
        if(f.node.children == null || f.node.children.length === 0) {
            setUrlFolder(e[0])
        }
    }

    return (
     <>
         <DirectoryTree
             defaultExpandAll
             onSelect={onSelect}
             treeData={tree}
             virtual={true}
             style={{fontSize:12+'px',width:300+'px',overflow:'auto',backgroundColor:'#001529',color:'#999'}}

         />
</>
    )
}