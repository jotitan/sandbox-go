import React, {useEffect, useState} from 'react'
import Gallery from 'react-grid-gallery'
import axios from "axios";
import {getBaseUrl} from "../treeFolder";


const correctOrientation = orientation => {
    switch(orientation){
        case 0:return 0;
        case 90:return 8;
        case 180:return 3;
        case 270:return 6;
        default:return 0;
    }
}

export default function MyGallery({urlFolder}) {
    const [images,setImages] = useState([])
    let baseUrl = getBaseUrl();
    const loadImages = url => {
        if(url === ''){return;}
        axios({
            method:'GET',
            url:url,
        }).then(d=>{
            // Filter image by time before
            setImages(d.data
                .filter(file=>file.ImageLink != null)
                .sort((img1,img2)=>new Date(img1.Date) - new Date(img2.Date))
                .map(img=>{
                let orient = correctOrientation(img.Orientation)
                return {
                    caption:"",thumbnail:baseUrl + img.ThumbnailLink,src:baseUrl + img.ImageLink,
                    thumbnailWidth:img.Width,//(orient === 0 || orient === 3)?img.Width:img.Height,
                    thumbnailHeight:img.Height//(orient === 0 || orient === 3)?img.Height:img.Height,
                }
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
                     onSelectImage={selectImage}
                     enableImageSelection={true}
                     backdropClosesModal={true} lightboxWidth={2000}/>
</>
    )
}