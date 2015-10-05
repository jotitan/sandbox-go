/* Explore the data of the cluster */

if(Loader){Loader.toLoad("html/explorer.html","Explorer");}

var Explorer = {
    breadcrumb:null,
    panelFolder:null,
    currentPath :"",
    currentTypeLoad:"",
    urlServer:"",
    init:function(){
        $.extend(this,Panel) ;
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

        this.div.bind('open',function(){
           Explorer._open(arguments);
        });
    },
    loadPath:function(path,display,noAddBC){
        this.currentPath = path;
        // Add element in breadcrumb
        if(!noAddBC){
            this.addBreadcrumb(path,display);
        }
        if(this.currentTypeLoad == ""){
            var url = "";
        }
        $.ajax({
            url:this.urlServer + '?path=' + path,
            dataType:'json',
            success:function(data){
                Explorer.display(data);
            }
        })
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
            var span = $('<span class="' + file + '">' + file + '</span>');
            if(data[file] == ""){
                span.bind('click',function(){
                    Explorer.loadPath(Explorer.currentPath + $(this).text() + "/",$(this).text());
                });
            }else{
                // Last element, display server where data is
                span.data("music",data[file])
                span.draggable({revert:true,helper:'clone'})
                // Element can be dragged
                //this.setInfoKey(data[file],span);
            }
            this.panelFolder.append(span);
        }
        $('.info-folders',this.div).html('' + this.panelFolder.find('span').length + ' element(s)');
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