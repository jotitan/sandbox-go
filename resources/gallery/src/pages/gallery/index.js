import React, {useEffect, useState} from 'react'
import {Col, Popconfirm, Row, Tooltip} from 'antd'
import Gallery from 'react-grid-gallery'
import axios from "axios";
import {getBaseUrl} from "../treeFolder";
import {DeleteFilled, FileImageOutlined, PictureOutlined} from "@ant-design/icons";

export default function MyGallery({urlFolder}) {
    const [images,setImages] = useState([]);
    const [currentImage,setCurrentImage] = useState(-1);

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
                return {
                    hdLink:baseUrl + img.HdLink,
                    caption:"",thumbnail:baseUrl + img.ThumbnailLink,src:baseUrl + img.ImageLink,
                    thumbnailWidth:img.Width,
                    thumbnailHeight:img.Height
                }
            }));
        })
    };

    useEffect(()=>{
        loadImages(urlFolder)
    },[urlFolder]);

    const selectImage = index=>{
        setImages(list=>{
            let copy = list.slice();
            copy[index].isSelected = list[index].isSelected != null ? !list[index].isSelected : true;
            return copy;
        });
    };

    const deleteSelection = ()=>{
        setImages(images.filter(i=>!i.isSelected));
    }

    const showSelected = ()=>{
        const selected = images.filter(i=>i.isSelected).length;
        return selected > 0 ? <>
            <Popconfirm placement="bottom" title={"Es tu sûr de vouloir supprimer ces photos"}
                        onConfirm={deleteSelection} okText="Oui" cancelText="Non">
                <Tooltip key={"image-info"} placement="top" title={"Supprimer la sélection"} overlayStyle={{zIndex:20000}}>
                    <DeleteFilled style={{cursor:'pointer'}}/>
                </Tooltip>
                <span style={{marginLeft:10+'px'}}>{selected}</span>
            </Popconfirm>
        </>:''
    };
    const whenOpenImage = indexImage=>{
        setCurrentImage(indexImage)
    }

    const getCustomActions = ()=> {
        return [
            <Tooltip key={"image-info"} placement="top" title={"Télécharger en HD"} overlayStyle={{zIndex:20000}}>
                <a target={"_blank"} rel="noopener noreferrer"
                    download={images != null && currentImage !== -1 ? images[currentImage].Name:''}
                    href={images != null && currentImage !== -1 ? images[currentImage].hdLink:''} >
                    <FileImageOutlined style={{color:'white',fontSize:22+'px'}}/>
                </a>
            </Tooltip>
            ]
    };


    return (
     <>
         <Row className={"options"}>
             <Col span={8}>
                 <PictureOutlined /> : {images.length}
             </Col>
             <Col span={8}>
                 {showSelected()}
             </Col>
         </Row>
         <Row className={"gallery"}>
             <Col span={24} style={{marginTop:30+'px'}}>
            <Gallery images={images}
                     imageCountSeparator={" / "}
                     showLightboxThumbnails={false}
                     showImageCount={false}
                     onSelectImage={selectImage}
                     enableImageSelection={true}
                     currentImageWillChange={whenOpenImage}
                     customControls={getCustomActions()}
                     backdropClosesModal={true} lightboxWidth={2000}/>
             </Col>
         </Row>
</>
    )
}