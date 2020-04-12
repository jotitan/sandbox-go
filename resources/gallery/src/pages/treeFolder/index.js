import React, {useEffect, useState} from 'react'
import {Tree} from 'antd'
import axios from "axios";

export const getBaseUrl = ()=>{
    switch (window.location.hostname) {
        case 'localhost':
            return 'http://localhost:9006';
        default : return window.location.origin;
    }
}

const sortByName = (a,b)=>a.Name === b.name ? 0:a.Name < b.Name ? -1:1;

const adapt = node => {
    let data = {title:node.Name.replace(/_/g," "),key:getBaseUrl() + node.Link}
    data.hasImages = node.HasImages;
    if(node.Children != null && node.Children.length > 0){
        data.children = node.Children.sort(sortByName).map(nc=>adapt(nc));
    }else{
        data.isLeaf=true
    }
    return data;
}

export default function TreeFolder({setUrlFolder}) {
    const [tree,setTree] = useState([]);
    const { DirectoryTree } = Tree;
    const [height,setHeight] = useState(window.innerHeight-95);

    useEffect(()=>{
        axios({
            method:'GET',
            url:getBaseUrl() + '/rootFolders',
        }).then(d=>{
            setTree([adapt(d.data)]);
        })
    },[]);

    const onSelect = (e,f)=>{
        if(f.node.children == null || f.node.children.length === 0) {
            setUrlFolder(e[0])
        }else{
            // Case when folder has sub folders but also images
            if(f.node.hasImages){
                setUrlFolder(e[0])
            }
        }
    }

    window.addEventListener('resize', ()=>setHeight(window.innerHeight-95));


    return (
        <>
            <DirectoryTree
                defaultExpandAll
                onSelect={onSelect}
                treeData={tree}
                height={height}
                virtual={true}
                style={{fontSize:12+'px',width:300+'px',overflow:'auto',backgroundColor:'#001529',color:'#999'}}

            />
        </>
    )
}