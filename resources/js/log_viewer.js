/* Show log of a server */

if(Loader){Loader.toLoad("html/log.html");}

var LoggerManager = {
    create:function(url,alias){
        $.ajax({
            url:'/urlsLog?url=' + url,
            success:function(data){
                new LoggerPanel(alias,data).open();
            }
        })
    }
}

function LoggerPanel(title,urls){
    this.urls = urls;
    this.logId = null;
    this.panel = null;
    this.isClosed = false;
    this.checkPing = {thread:null,time:0};

    this.init = function(){
        var div = CloneDiv('idLogTemplate','idLog');
        div.resizable();
        this.panel = $('.log-box',div)
        $.extend(this,Panel)
        this.initPanel(div,'Log : ' + title)

        $('.title>span:first',this.div).append(' of ' + title)
        var _self = this;
        this.div.bind('close',function(){
            _self.stop();
            _self.div.remove();
        })

        $('.trash',this.div).bind('click',function(){
            $('div:data(original)',_self.panel).remove();
        })
        this.Searcher.div = $('.search_panel',this.div);
        this.Searcher.init(this.panel);

        $('.search',this.div).bind('click',function(){
            $('.search_panel',_self.div).toggle('slow',function(){
                if(!$(this).is(':visible')){
                    _self.Searcher.reset();
                }
            });
            $('.search_panel :text.search-field',_self.div).focus();
        })

        this.start();
    }

    this.showMessage = function(message){
       if(message.indexOf("_id_log_") == 0){
           this.logId = message.replace("_id_log_","");
       }else{
           var div = $('<div></div>');
           div.data('original',message);
           this.Searcher.formatIfSearch(div);
           this.panel.append(div);
           this.panel.scrollTop(this.panel.prop('scrollHeight'));
       }
    }
    this.Searcher = {
        prefix:'<span style="background-color:#FFFF00">',
        suffix:'</span>',
        div:null,
        panel:null,
        currentShowLine:-1,
        resultsDiv:[],
        currentValue:null,
        init:function(panel){
            this.panel = panel;
            var _self = this;
            $('.previous',this.div).unbind('click').bind('click',function(e){
                _self.previous();
            });
            $('.next',this.div).unbind('click').bind('click',function(){
                _self.next();
            });
            $(':text.search-field',this.div).unbind('keyup').bind('keyup',function(e){
                if(e.keyCode == 13){
                    if(!e.shiftKey){
                        _self.do($(this).val());
                    }else{
                        _self.previous();
                    }
                }
            });
        },
        previous : function(){
            this.currentShowLine--;
            if(this.currentShowLine < 0){
                this.currentShowLine = 0;
                return;
            }
            this.move();
        },
        next:function(){
          this.currentShowLine++;
          if(this.currentShowLine >= this.resultsDiv.length){
            this.currentShowLine =this.resultsDiv.length-1;
            return;
          }
          this.move();
        },
        move:function(){
            $('.nb-results > span:first',this.div).html(this.currentShowLine+1);
            var top = this.resultsDiv[this.currentShowLine].offset().top;
            var current = this.panel.scrollTop();
            this.panel.scrollTop(top+current-150);
        },
        reset:function(){
           this.currentValue = null;
           this.currentShowLine = -1;
           this.resultsDiv = [];
           $('div:data(original)',this.panel).each(function(){
             $(this).html($(this).data('original').replace(/ /g,'&nbsp;'))
           });
        },
        do:function(value){
           if(this.currentValue != null && this.currentValue == value){
               return this.next();
           }
           this.reset();
           this.currentValue = value;
           if(value == ""){
               return;
           }
           value=value.toLowerCase();
           var _self = this;
           var nbResults = $('div:data(original)',this.panel).filter(function(d,f){
             return $(f).data('original').match(value,"gi")!=null;
           }).each(function(){
               _self.resultsDiv.push($(this));
               $(this).html(_self.formatDiv(value,$(this).data('original')));
           }).length;
           $('.nb-results > span:last',this.div).html(nbResults);
           this.next();
        } ,
        formatDiv:function(value,field){
            var final = "";
            var pos = 0;
            while(true){
                var newpos = field.toLowerCase().indexOf(value,pos) ;
                if(newpos!=-1){
                    final +=field.substr(pos,newpos-pos).replace(/ /g,'&nbsp;') + this.prefix
                            + field.substr(newpos,value.length) + this.suffix;
                    pos = newpos + value.length;
                }else{break;}
            }
            return final + field.substr(pos).replace(/ /g,'&nbsp;');
        },
        formatIfSearch:function(div){
            if(this.currentValue!=null && this.currentValue!="" && div.data('original').match(this.currentValue,"gi")){
                this.resultsDiv.push(div);
                var newVal = this.formatDiv(this.currentValue,div.data('original'));
                div.html(newVal);
            }else{
                div.html(div.data('original').replace(/ /g,'&nbsp;'));
            }
        }
    },

    this.stop = function(){
        if(this.checkPing.thread != null) {
            clearInterval(this.checkPing.thread);
            this.checkPing.thread = null;
        }
        if(this.isClosed){
            return;
        }
        $.get(this.urls.stop + "?idLog=" + this.logId);
    }

    this.start = function(){
        var es = new EventSource(this.urls.start)
        var _self = this;
        es.onmessage = function(event){
            _self.showMessage(event.data);
        }
        es.onerror = function(){
            _self.showMessage('<span style="font-style:italic">End of log...</span>')
            _self.isClosed = true;
            es.close();
        }
        // Event to check logger is still alive. Receive event every 5 min
        es.addEventListener('ping',function(){
            _self.checkPing.time = new Date().getTime();
        },false)

        this.checkPing = {
            time:new Date().getTime(),
            thread:setInterval(function(){
                // Check last ping
                if(new Date().getTime() - _self.time > 120000){
                    _self.stop()
                }
            },60000)
        };
    }
    this.init();
}