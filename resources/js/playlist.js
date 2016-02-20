/* Show tasks from server */

if(Loader){Loader.toLoad("html/playlist.html","PlaylistPanel");}

var PlaylistPanel = {
    listDiv:null,
    list:[],
    current:-1,
    saveInLS:true,
    shareManager:null,
    // To avoid load music
    noLoadMusic:false,
    init:function(idDiv,title,noLoad){
        idDiv = idDiv || '#idPlaylist';
        title = title || 'Playlist'
        $.extend(true,this,Panel)
        this.initPanel($(idDiv),'<span class="glyphicon glyphicon-music icon"></span>' + title,true)
        this.div.resizable()
        this.listDiv = $('.playlist',this.div);
        // Select behaviour
        var _self = this;
        this.listDiv.on("click",'div',function(){
            $('div',_self.listDiv).removeClass('focused');
            $(this).addClass('focused');
        });
        this.listDiv.on("dblclick",'div',function(e){
           window.getSelection().removeAllRanges();
           if(!_self.noLoadMusic){
                MusicPlayer.load($(this).data("music"));
           }
           if(_self.shareManager!=null){
                _self.shareManager.event('playMusic',$(this).data("music").id);
           }
           _self.setActualPlayed($(this));
           _self.saveCurrent();
        });
        $(document).unbind('delete_event').bind('delete_event',function(){
            // Delete music. Find position element in list
            _self.removeMusic(_self.getFocusedPosition());
        });
        $(document).unbind('next_event').bind('next_event',function(){
            _self.next();
            $(document).trigger('show_next_event');
        });
        $(document).unbind('previous_event').bind('previous_event',function(){
            _self.previous();
            $(document).trigger('show_previous_event');
        });
        this.listDiv.droppable({
            drop:function(event,ui){
                var idMusic = ui.draggable.data('id');
                if(idMusic == null){
                    // Get url if exist to load all, folder case
                    var url = ui.draggable.data('url_drag');
                     if(url != null){
                        _self.addMusicsFromUrl(url);
                     }
                }else{
                    // Get info from id music
                    _self.addMusicFromId(idMusic);
                }
            }
        });
        this.div.bind('close',function(){
           _self.cleanPlaylist();
           // If share, close it
           if(_self.shareManager!=null){
            _self.shareManager.disable();
           }
        });
        // Load saved playlist
        if(noLoad == true){
            this.saveInLS = false;
        }
        this.load();
    },
    // return songs around current
    getSong:function(shift){
        if(this.current != -1 && this.current + shift >=0 && this.current+shift <this.list.length){
            return this.list[this.current+shift];
        }
            return null;
    },
    getSong:function(shift,from){
        from = from || this.current;
        if(from != -1 && from + shift >=0 && from+shift <this.list.length){
            return this.list[from+shift];
        }
        return null;
    },
    getNSongs:function(n){
        var musics = {};
        if(this.list == null || this.list.length == 0){
            return musics;
        }
        var c = this.current !=-1 ? this.current : 0;
        musics[0] = this.list[c];
        for(var i = c ; i <= c+n && i < this.list.length ; i++ ){
            musics[i-c] = this.list[i];
        }
        for(var i = c ; i >= c-n && i >=0 ; i-- ){
            musics[i-c] = this.list[i];
        }
        return musics;
    },
    cleanPlaylist:function(){
        if(this.saveInLS && localStorage){
            delete(localStorage["playlist"]);
            delete(localStorage["current"]);
        }
        this.listDiv.empty();
        this.list = [];
        this.updateTotal();
    },
    // No need to propagate share
    playMusic:function(id){
        this.list.forEach(function(music,i){
           if(music.id == id){
                var line = $('>div:nth-child(' + (i+1) + ')',this.listDiv);
                MusicPlayer.load(music);
                this.setActualPlayed(line);
                this.saveCurrent();
                return;
           }
        },this);

    },
    setActualPlayed:function(line){
        $('div',this.listDiv).removeClass('played selected focused');
        line.addClass('played');
        this.current = this.getPlayedPosition();
        this.listDiv.scrollTop(line.position().top);
    } ,
    getPlayedPosition:function(){
        var nb = $('> div',this.listDiv).length
        var afters = $('.played~div',this.listDiv).length
        return nb-afters-1;
    },
    getFocusedPosition:function(){
        var nb = $('> div',this.listDiv).length;
        var afters = $('.focused~div',this.listDiv).length
        return nb-afters;
    },
    // Return selected or first in list
    getOne:function(){
        var focused = $('div.focused',this.listDiv);
        if(focused.length > 0){
            this.current = this.getFocusedPosition();
            this._selectLine();
            return focused.data('music');
        }
        focused = $('div.played',this.listDiv);
        if(focused.length > 0){
            this.current = this.getPlayedPosition();
            this._selectLine();
            return focused.data('music');
        }
        if(this.list.length > 0){
            this.current = 0;
            this._selectLine();
            return this.list[0];
        }
        return null;
    },
    load:function(){
        if(this.saveInLS && localStorage && localStorage["playlist"]!=null){
            var musics = JSON.parse(localStorage["playlist"]);
            var currentMusic = localStorage["current"];
            musics.forEach(function(m,i){
                this.add(m,true);
                if(currentMusic == m.id){
                    this.current = i;
                }
            },this);
            this._selectLine();
            this.open();
        }
    },
    // Save current playlist and current music in localstorage
    save:function(){
        if(this.saveInLS && localStorage){
            localStorage["playlist"] = JSON.stringify(this.list);
        }
    },
    updateTotal:function(){
        $('.total_playlist',this.div).html(this.list.length);
    },
    saveCurrent:function(){
        if(this.saveInLS && localStorage){
            localStorage["current"] = this.list[this.current].id;
        }
    },
    removeMusicId:function(id,noShare){
        this.list.forEach(function(music,i){
           if(music.id == id){
                return this.removeMusic(i,noShare);
           }
        },this);
    },
    removeMusic:function(index,noShare){
        $('>div:nth-child(' + index + ')',this.listDiv).remove();
        var music = this.list.splice(index-1,1)[0];
        this.save();
        this.updateTotal();
        if((noShare == null || noShare == false) && this.shareManager!=null){
            this.shareManager.event("remove",music.id);
        }
    },
    // Add many musics from list of id
    addMusicsFromIds:function(datas,noShare){
        var ids = datas.ids;
        this.current = datas.current || -1;
        var _self = this;
        $.ajax({
            url:'/musicsInfo?ids=' + JSON.stringify(ids),
            dataType:'json',
            success:function(data){
                data.forEach(function(music){
                    if((noShare == null || noShare == false) && _self.shareManager!=null){
                        _self.shareManager.event('add',music.id);
                    }
                    _self.add(music,true,true);
                });
                _self.save();
                _self.updateTotal();
                _self._selectLine();
            }
        });
    },
    addMusicFromId:function(id,noShare){
        if(noShare == null || noShare == false){
            if(this.shareManager!=null){
                this.shareManager.event('add',id);
            }
        }
        var _self = this;
        $.ajax({
            url:'/musicInfo?id=' + id,
            dataType:'json',
            success:function(data){
                // No need to create a real Music, just a container with properties, no methods
                _self.add(data);
            }
        })
    },
    addMusicsFromUrl:function(url){
        var _self = this;
        $.ajax({
            url:url,
            dataType:'json',
            success:function(data){
                // If list receive with id in each element, add music
                var ids = data.filter(function(m){return m.id != null;}).map(function(m){return parseInt(m.id);})
                _self.addMusicsFromIds({ids:ids});
            }
        })
    },
    // Add a new music in list
    add:function(music,noSave,noTotal){
        var line = $('<div></div>');
        line.append('<span class="glyphicon glyphicon-remove remove" title="Remove"></span>');
        //line.append('<span>' + $('>div',this.listDiv).length + '</span>');
        line.append('<span>' + music.title + '</span>');
        line.append('<span>' + MusicPlayer._formatTime(music.length) + '</span>');
        line.append('<span class="glyphicon glyphicon-play" title="Play"></span>');
        var _self = this;
        $('.glyphicon-play',line).bind('click',function(){
            _self.setActualPlayed($(this).closest('div'));
            if(_self.shareManager!=null){
                _self.shareManager.event('playMusic',music.id);
            }
            MusicPlayer.load(music);
        });
        $('.glyphicon-remove',line).bind('click',function(){
            var nb = _self.listDiv.find('div').length;
            var pos = nb - $(this).parent().find('~div').length;
            _self.removeMusic(pos);
        });
        line.data("music",music);
        this.listDiv.append(line);
        this.list.push(music);
        if(noSave == null || noSave == false){
            this.save();
        }
        if(noTotal == null || noTotal == false){
            this.updateTotal();
        }
    },
    _selectLine:function(){
        if(this.current == -1){
            return
        }
        var line = $('div:nth-child(' + (this.current+1) + ')',this.listDiv);
        this.setActualPlayed(line);
    },
    next:function(){
        if(this.current+1>=this.list.length){
            return;
        }
        this.current++;
        this._selectLine();
        if(this.shareManager!=null){
            this.shareManager.event('next');
        }
        if(!this.noLoadMusic){
            MusicPlayer.load(this.list[this.current]);
        }
    },
    previous:function(){
        if(this.current<=0){
            return;
        }
        this.current--;
        this._selectLine();
        if(this.shareManager!=null){
            this.shareManager.event('previous');
        }
        if(!this.noLoadMusic){
            MusicPlayer.load(this.list[this.current]);
        }
    },
    setShareManager:function(manager){
        this.shareManager = manager;
    }
}

var RemotePlaylist = {
    init2:function(){
        // If remote player, shareManager exist for sure
        var _self = this;
        // Manage remote buton
        $('.controls>.glyphicon-fast-backward',this.div).bind('click',function(){
            _self.previous();
            _self.shareManager.event("previous");
        });
        $('.controls>.glyphicon-fast-forward',this.div).bind('click',function(){
            _self.next();
            _self.shareManager.event("next");
        });
        $('.controls>.glyphicon-play',this.div).bind('click',function(){
            _self.play();
            _self.shareManager.event("play");
        });
        $('.controls>.glyphicon-pause',this.div).bind('click',function(){
            _self.pause();
            _self.shareManager.event("pause");
        });
        this.listDiv.on("dblclick",'div',function(e){
            _self.play();
        });
        this.div.bind('close',function(){
            if(_self.shareManager != null){
                _self.shareManager.disable(true);
            }
        });
    },
    // show played music
    showMusic:function(id){
        this.list.forEach(function(music,i){
           if(music.id == id){
                this.play();
                var line = $('>div:nth-child(' + (i+1) + ')',this.listDiv);
                this.setActualPlayed(line);
                return;
           }
        },this);
    },
    pause:function(){
        $('.controls>.glyphicon-play',this.div).show();
        $('.controls>.glyphicon-pause',this.div).hide();
    },
    play:function(){
        $('.controls>.glyphicon-play',this.div).hide();
        $('.controls>.glyphicon-pause',this.div).show();
    }
}

function connectToShare(){
    $.extend(RemotePlaylist,PlaylistPanel);
    RemotePlaylist.noLoadMusic=true;
    RemotePlaylist.list = [];
    RemotePlaylist.init('#idRemotePlaylist','Remote playlist',true);
    RemotePlaylist.init2();

    // Show shares
    Share.getShares(function(data){
        data.forEach(function(s){
            var share = $('<div> => ' + s.Name + ' : <button>Connect</button></div>');
            $('button',share).bind('click',function(){
                RemotePlaylist.listDiv.empty();
                $('.title > span:first',RemotePlaylist.div).html('Remote playlist for ' + s.Name);
                CreateClone(s.Id,RemotePlaylist);
            });
            RemotePlaylist.listDiv.append(share);
            RemotePlaylist.open();
        });
    });
}
