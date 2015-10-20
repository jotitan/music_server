
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
               MusicPlayer.pause();
            });
            $('.next',this.div).bind('click',function(){
                MusicPlayer.playlist.next();
            });
            $('.previous',this.div).bind('click',function(){
                MusicPlayer.playlist.previous();
            });
            MusicPlayer.player.volume = 0.5;
            // Volume Behaviour
            $('.volume-plus',this.div).bind('click',function(){
                MusicPlayer.volume.up();
            });
            $('.volume-minus',this.div).bind('click',function(){
                MusicPlayer.volume.down();
            });
            $(document).bind('volume_up',function(){MusicPlayer.volume.up();})
            $(document).bind('volume_down',function(){MusicPlayer.volume.down();})
            VolumeDrawer.init('idVolume');
            VolumeDrawer.draw(Math.round(MusicPlayer.player.volume*10))
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
    volume:{
        up:function(){
            if(MusicPlayer.player.volume >= 0.9){
                 MusicPlayer.player.volume=1;
             }else{
                 MusicPlayer.player.volume+=0.1;
             }
             VolumeDrawer.draw(Math.round(MusicPlayer.player.volume*10))
        },
        down:function(){
            if(MusicPlayer.player.volume <= 0.1){
                MusicPlayer.player.volume=0;
            }else{
                MusicPlayer.player.volume-=0.1;
            }
            VolumeDrawer.draw(Math.round(MusicPlayer.player.volume*10))
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
        });
        // Detect key controls
        $(document).bind('keypress',function(e){
            var key = (e.keyCode != 0)?e.keyCode:e.charCode;
            switch(key){
                case 46 : $(document).trigger('delete_event');break;
                case 112 : $(document).trigger('pause_event');break;
                case 34 : $(document).trigger('next_event');break;
                case 33 : $(document).trigger('previous_event');break;
                case 43 : $(document).trigger('volume_up');break;
                case 45 : $(document).trigger('volume_down');break;
            }
        });
        $(document).unbind('pause_event').bind('pause_event',function(){
            if(MusicPlayer.player.src == ""){
                return;
            }
            if(MusicPlayer.player.paused){
                MusicPlayer.play();
            }else{
                MusicPlayer.pause();
            }
        });
    },
    load:function(music){
        this.player.src = music.src;
        this.controls.setTitle(music.title);
        this.play();
    },
    pause:function(){
        this.player.pause();
        $('.pause',this.div).hide();
        $('.play',this.div).show();
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

var VolumeDrawer = {
    canvas:null,
    step:Math.PI/5,
    init:function(id){
        this.canvas = $('#' + id).get(0).getContext("2d");
    },
    draw:function(pourcent){
        this.canvas.clearRect(0,0,30,30)
        this.canvas.fillStyle = '#303030'
        this.canvas.save()
        this.canvas.translate(12,12)
        for(var i = 0 ; i < 10 ; i++){
          if(i > pourcent-1){
              this.canvas.fillStyle = '#c6c6c6'
          }
          this.canvas.rotate(this.step);
          this.canvas.fillRect(0,4,2,8)
        }
        this.canvas.restore()
    }
}

