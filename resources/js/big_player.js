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
        PlaylistPanel.getOne();
        this.visible = true;
        var musics = PlaylistPanel.getNSongs(3);
        this.currentSong = PlaylistPanel.current;
        $('.title-song,.album-song').empty();
        for(var pos in musics){
            this._loadCoverAt(musics[pos],$('.posN' + pos));
            $('.posN' + pos + ' .title-song').html(musics[pos].title);
            $('.posN' + pos + ' .album-song').html(musics[pos].artist + " - " + musics[pos].album);
        }
        this._loadCover(musics[0]);
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
    _loadCover:function(music){
        if(music == null || music.cover == ""){return;}
        $('.posN0 .cover-album').html('<img src="' + music.cover + '"/>');
      /*CoverHelper.get(music.artist,music.album,function(src){
          $('.posN0 .cover-album').html('<img src="' + src + '"/>');
      }); */
    },
    _loadCoverAt:function(music,div){
        if(music == null || music.cover == ""){return;}
        $('.cover-album',div).html('<img src="' + music.cover + '"/>');
      /*CoverHelper.get(music.artist,music.album,function(src){
          $('.cover-album',div).html('<img src="' + src + '"/>');
      });*/
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
        $('.posN3',this.playerDiv).removeClass('posN3').addClass('posN-3 temp').find('div').empty();
        $('.posN1',this.playerDiv).switchClass('posN1 level1','posN2 level2',time,this.ease);
        $('.posN2',this.playerDiv).switchClass('posN2 level2','posN3',time,this.ease);
        $('.posN0',this.playerDiv).switchClass('posN0','posN1 level1',time,this.ease);
        manager.setBegin(new Date().getTime(),time);

        if(noAction == null || noAction == false){
            PlaylistPanel.previous();
        }
        var music = PlaylistPanel.getSong(-3,--this.currentSong);
        if(music != null){
            this._loadCoverAt(music,$('.posN-3.temp'));
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
        if(this.currentSong>=PlaylistPanel.list.length -1){return;}
        this._showPause();
        $('.posN1 .album-song',this.playerDiv).switchClass('hide-album','show-album',time,this.ease);
        $('.posN1 .title-song',this.playerDiv).switchClass('','style-song',time,this.ease);
        $('.posN0 .album-song',this.playerDiv).switchClass('show-album','hide-album',time,this.ease);
        $('.posN0 .title-song',this.playerDiv).switchClass('style-song','',time,this.ease);
        $('.posN3',this.playerDiv).switchClass('posN3','posN2 level2',time,this.ease);
        $('.posN2',this.playerDiv).switchClass('posN2 level2','posN1 level1',time,this.ease);
        $('.posN1',this.playerDiv).switchClass('posN1 level1','posN0',time,this.ease);
        $('.posN0',this.playerDiv).switchClass('posN0','posN-1 level1',time,this.ease);
        $('.posN-3',this.playerDiv).removeClass('posN-3').addClass('posN3 temp').find('div').empty();
        $('.posN-1',this.playerDiv).switchClass('posN-1 level1','posN-2 level2',time,this.ease);
        $('.posN-2',this.playerDiv).switchClass('posN-2 level2','posN-3',time,this.ease);
        manager.setBegin(new Date().getTime(),time);
        if(noAction == null || noAction == false){
            PlaylistPanel.next();
        }
        var music = PlaylistPanel.getSong(3,++this.currentSong);
        if(music != null){
            this._loadCoverAt(music,$('.posN3.temp'));
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


var CoverHelper = {
    get:function(artist,album,callback,useOnlyArtist,tries){
        var params = encodeURIComponent(((!useOnlyArtist)?'release:"' + album + '" AND ':'') + 'artist:"' + artist + '"');
        $.ajax(
            {
                url:'http://musicbrainz.org/ws/2/release/?query=' + params,
                dataType:'xml',
                success:function(data){
                    var results = $(data).find('metadata release');
                    if(results.length > 0){
                        var r = results.get(0);
                        // Get release group if exist
                        if(r.getElementsByTagName('release-group').length > 0){
                            var id = r.getElementsByTagName('release-group')[0].getAttribute('id');
                            callback(CoverHelper._getReleaseGroupUrlCover(id));
                        }else{
                            var id = r.getAttribute("id");
                            callback(CoverHelper._getReleaseUrlCover(id));
                        }
                    }else{
                        // Keep the first word
                        if(album.indexOf(" ")!=-1){
                            album = album.substr(0,album.indexOf(" "));
                            CoverHelper.get(artist,album,callback);
                        }else{
                            if(!useOnlyArtist){
                                CoverHelper.get(artist,album,callback,true);
                            }
                        }
                    }
                },
                error:function(){
                    if(tries!=null && tries<=0){return;}
                    console.log("error with server, retry",artist,album);
                    setTimeout(function(){CoverHelper.get(artist,album,callback,useOnlyArtist,tries!=null ? --tries:3);},2000);
                }
            }
        );
    },
    _getReleaseUrlCover:function(id){
        return "http://coverartarchive.org/release/" + id + "/front-250";
    },
    _getReleaseGroupUrlCover:function(id){
          return "http://coverartarchive.org/release-group/" + id + "/front-250";
      }
}
