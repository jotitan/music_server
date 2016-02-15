
function Music(id,src,title,time){
    this.id = id;
    this.src = src;
    this.title = title;
    this.time = (time !=null)?parseInt(time):0;
}


var MusicPlayer = {
    player:null,
    // manage name device
    device:{
        name:"",
        input:null,
        init:function(){
            this.input = $('#idDeviceName');
            if(localStorage && localStorage["deviceName"]!=null){
                this.name =localStorage["deviceName"];
            }else {
                this.name = "No name";
            }
            this.input.val(this.name);
             this.input.bind('click',function(){
               if($(this).hasClass('disabled')){
                $(this).removeClass('disabled').bind('keydown',function(e){
                   if(e.keyCode == 13){
                      MusicPlayer.device.save();
                   }
                });
               }
            });
        },
        save:function(){
           this.name = this.input.val();
           this.input.addClass('disabled').unbind('keydown');
           if(localStorage){
            localStorage["deviceName"] = this.name;
           }
        }
    },
    // Manage the list of music
    setPlaylist:function(playlist){
        this.playlist = playlist;
        $('.next,.previous',this.div).show();
    },
    // Contains all controls to manipulate player
    controls:{
        div:null,
        seeker:null,
        shareManager:null,
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
               if(_self.shareManager!=null){
                _self.shareManager.event('play');
               }
            });
            $('.pause',this.div).bind('click',function(){
               MusicPlayer.pause();
               if(_self.shareManager!=null){
                   _self.shareManager.event('pause');
                  }
            });
            $('.next',this.div).bind('click',function(){
                _self.next();
            });
            $('.previous',this.div).bind('click',function(){
                _self.previous();
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
        next:function(){
            MusicPlayer.playlist.next();
            if(this.shareManager!=null){
                this.shareManager.event('next');
           }
        },
        previous:function(){
            MusicPlayer.playlist.previous();
            if(this.shareManager!=null){
                this.shareManager.event('previous');
           }
        },
        setShareManager:function(manager){
          this.shareManager = manager;
        },
        setTitle:function(title,artist){
            $('.title',this.div).text(title);
            $('title').html(artist + " - " + title);
        },
        setMax:function(value){
            this.seeker.slider('option','max',value)
            $('.duration',this.div).text(MusicPlayer._formatTime(value));
        } ,
        update:function(value){
            // Check value, if max == value, launch next song
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
            MusicPlayer.initMusic();
        })
        this.player.addEventListener('error',function(e){
            console.log("Error when loading music")
        });
        this.player.addEventListener('timeupdate',function(e){
            MusicPlayer.controls.update(MusicPlayer.player.currentTime);
        });
        this.player.addEventListener('ended',function(e){
            MusicPlayer.controls.next();
        });
        // Detect key controls
        $(document).bind('keyup',function(e){
            var key = (e.keyCode != 0)?e.keyCode:e.charCode;
            switch(key){
                case 46 : $(document).trigger('delete_event');break;
                case 80 : $(document).trigger('pause_event');break;
                case 34 : $(document).trigger('next_event');break;
                case 33 : $(document).trigger('previous_event');break;
                case 43 : $(document).trigger('volume_up');break;
                case 45 : $(document).trigger('volume_down');break;
                case 54 : $(document).trigger('volume_down');break;
                case 61 : if(e.shiftKey){$(document).trigger('volume_up');}break;
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
        this.device.init();
    },
    load:function(music){
        if(music == null){return;}
        this.player.src = music.src;
        this.controls.setTitle(music.title,music.artist);
        this.play();
    },
    pause:function(){
        this.player.pause();
        $('.pause',this.div).hide();
        $('.play',this.div).show();
    },
    play:function(){
        if(this.player.src == "" && this.playlist != null){
            // Try to load selected playlist music or first
            this.load(this.playlist.getOne());
            return
        }
        MusicPlayer.player.play();
        $('.play',this.div).hide();
        $('.pause',this.div).show();
    },
    // launch after load
    initMusic:function(){
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

