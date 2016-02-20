/* Show tasks from server */

if(Loader){Loader.toLoad("html/big_player.html","BigPlayerPanel");}

var BigPlayerPanel = {
    div:null,
    playerDiv:null,
    time:800,
    defaultTime:800,
    lastBegin:0,
    ease:'',
    visible:false,
    timeField : null,
    currentSong:0,
    init:function(){
        this.div = $('#idBigPlayer');
        this.playerDiv = $('.big-player',this.div);
        this.timeField = $('.field-time',this.div);
    },
    _showPlaying:function(){
        if(MusicPlayer.isPause()){
            this._showPause();
        }else{
            this._showPlay();
        }
    },
    open:function(){
        if(!PlaylistPanel.isVisible()){
            return;
        }
        this.visible = true;
        var musics = PlaylistPanel.getNSongs(3);
        this.currentSong = PlaylistPanel.current != -1 ? PlaylistPanel.current : 0;
        for(var pos in musics){
            $('.posN' + pos + ' .title-song').html(musics[pos].title);
            $('.posN' + pos + ' .album-song').html(musics[pos].artist + " - " + musics[pos].album);
        }
        this._showPlaying();
        var _self = this;
        $(document).unbind('playing_event.big').bind('playing_event.big',function(){
           _self._showPlaying();
        });
        $(document).unbind('show_next_event.big').bind('show_next_event.big',function(){
           _self.next(true);
        });
        $(document).unbind('show_previous_event.big').bind('show_previous_event.big',function(){
            _self.previous(true);
        });
        $(document).unbind('update_time.big').bind('update_time.big',function(e,ftime){
            _self.timeField.html(ftime);
        });
        this.div.show();
    },
    close:function(){
        this.div.hide();
    },
    _showPause:function(){
        $('.pause-button',this.playerDiv).show();
        $('.play-button',this.playerDiv).hide();
    },
    _showPlay:function(){
        $('.pause-button',this.playerDiv).hide();
        $('.play-button',this.playerDiv).show();
    },
    play:function(){
        MusicPlayer.play();
        this._showPause();
    },
    pause:function(){
        MusicPlayer.pause();
        this._showPlay();
    },
    previous:function(noAction){
        var _self = this;
        QueueEffectManager.add(function(time,manager){_self._doPrevious(noAction,time,manager)});
    },
    _doPrevious:function(noAction,time,manager){
        if(this.currentSong<=0){return;}
        this._showPause();
        $('.posN-1 .album-song',this.playerDiv).switchClass('hide-album','show-album',time,this.ease);
        $('.posN-1 .title-song',this.playerDiv).switchClass('','style-song',time,this.ease);
        $('.posN0 .album-song',this.playerDiv).switchClass('show-album','hide-album',time,this.ease);
        $('.posN0 .title-song',this.playerDiv).switchClass('style-song','',time,this.ease);
        $('.posN-1',this.playerDiv).switchClass('posN-1 level1','posN0',time,this.ease);
        $('.posN-2',this.playerDiv).switchClass('posN-2 level2','posN-1 level1',time,this.ease);
        $('.posN-3',this.playerDiv).switchClass('posN-3','posN-2 level2',time,this.ease);
        $('.posN3',this.playerDiv).removeClass('posN3').addClass('posN-3 temp');
        $('.posN1',this.playerDiv).switchClass('posN1 level1','posN2 level2',time,this.ease);
        $('.posN2',this.playerDiv).switchClass('posN2 level2','posN3',time,this.ease);
        $('.posN0',this.playerDiv).switchClass('posN0','posN1 level1',time,this.ease);
        manager.setBegin(new Date().getTime(),time);

        if(noAction == null || noAction == false){
            PlaylistPanel.previous();
        }
        var music = PlaylistPanel.getSong(-3,--this.currentSong);
        if(music != null){
            $('.posN-3.temp .title-song').html(music.title);
            $('.posN-3.temp .album-song').html(music.artist + " - " + music.album);
            $('.posN-3.temp').removeClass('temp');
        }
    },
    next:function(noAction){
        var _self = this;
        QueueEffectManager.add(function(time,manager){_self._doNext(noAction,time,manager)});
    },
    _doNext:function(noAction,time,manager){

        this._showPause();
        $('.posN1 .album-song',this.playerDiv).switchClass('hide-album','show-album',time,this.ease);
        $('.posN1 .title-song',this.playerDiv).switchClass('','style-song',time,this.ease);
        $('.posN0 .album-song',this.playerDiv).switchClass('show-album','hide-album',time,this.ease);
        $('.posN0 .title-song',this.playerDiv).switchClass('style-song','',time,this.ease);
        $('.posN3',this.playerDiv).switchClass('posN3','posN2 level2',time,this.ease);
        $('.posN2',this.playerDiv).switchClass('posN2 level2','posN1 level1',time,this.ease);
        $('.posN1',this.playerDiv).switchClass('posN1 level1','posN0',time,this.ease);
        $('.posN0',this.playerDiv).switchClass('posN0','posN-1 level1',time,this.ease);
        $('.posN-3',this.playerDiv).removeClass('posN-3').addClass('posN3 temp');
        $('.posN-1',this.playerDiv).switchClass('posN-1 level1','posN-2 level2',time,this.ease);
        $('.posN-2',this.playerDiv).switchClass('posN-2 level2','posN-3',time,this.ease);
        manager.setBegin(new Date().getTime(),time);
        if(noAction == null || noAction == false){
            PlaylistPanel.next();
        }
        var music = PlaylistPanel.getSong(3,++this.currentSong);
        if(music != null){
            $('.posN3.temp .title-song').html(music.title);
            $('.posN3.temp .album-song').html(music.artist + " - " + music.album);
            $('.posN3.temp').removeClass('temp');
        }
    }
}

var QueueEffectManager = {
    queue:[],
    current:null,
    timeEffect:800,
    add:function(fct){
        this.queue.push(fct);
        this.execute();
    },
    // time is a long of effect
    // begin is computed after all effects launch
    setBegin:function(begin,time){
        this.current = {begin:begin,time:time};
    },
    execute:function(){
        if(this.current != null){
            var elapse = this.current.time - (new Date().getTime() - this.current.begin);
            if(elapse > 0){
                setTimeout(function(){QueueEffectManager.execute();},elapse+50);
                return;
            }
            this.current = null;
        }
        var timeEffect = this.queue.length > 1 ? this.timeEffect / 4 : this.timeEffect;
        var effect = this.queue.splice(0,1)[0];
        // launch effect
        effect(timeEffect,this);
    }
}
