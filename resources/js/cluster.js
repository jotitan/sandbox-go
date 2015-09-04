/* Cluster manager */

if(Loader){Loader.toLoad("html/node.html");}

var disconnectTime


function initClusterStatus(){
    WindowsNavManager.init("idWindowsNav")
    MiniStatusViewer.init("idMiniStatus")

    loadClusterdata()
    autoRefresh(FREQUENCE_REFRESH)

    // To toogle menu when click on button
    $('.dropdown-menu > li').bind('click',function(){$('.collapse').collapse('hide')})

    ClusterAction.init()
    GraphicAction.init()

    $('.show-time').each(function(){
        new Timer($(this))
    })
}

function autoRefresh(frq){
    setInterval(function(){
        loadClusterdata()
    },frq)
}

function loadLightCluster(){
  $.ajax({
      url:'/cluster',
      dataType:'json',
      success:function(data){
          showLightCluster(data)
      }
  })
}

function showLightCluster(data){
    $('.light_node').removeData('update')
    for (var i in data){
        if (i != "cluster"){
            id = getId(i)
            node = $('.light_node#' + id,'#status')
            if (node.length == 0){
                node = $('<div id="' + id + '" class="light_node" title="' + i + '"></div>')
                $('#status').append(node)
            }
            node.data('update','1')
            var className = getClassStatus(data[i].Status)
            if(node.attr('class').indexOf(className) == -1){
                removeClass = node.attr('class').replace('light_node ','')
                node.switchClass(removeClass!='light_node'?removeClass:'',className)
            }
        }
    }
    $('.light_node:not(:data(update)):visible').remove()
}


function loadSSE(){
    var SSE = new EventSource('/statsAsSSE');
    SSE.onmessage = function(event){
        console.log(event)
    }
    SSE.onerror = function(){
        console.log('error',arguments)
        SSE.close();
    }
}

function loadClusterdata(){
    loadSSE();
    $.ajax({
        url:'/allStats',
        dataType:'json',
        success:function(data){
            disconnectTime = null
            $('.node').removeClass('disconnect')
            $('#idInfoCluster').html('')
            showCluster(data)
        },error:function(){
            disconnectTime = disconnectTime || new Date()
            delta = Math.round((new Date(new Date().getTime() - disconnectTime.getTime()))/1000)
            $('#idInfoCluster').html('Off since ' + delta + ' s')
            $('.node').addClass('disconnect')
        }
    })
}

/* Display cluster info */
function showCluster(data){
    $('.node').removeData('update')
    nbUp = 0
    taskers = 0
    tasks = 0
    //replica =0
    for (var i in data){
        taskers+=data[i]["NbTaskers"]
        tasks+=data[i]["NbTasks"]
        data[i].Status = (data[i].ID == "") ? 0:1
        ClusterAction.loadInfo(data[i].ID,data[i])
        MiniStatusViewer.add(i,true)
    }
    $('#idNbTaskers').html(taskers)
    $('#idNbTasks').html(tasks)
    BarManager.update(tasks)
    pourcent = taskers *100/(Math.max(tasks,taskers))
    $('#idTotalBar').css('width',pourcent + '%')
    MiniStatusViewer.refresh()
    $('#idTotalBar').removeClass().addClass("progress-bar progress-bar-success");
    $('.node:not(:data(update)):visible').remove()
}

/* Display cluster memory status in footer */
var BarManager = {
    data:[],
    bar:null,
    titleBloc:null,
    nbByGraph:50,
    formater:null,
    init:function(id,formater){
         this.bar = $('#' + id).peity("line",{width:100,height:20})
         this.titleBloc = $('#' + id).next()
         this.formater = formater;
    },
    update:function(value){
        if(this.bar == null){return;}
        if(this.data.length > this.nbByGraph){
            this.data.shift();
        }
        this.data.push(value)
        this.bar.text(this.data.join(",")).change()
        if(this.formater != null){
            value = this.formater(value)
        }
        this.titleBloc.attr('title',value)
    }
}

function statusCluster(nbUp,nbTotal,replica){
   switch (Math.min(nbTotal - nbUp,replica+1)){
       case 0 :$('#idTotalBar').removeClass().addClass("progress-bar progress-bar-success");break;
       case 3 :$('#idTotalBar').removeClass().addClass("progress-bar progress-bar-danger");break;
       default :$('#idTotalBar').removeClass().addClass("progress-bar progress-bar-warning")
   }
}

var ClusterAction = {
    // Current displayed data
    currentDatas:[],
    currentGraph:null,
    graphPanelManager:null,
    init:function(){
        this.graphPanelManager = $('#addGraphPanel')
        BarManager.init("mem-status",function(v){return v + " task(s)"})
        $('.message-cluster').bind('click',function(){
            $(this).hide('fade');
        })
    },
    message:function(title,message,error){
      $('.message-cluster').html('<span class="title">' + title + '</span> : ' + message)
        .removeClass('error success').addClass(error ? 'error':'success').show('fade');
    },
    getAlias:function(url){
      var id = getId(url);
      return this.currentDatas[id].Alias || url;
    },
    setAlias:function(node,oldName){
        var aliasName = prompt("Set the alias of the server",oldName)
        if (aliasName != null && aliasName!=""){
            $.ajax({
                url:'/setAlias?server=' + node + '&alias=' + aliasName
            })
        }
    },
    loadInfo:function(url,data){
        this.currentDatas[getId(url)] = data
        // Create if not exist
        if ($('#' + getId(url)).length == 0){
            div = $('#template').clone()
            div.attr('id',getId(url)).css('display','').find('[name="url"]').text(url)
            div.data('url',url)
            $('#idCluster').append(div)
            if($('.agent_node',div).length > 0){
                ClusterAction.agent.init(div,url,data.Alias)
            }

        }
        // Check name change
        currentName =$('span[name="url"]','#' + getId(url)).text()
         $('span[name="url"]','#' + getId(url)).attr('title',url)
        if (data.Alias!="" && data.Alias!=currentName){
            $('span[name="url"]','#' + getId(url)).text(data.Alias)
        }
        var node = $('#' + getId(url));
        node.data('update','1')
        var className = getClassStatus(data.Status)
        if(node.attr('class').indexOf(className) == -1){
            removeClass = node.attr('class').replace('node ','')
            node.switchClass(removeClass!='node'?removeClass:'',className)
        }
        node.find('span[name],div[name]').each(function(i){
            $('div',this).removeData('visited')
            format = $(this).data("format") // Format the content
            modify = $(this).data("modify") // Format the element
            valid = $(this).data("valid")
            type = $(this).data("type")
            name = $(this).attr('name')
            // Progress task bar for exemple
            if (type == "list" && data[name]){
                data[name].forEach(function(e){
                    element = $('.element_' + e.ID + ':first',this)
                    if(element.length == 0){
                        element = $('.template',this).clone()
                        $(this).append(element)
                        element.removeClass("template").addClass('element_' + e.ID).addClass('visitable')
                    }
                    element.data('visited',true)
                    $('[data-id]',element).each(function(){
                        modify = $(this).data("modify")
                        if(modify != null){
                            moreInfo = $(this).data("moreinfo")
                            var valueMoreInfo = (moreInfo != null)?e[moreInfo]:null;
                            window[modify]($(this),e[$(this).data('id')],valueMoreInfo)
                        }else{
                            $(this).text(e[$(this).data("id")])
                        }
                    })
                },this)

                return
            }
            if (format!=null){
                $(this).text(window[format](data[name]))
            } else{
                if (modify != null){
                    window[modify]($(this),data[name])
                }else{
                    if(data[name] != null){
                        $(this).text(data[name])
                    }
                }
            }
            if (valid != null){
                if (window[valid](data[name])){
                    $(this).parent().show()
                }else{
                    $(this).parent().hide()
                }
            }
        })
        $('div.visitable:not(:data(visited))',node).remove()
    },
    // Show temporary a message
    showTemp:function(urlNode,message,duration){
        var node = $('#' + getId(urlNode))
        div = $('<div class="temp_message">' + message + '</div>')
        node.append(div)
        div.animate({opacity:1},600).delay(duration).animate({opacity:0},500,function(){div.remove()})
    }
}

function getId(url){
    return "node_" + url.replace(/[.:/]/g,"_")
}
