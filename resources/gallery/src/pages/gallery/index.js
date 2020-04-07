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

    useEffect(()=>{
        loadImages(urlFolder)
    },[urlFolder])

    return (
     <>
            <Gallery images={images} enableImageSelection={false}
                     imageCountSeparator={" / "}
                     showLightboxThumbnails={false}
                     showImageCount={false}
                     backdropClosesModal={true} lightboxWidth={2000}/>
</>
    )
}