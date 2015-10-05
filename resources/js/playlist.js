/* Show tasks from server */

if(Loader){Loader.toLoad("html/playlist.html","PlaylistPanel");}

var PlaylistPanel = {
    listDiv:null,
    list:[],
    current:-1,
    init:function(){
        $.extend(this,Panel)
        this.initPanel($('#idPlaylist'),'<span class="glyphicon glyphicon-list-alt icon"></span>Tasks')
        this.div.resizable()
        this.listDiv = $('.playlist',this.div)
        // Select behaviour
        this.listDiv.on("click",'div:not(.head)',function(){
            $('div',PlaylistPanel.listDiv).removeClass('focused');
            $(this).addClass('focused');
        });
        this.listDiv.on("dblclick",'div:not(.head)',function(e){
           window.getSelection().removeAllRanges()
           $('div',PlaylistPanel.listDiv).removeClass('played selected');
           $(this).addClass('played');
           MusicPlayer.load($(this).data("music"));
           PlaylistPanel.current = $(this).data("position");
        });
        this.listDiv.droppable({
            drop:function(event,ui){
                var idMusic = ui.draggable.data('music');
                // Get info from id music
                PlaylistPanel.addMusicFromId(idMusic);
            }
        })

    },
    addMusicFromId:function(id){
        $.ajax({
            url:'/musicInfo?id=' + id,
            dataType:'json',
            success:function(data){
                // No need to create a real Music, just a container with properties, no methods
                PlaylistPanel.add(data)
            }
        })
    },
    // Add a new music in list
    add:function(music){
        var position = $('div',this.listDiv).length;
        var line = $('<div><span>' + position + '</span><span>' + music.title + '</span><span>' + MusicPlayer._formatTime(music.time) + '</span></div>');
        line.data("position",position-1);
        line.data("music",music);
        this.listDiv.append(line);
        this.list.push(music);
    },
    _selectLine:function(){
        var line = $('div:nth-child(' + (this.current+2) + ')',this.listDiv);
        $('div',this.listDiv).removeClass('played focused');
        line.addClass('played');
    },
    next:function(){
        if(this.current+1>=this.list.length){
            return;
        }
        this.current++;
        this._selectLine();
        MusicPlayer.load(this.list[this.current]);
    },
    previous:function(){
        if(this.current<=0){
            return;
        }
        this.current--;
        this._selectLine();
        MusicPlayer.load(this.list[this.current]);
    }

}