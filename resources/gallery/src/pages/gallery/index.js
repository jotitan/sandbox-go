import React, {useEffect, useState} from 'react'
import Gallery from 'react-grid-gallery'
import axios from "axios";


export default function MyGallery({urlFolder}) {
    const [images,setImages] = useState([])
    const loadImages = url => {
        if(url === ''){return;}
        axios({
            method:'GET',
            url:url,
        }).then(d=>{
            setImages(d.data.map(img=>{
                return {caption:"",thumbnail:'http://localhost:9004' + img.ThumbnailLink,src:'http://localhost:9004' + img.ImageLink,thumbnailWidth:img.Width/2,thumbnailHeight:img.Height/2}
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