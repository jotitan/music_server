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
                this.name = "Default";
            }
            this.input.val(this.name);
            this.input.bind('click',(e)=>{
                if(this.input.hasClass('disabled')){
                    $(this.input).removeClass('disabled').bind('keydown',(e)=>{
                        if(e.keyCode == 13){
                            MusicPlayer.device.save();
                        }
                    }).bind('blur',(e)=>MusicPlayer.device.save());
                }
            });
        },
        getName(){
            return name != "" ? name : "No device name";
        },
        save:function(){
            this.name = this.input.val();
            this.input.addClass('disabled').unbind('keydown,blur');
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
        // Hide title
        quizzMode:false,
        seeker:null,
        // Default implementation to avoid test shareManager every time
        shareManager:Share.emptyManager,
        init:function(idDiv){
            this.div = $('#' + idDiv)
            this.seeker = $('.seeker',this.div);
            this.seeker.slider({
                min:0,
                value:0,
                slide:(e,ui)=>MusicPlayer.player.currentTime = ui.value
            });
            MusicPlayer.player.volume = 0.6;
            this.initActions();
            // Volume Behaviour
            VolumeDrawer.init('idVolume');
            VolumeDrawer.draw(Math.round(MusicPlayer.player.volume*10))
        },
        // Init button actions to control local player
        initActions:function(){
            $('.play',this.div).bind('click',()=>ActivePlaylist.getReal().play());
            $('.pause',this.div).bind('click',()=>ActivePlaylist.getReal().pause());
            $('.next',this.div).bind('click',()=>ActivePlaylist.getReal().next());
            $('.previous',this.div).bind('click',()=>ActivePlaylist.getReal().previous());
            $('.volume-plus',this.div).bind('click',()=>ActivePlaylist.getReal().volumeUp());
            $('.volume-minus',this.div).bind('click',()=>ActivePlaylist.getReal().volumeDown());
            $(document).bind('volume_up',()=>ActivePlaylist.getReal().volumeUp());
            $(document).bind('volume_down',()=>ActivePlaylist.getReal().volumeDown());
            $(document).bind('focus-panel',(e,id)=>{
                if(id.indexOf('idRemotePlaylist') == 0){

                    $('.local','#player').hide();
                    $('.remote','#player').show();
                }else{
                    $('.local','#player').show();
                    $('.remote','#player').hide();
                }
            });
        },
        setShareManager:function(manager){
            this.shareManager = manager;
        },
        setTitle:function(music){
            if(this.quizzMode){
                $('.title',this.div).text('#### QUIZZ n°' + (PlaylistPanel.current+1) + ' ####');
            }else {
                $('.title', this.div).text(music.title);
            }
            $('title').html(music.artist + " - " + music.title);
        },
        setMax:function(value){
            this.seeker.slider('option','max',value)
            $('.duration',this.div).text(MusicPlayer._formatTime(value));
        } ,
        update:function(value){
            // Check value, if max == value, launch next song
            this.seeker.slider('option','value',value)
            var ftime = MusicPlayer._formatTime(value);
            $('.position',this.div).text(ftime);
            $(document).trigger('update_time',ftime);
        }
    },
    volume:{
        step:0.05,
        set:function(value){
            MusicPlayer.player.volume=value/100;
            this.draw();
        },
        draw:function(){
            VolumeDrawer.draw(Math.round(MusicPlayer.player.volume*10))
        },
        up:function(){
            if(MusicPlayer.player.volume >= 1-this.step){
                MusicPlayer.player.volume=1;
            }else{
                MusicPlayer.player.volume+=this.step;
            }
            this.draw();
            ajax({url:basename + 'volume?volume=up'});
            // Share event
            MusicPlayer.controls.shareManager.event('volume',Math.round(MusicPlayer.player.volume*100));
        },
        down:function(){
            if(MusicPlayer.player.volume <= this.step){
                MusicPlayer.player.volume=0;
            }else{
                MusicPlayer.player.volume-=this.step;
            }
            this.draw();
            ajax({url:basename + 'volume?volume=down'});
            MusicPlayer.controls.shareManager.event('volume',Math.round(MusicPlayer.player.volume*100));
        }
    },
    init:function(){
        this.player = $('#idPlayer').get(0);
        this.controls.init('player')
        this.player.addEventListener('canplay',()=>MusicPlayer.initMusic());
        this.player.addEventListener('error',()=>Logger.error("Error when loading music"));
        this.player.addEventListener('timeupdate',()=>MusicPlayer.controls.update(MusicPlayer.player.currentTime));
        this.player.addEventListener('ended',()=>PlaylistPanel.next());
        // Detect key controls
        $(document).bind('keyup',function(e){
            if(PlaylistPanel.searching){return;}
            var key = (e.keyCode !== 0)?e.keyCode:e.charCode;
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
        $(document).unbind('pause_event.player').bind('pause_event.player',function(){
            if(ActivePlaylist.get().isPaused){
                ActivePlaylist.get().play();
            }else{
                ActivePlaylist.get().pause();
            }
        });
        this.device.init();
        // Get nb musics
        ajax({url:basename + 'nbMusics',success:data=>$('#nbMusics').html(data)});
    },
    load:function(music){
        if(music == null){return;}
        Radio.stopRadio();
        Logger.info("Load " + music.src)
        this.player.src = basename + music.src;
        this.controls.setTitle(music);
        this.controls.setMax(music.length);
        if(MusicProgressBar!=null){
            MusicProgressBar.load(parseInt(music.length));
        }
        this.play();
    },
    // load a music by url, like a radio
    loadUrl:function(url,title){
        Logger.info("Load URL " + url + " (" + title + ")");
        this.player.src = url;
        this.controls.setTitle(title);
        this.play();
    },
    stop:function(){
        this.player.src = "";
        this.pause();
    },
    _showPlaying:function(play){
        if(play){
            $('.play',this.controls.div).hide();
            $('.pause',this.controls.div).show();
        }else{
            $('.pause',this.controls.div).hide();
            $('.play',this.controls.div).show();
        }
        $(document).trigger('playing_event');
    },
    pause:function(){
        this.player.pause();
        this._showPlaying(false);
    },
    isPause:function(){
        return !$('.pause',this.div).is(':visible');
    },
    play:function(){
        if(this.player.src == "" && this.playlist != null){
            // Try to load selected playlist music or first
            this.load(this.playlist.getOne());
            return ActivePlaylist.get().shareCurrent();
        }
        this.player.play();
        this._showPlaying(true);
    },
    // launch after load
    initMusic:function(){
        //this.controls.setMax(this.player.duration);
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

let VolumeDrawer = {
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
        for(let i = 0 ; i < 10 ; i++){
            if(i > pourcent-1){
                this.canvas.fillStyle = '#c6c6c6'
            }
            this.canvas.rotate(this.step);
            this.canvas.fillRect(0,4,2,8)
        }
        this.canvas.restore()
    }
}

