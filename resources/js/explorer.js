/* Explore the data of the cluster */

if(Loader){Loader.toLoad("html/explorer.html","Explorer");}

var Explorer = {
    breadcrumb:null,
    panelFolder:null,
    currentPath :"",
    currentTypeLoad:"",
    urlServer:"",
    fctClick:null,
    init:function(){
        $.extend(true,this,Panel) ;
        this.initPanel($('#idExplorePanel'),'<span class="glyphicon glyphicon-hdd"></span> Explore');
        this.div.resizable({minWidth:250});
        this.breadcrumb = $('.breadcrumb',this.div)
        var _self = this;
        this.breadcrumb.on('click','li',function(){
            // delete nexts
            $(this).find('~').remove();
            _self.loadPath($('a',$(this)).data('path'),"",true);
        });
        this.panelFolder = $('.folders',this.div);

        $('.switch',this.div).bind('click',function(){
            Explorer.changeZoom();
        })

        this.div.bind('open',function(){
           Explorer._open(arguments);
        });
        $('.info-folders > span.filter > :text',this.div).bind('keyup',function(){
            var value = $(this).val().toLowerCase();
            if (value.length <=2){
                $('>span',Explorer.panelFolder).show();
                return;
            }
            if (value.length > 2){
                // Fitler results
                $('>span:not([data-idx^="' + value + '"])',Explorer.panelFolder).hide()
                $('>span[class^="' + value + '"]',Explorer.panelFolder).show()
            }
        });
    },
    addClickBehave:function(fct){
       this.fctClick = fct;
    },
    loadPath:function(path,display,noAddBC){
        $('.info-folders > span.filter > :text',this.div).val("")
        this.currentPath = path;
        // Add element in breadcrumb
        if(!noAddBC){
            this.addBreadcrumb(path,display);
        }
        if(this.currentTypeLoad == ""){
            var url = "";
        }
        $.ajax({
            url:this.urlServer + '?' + path,
            dataType:'json',
            success:function(data){
                Explorer.display(data);
            }
        })
    },
    changeZoom:function(){
        if ($('.folders',this.div).hasClass('block')){
            $('.folders',this.div).removeClass('block').addClass('line');
        }else{
            $('.folders',this.div).removeClass('line').addClass('block');
        }
    },
    // Call when first open
    _open:function(){
        this.breadcrumb.empty();
        this.urlServer = arguments[0][1];
        this.loadPath("","Home");
    },
    addBreadcrumb:function(path,display){
        display = display || path;
        this.breadcrumb.append('<li><a href="#" data-path="' + path + '">' + display + '</a></li>');
    },
    display:function(data){
        this.panelFolder.empty();
        var _self = this;
        var nb = 0;
        for(var file in data){
            var name = "";
            var url = "";
            if(Number(file) == file){
                // Case when {}
                name = data[file].name;
                url = data[file].url;
            }else{
                // Normal case of map
                name = file;
                url = "path=" + Explorer.currentPath + file + "/"
            }
            var info = "";
            if(data[file].info!=null){
                info = '<span class="info">' + MusicPlayer._formatTime(data[file].info) + '</span>';
            }
            // Info json with name and either url (param after url) or id
            var span = $('<span data-idx="' + name.toLowerCase() + '" data-url="' + url + '">' + name + info + '</span>');
            if(url != null){
                span.bind('click',function(){
                    Explorer.loadPath($(this).data('url'),$(this).text());
                });
            }else{
                // Last element, display server where data is
                span.data("id",data[file].id)
                span.draggable({revert:true,helper:'clone'})
                // Dbl click to playlist
                if(this.fctClick){
                    span.bind('dblclick',function(){
                        Explorer.fctClick(data[file].id);
                    });
                }
            }
            this.panelFolder.append(span);
        }
        $('.info-folders > span.counter',this.div).html('' + this.panelFolder.find('>span').length + ' element(s)');
    },
    setInfoKey:function(key,span){
        span.attr('data-trigger','focus').attr('tabindex','0').popover({
            title:span.text(),
            html:true,
            content:function(){
                var data = ClusterAction.whereIsKey(key,null,true);
                var str = '<div><div style="font-weight:bold">Master : ' + ClusterAction.getAlias(data[0]) + '</div>';
                str+='<div>Replica(s) : ';
                for(var i = 1 ; i < data.length ; i++){
                    str+=ClusterAction.getAlias(data[i]) + ' ';
                }
                return str + '</div></div>';
            }

        })
    }
}