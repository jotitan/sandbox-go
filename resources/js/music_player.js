
function Music(id,src,title,time){
    this.id = id;
    this.src = src;
    this.title = title;
    this.time = (time !=null)?parseInt(time):0;
}


var MusicPlayer = {
    player:null,
    // Manage the list of music
    setPlaylist:function(playlist){
        this.playlist = playlist;
        $('.next,.previous',this.div).show();
    },
    // Contains all controls to manipulate player
    controls:{
        div:null,
        seeker:null,
        init:function(idDiv){
            this.div = $('#' + idDiv)
            this.seeker = $('.seeker',this.div);
            this.seeker.slider({
                min:0,
                value:0,
                slide:function(e,ui){
                    MusicPlayer.player.currentTime = ui.value;
                }
            });
            var _self = this;
            $('.play',this.div).bind('click',function(){
               MusicPlayer.play();
            });
            $('.pause',this.div).bind('click',function(){
               MusicPlayer.player.pause();
               $(this).hide();
              $('.play',_self.div).show();
            });
            $('.next',this.div).bind('click',function(){
                MusicPlayer.playlist.next();
            });
            $('.previous',this.div).bind('click',function(){
                MusicPlayer.playlist.previous();
            });
        },
        setTitle:function(title){
            $('.title',this.div).text(title);
        },
        setMax:function(value){
            this.seeker.slider('option','max',value)
            $('.duration',this.div).text(MusicPlayer._formatTime(value));
        } ,
        update:function(value){
            this.seeker.slider('option','value',value)
            $('.position',this.div).text(MusicPlayer._formatTime(value));
        }
    },

    init:function(){
        this.player = $('#idPlayer').get(0);
        this.controls.init('player')
        this.player.addEventListener('canplay',function(e){
            MusicPlayer.checkProgress();
        })
        this.player.addEventListener('error',function(e){
            console.log("Error when loading music")
        });
        this.player.addEventListener('timeupdate',function(e){
            MusicPlayer.controls.update(MusicPlayer.player.currentTime);
        });
        this.player.addEventListener('ended',function(e){
            if(MusicPlayer.playlist!=null){
                MusicPlayer.playlist.next();
            }
        })
    },
    load:function(music){
        this.player.src = music.src;
        this.controls.setTitle(music.title);
        this.play();
    },
    play:function(){
        MusicPlayer.player.play();
        $('.play',this.div).hide();
        $('.pause',this.div).show();
    },

    // launch after load
    checkProgress:function(){
        this.controls.setMax(this.player.duration);
        this.controls.update(0);
    },
    // Format time in second in minutes:secondes
    _formatTime:function(time){
       if(time == null || isNaN(time)) {
        return "00:00";
       }
       time = Math.round(time);
       if(time < 60){
          return "00:" + ((time < 10)?"0":"") + time;
       }
       var min = Math.floor(time/60);
       var rest = time%60;
       return ((min < 10)?"0":"") + min + ":" + ((rest < 10)?"0":"") + rest;
    }

}