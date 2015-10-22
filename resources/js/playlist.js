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
           PlaylistPanel.saveCurrent();
        });
        $(document).unbind('delete_event').bind('delete_event',function(){
            // Delete music. Find position element in list
            var nb = $('> div:not(.head)',PlaylistPanel.listDiv).length
            var afters = $('.focused:visible~div',PlaylistPanel.listDiv).length
            PlaylistPanel.removeMusic(nb-afters);
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
                // Get info from id music
                PlaylistPanel.addMusicFromId(idMusic);
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
    // Return selected or first in list
    getOne:function(){
        var focused = $('div:not(.head).focused',this.listDiv);
        if(focused.length > 0){
            return focused.data('music');
        }
        if(this.list.length > 0){
            return this.list[0];
        }
        return null;
    },
    load:function(){
        if(localStorage && localStorage["playlist"]!=null){
            var musics = JSON.parse(localStorage["playlist"]);
            musics.forEach(function(m){
                PlaylistPanel.add(m,true);
            });
            this.current = parseInt(localStorage["current"]) || -1
            this._selectLine();
            this.open();
        }
    },
    // Save current playlist and current music in localstorage
    save:function(){
        if(localStorage){
            localStorage["playlist"] = JSON.stringify(this.list);
        }
    },
    saveCurrent:function(){
        if(localStorage){
            localStorage["current"] = this.current;
        }
    },
    removeMusic:function(index){
        $('>div:nth-child(' + (index+1) + ')',this.listDiv).remove();
        this.list.splice(index,1);
        // Play next song ?
        this.save()
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
    add:function(music,noSave){
        var position = $('div',this.listDiv).length;
        var line = $('<div><span>' + position + '</span><span>' + music.title + '</span><span>' + MusicPlayer._formatTime(music.length) + '</span></div>');
        line.data("position",position-1);
        line.data("music",music);
        this.listDiv.append(line);
        this.list.push(music);
        if(noSave == null || noSave == false){
            this.save();
        }
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