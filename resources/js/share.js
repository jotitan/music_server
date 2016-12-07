// Use to remote control playlist

// when share, create a connexion to server
var Share = {
    original:null,
    enable:function(){
        if(MusicPlayer.device.name == "No name"){
            alert("Define a real name for device");
            return;
        }
        $('.share-button').addClass('active');
        // Get unique share id from server
        this.original = CreateOriginal(PlaylistPanel);
        PlaylistPanel.open();
    },
    disable:function(){
        this.original.disable();
        this.original = null;
        $('.share-button').removeClass('active');
    },
    getShares:function(callback){
        $.ajax({
           url:'/shares',
           dataType:'json',
           success:function(data){
                callback(data);
           }
        });
    },
    init:function(){
        // Manage original share
        $('.share-button').bind('click',function(){
            if(Share.original!=null){
                Share.disable();
            }else{
                Share.enable();
            }
        });
    }
}

// Create a clone manager (for a specific share id)
function CreateClone(id,remotePlaylist){
     // Add behaviour on remotePlaylist, receive event for remoteplaylist

    var manager = {id:id};
    var sse = new EventSource('/share?id=' + id + '&device=' + MusicPlayer.device.name);

     // Set receive events behaviour
     sse.addEventListener('add',function(data){
         remotePlaylist.addMusicFromId(data.data,true);
     });
     sse.addEventListener('close',function(data){
          manager.disable();
      });
     sse.addEventListener('playlist',function(data){
         var info = JSON.parse(data.data);
         remotePlaylist.cleanPlaylist();
         remotePlaylist.addMusicsFromIds(info,true);
         info.playing ? remotePlaylist.play() : remotePlaylist.pause();
     });
     sse.addEventListener('remove',function(data){remotePlaylist.removeMusicId(data.data,true);});
     sse.addEventListener('playMusic',function(data){remotePlaylist.showMusicByPosition(data.data);});
     sse.addEventListener('next',function(){remotePlaylist.next(true);});
     sse.addEventListener('previous',function(){remotePlaylist.previous(true);});
     sse.addEventListener('pause',function(){remotePlaylist.pause();});
     sse.addEventListener('play',function(){remotePlaylist.play();});

    manager.sse = sse;
    manager.event = function(event,data){
        data = data == null ? "" : data;
        $.ajax({
            url:'/shareUpdate',
            data:{id:this.id,event:event,data:data}
        });
    };
    manager.disable = function(noclose){
        this.sse.close();
        this.event('close');
        if(noclose == null ||noclose == false){
            remotePlaylist.close();
        }
    }
    remotePlaylist.setShareManager(manager);
}

// Create a original manager (for a specific share id)
function CreateOriginal(playlist){
    var manager = {};
    var sse = new EventSource('/share?device=' + MusicPlayer.device.name);

     // Set receive events behaviour
     sse.addEventListener('id',function(data){
         manager.id = parseInt(data.data);
     });
     sse.addEventListener('add',function(data){
          playlist.addMusicFromId(data.data,true);
      });
     sse.addEventListener('playlist',function(data){
         playlist.addMusicsFromIds(JSON.parse(data.data),true);
     });
     sse.addEventListener('askPlaylist',function(data){
         var ids = playlist.list.map(function(m){return parseInt(m.id)});
         var data = {ids:ids,current:playlist.current,playing:!MusicPlayer.isPause()};
         $.ajax({
             url:'/shareUpdate',
             data:{id:manager.id,event:'playlist',data:JSON.stringify(data)}
         });
     });
     sse.addEventListener('remove',function(data){playlist.removeMusicId(data.data);});
     sse.addEventListener('playMusic',function(data){playlist.playMusic(data.data);});
     sse.addEventListener('next',function(){playlist.next(true);});
     sse.addEventListener('previous',function(){playlist.previous(true);});
     sse.addEventListener('pause',function(){MusicPlayer.pause();});
     sse.addEventListener('play',function(){MusicPlayer.play();});
    manager.sse = sse;

    manager.event = function(event,data){
        data = data == null ? "" : data;
        $.ajax({
            url:'/shareUpdate',
            data:{id:this.id,event:event,data:data}
        });
    };
    manager.disable = function(){
        MusicPlayer.controls.setShareManager(null);
        playlist.setShareManager(null);
        this.sse.close();
        this.event('close');
    };
    MusicPlayer.controls.setShareManager(manager);
    playlist.setShareManager(manager);
    return manager;
}

// Create a simple and light remote control (only play, pause, next and previous)
function CreateRemote(id,target){
    if(target == null){
        return null;
    }
    var manager = {id:id};
    var sse = new EventSource('/share?id=' + id + '&device=' + MusicPlayer.device.name);

    sse.addEventListener('close',function(){
        manager.disable();
    });
    sse.addEventListener('load',function(response){
        manager.loadMusic(response.data);
        target.setIsPlaying(true);
    });

    sse.addEventListener('pause',function(response){
        target.setIsPlaying(false);
    });
    sse.addEventListener('play',function(response){
        target.setIsPlaying(true);
    });

    sse.addEventListener('playlist',function(response){
        var data = JSON.parse(response.data);
        if(data.current!=null){
            var idMusic = data.ids[data.current];
            manager.loadMusic(idMusic);
            target.setIsPlaying(data.playing);
        }
    });

    manager.sse = sse;

    manager.loadMusic = function(id){
        $.ajax({
            url:'/musicInfo?id=' + id,
            success:function(data){
                target.updateMusic(JSON.parse(data));
            }
        })
    }
    manager.event = function(event,data){
        data = data == null ? "" : data;
        $.ajax({
            url:'/shareUpdate',
            data:{id:this.id,event:event,data:data}
        });
    };
    manager.disable = function(noclose){
        this.sse.close();
        this.event('close');
        if(noclose == null ||noclose == false){
            remotePlaylist.close();
        }
    }
    return manager;
}

var RemoteControlManager = {
    manager:null,
    div:null,
    divSelect:null,
    url:'',
    init:function(idDiv,idSelect,musicUrl){
        var _self = this;
        this.url = musicUrl || '';
        this.div = $('#' + idDiv);
        this.divSelect = $('#' + idSelect);
        $('.play',this.div).bind('click',function(){
            _self.manager.event('play');
            _self.setIsPlaying(true);
        });

        $('.pause',this.div).bind('click',function(){
            _self.manager.event('pause');
            _self.setIsPlaying(false);
        });

        $('.previous',this.div).bind('click',function() {
            _self.manager.event('previous');
        });

        $('.next',this.div).bind('click',function() {
            _self.manager.event('next');
        });

        this.divSelect.bind('change',function(){
            _self.manager = CreateRemote($(this).val(),_self);
            _self.divSelect.hide();
            _self.div.show();
        });
        Share.getShares(function(data) {
            if(data == null || data.length == 0){
                _self.divSelect.hide();
                return;
            }
            _self.divSelect.empty().append('<option>...</option>');
            data.forEach(function (s) {
                _self.divSelect.append('<option value="' + s.Id + '">' + s.Name + '</option>');
            });
        });
        return this;
    },
    updateMusic:function(music) {
        $('.title',this.div).html(music.title + " - " + music.artist);
    },
    setIsPlaying:function(isPlaying){
        if(isPlaying){
            $('.play',this.div).hide();
            $('.pause',this.div).show();
        }else{
            $('.pause',this.div).hide();
            $('.play',this.div).show();
        }
    }
}