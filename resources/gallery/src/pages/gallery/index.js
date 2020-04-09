import React, {useEffect, useState} from 'react'
import Gallery from 'react-grid-gallery'
import axios from "axios";
import {getBaseUrl} from "../treeFolder";



export default function MyGallery({urlFolder}) {
    const [images,setImages] = useState([])
    let baseUrl = getBaseUrl();
    const loadImages = url => {
        if(url === ''){return;}
        axios({
            method:'GET',
            url:url,
        }).then(d=>{
            setImages(d.data.filter(file=>file.ImageLink != null).map(img=>{
                return {caption:"",thumbnail:baseUrl + img.ThumbnailLink,src:baseUrl + img.ImageLink,thumbnailWidth:img.Width/2,thumbnailHeight:img.Height/2}
            }));
        })
    }

    const selectImage = index=>{
        setImages(list=>{
            let copy = list.slice();
            copy[index].isSelected = list[index].isSelected != null ? !list[index].isSelected : true;
            return copy;
        });

    };

    useEffect(()=>{
        loadImages(urlFolder)
    },[urlFolder])

    return (
     <>
            <Gallery images={images} enableImageSelection={false}
                     imageCountSeparator={" / "}
                     showLightboxThumbnails={false}
                     showImageCount={false}
                     //onSelectImage={selectImage}
                     //enableImageSelection={true}
                     backdropClosesModal={true} lightboxWidth={2000}/>
</>
    )
}