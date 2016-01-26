/* Show tasks from server */

if(Loader){Loader.toLoad("html/playlist.html","PlaylistPanel");}

var PlaylistPanel = {
    listDiv:null,
    list:[],
    current:-1,
    init:function(){
        $.extend(true,this,Panel)
        this.initPanel($('#idPlaylist'),'<span class="glyphicon glyphicon-music icon"></span>Playlist',true)
        this.div.resizable()
        this.listDiv = $('.playlist',this.div)
        // Select behaviour
        this.listDiv.on("click",'div',function(){
            $('div',PlaylistPanel.listDiv).removeClass('focused');
            $(this).addClass('focused');
        });
        this.listDiv.on("dblclick",'div',function(e){
           window.getSelection().removeAllRanges()
           MusicPlayer.load($(this).data("music"));
           PlaylistPanel.setActualPlayed($(this));
           PlaylistPanel.saveCurrent();
        });
        $(document).unbind('delete_event').bind('delete_event',function(){
            // Delete music. Find position element in list
            PlaylistPanel.removeMusic(PlaylistPanel.getFocusedPosition());
        });
        $(document).unbind('next_event').bind('next_event',function(){
            PlaylistPanel.next();
        });
        $(document).unbind('previous_event').bind('previous_event',function(){
            PlaylistPanel.previous();
        });
        this.listDiv.droppable({
            drop:function(event,ui){
                var idMusic = ui.draggable.data('id');
                if(idMusic == null){
                    // Get url if exist to load all
                    var url = ui.draggable.data('url_drag');
                     if(url != null){
                        PlaylistPanel.addMusicsFromUrl(url);
                     }
                }else{
                    // Get info from id music
                    PlaylistPanel.addMusicFromId(idMusic);
                }
            }
        })
        this.div.bind('close',function(){
           if(localStorage){
            delete(localStorage["playlist"])
           }
        });
        // Load saved playlist
        this.load();
    },
    setActualPlayed:function(line){
        $('div',PlaylistPanel.listDiv).removeClass('played selected focused');
        line.addClass('played');
        this.current = this.getPlayedPosition();
        this.listDiv.scrollTop(line.position().top);
    } ,
    getPlayedPosition:function(){
        var nb = $('> div',this.listDiv).length
        var afters = $('.played:visible~div',this.listDiv).length
        return nb-afters-1;
    },
    getFocusedPosition:function(){
        var nb = $('> div',this.listDiv).length;
        var afters = $('.focused:visible~div',this.listDiv).length
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
        if(localStorage && localStorage["playlist"]!=null){
            var musics = JSON.parse(localStorage["playlist"]);
            var currentMusic = localStorage["current"];
            musics.forEach(function(m,i){
                this.add(m,true);
                if(currentMusic == m.id){
                    this.current = i;
                }
            },this);
            this.open();
            this._selectLine();
        }
    },
    // Save current playlist and current music in localstorage
    save:function(){
        if(localStorage){
            localStorage["playlist"] = JSON.stringify(this.list);
        }
    },
    updateTotal:function(){
        $('.total_playlist',this.div).html(this.list.length);
    },
    saveCurrent:function(){
        if(localStorage){
            localStorage["current"] = this.list[this.current].id;
        }
    },
    removeMusic:function(index){
        $('>div:nth-child(' + index + ')',this.listDiv).remove();
        this.list.splice(index-1,1);
        this.save();
        this.updateTotal();
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
    addMusicsFromUrl:function(url){
        $.ajax({
            url:url,
            dataType:'json',
            success:function(data){
                // If list receive with id in each element, add music
                data.forEach(function(music){
                   if(music.id != null){
                        PlaylistPanel.addMusicFromId(music.id);
                   }
                });
            }
        })
    },
    // Add a new music in list
    add:function(music,noSave){
        var line = $('<div></div>');
        line.append('<span class="glyphicon glyphicon-remove remove" title="Remove"></span>');
        //line.append('<span>' + $('>div',this.listDiv).length + '</span>');
        line.append('<span>' + music.title + '</span>');
        line.append('<span>' + MusicPlayer._formatTime(music.length) + '</span>');
        line.append('<span class="glyphicon glyphicon-play" title="Play"></span>');
        $('.glyphicon-play',line).bind('click',function(){
            PlaylistPanel.setActualPlayed($(this).closest('div'));
            MusicPlayer.load(music);
        });
        $('.glyphicon-remove',line).bind('click',function(){
            var nb = PlaylistPanel.listDiv.find('div').length;
            var pos = nb - $(this).parent().find('~div').length;
            PlaylistPanel.removeMusic(pos);
        });
        line.data("music",music);
        this.listDiv.append(line);
        this.list.push(music);
        if(noSave == null || noSave == false){
            this.save();
        }
        this.updateTotal();
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