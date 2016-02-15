// Use to remote control playlist

// when share, create a connexion to server
var Share = {
    original:null,
    enable:function(){
        $('.share-button').addClass('active');
        // Get unique share id from server
        this.original = CreateOriginal(PlaylistPanel);
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
     sse.addEventListener('clone',function(data){
          manager.disable();
      });
     sse.addEventListener('playlist',function(data){
         remotePlaylist.addMusicsFromIds(JSON.parse(data.data),true);
     });
     sse.addEventListener('remove',function(data){remotePlaylist.removeMusicId(data.data,true);});
     sse.addEventListener('remove',function(data){remotePlaylist.removeMusicId(data.data,true);});
     sse.addEventListener('playMusic',function(data){remotePlaylist.showMusic(data.data);});
     sse.addEventListener('next',function(){remotePlaylist.next();});
     sse.addEventListener('previous',function(){remotePlaylist.previous();});
     sse.addEventListener('pause',function(){remotePlaylist.pause();});
     sse.addEventListener('play',function(){remotePlaylist.play();});

    manager.sse = sse;
    manager.event = function(event,data){
        data = data || "";
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
         $.ajax({
             url:'/shareUpdate',
             data:{id:manager.id,event:'playlist',data:JSON.stringify(ids)}
         });
     });
     sse.addEventListener('remove',function(data){playlist.removeMusicId(data.data);});
     sse.addEventListener('playMusic',function(data){playlist.playMusic(data.data);});
     sse.addEventListener('next',function(){playlist.next();});
     sse.addEventListener('previous',function(){playlist.previous();});
     sse.addEventListener('pause',function(){MusicPlayer.pause();});
     sse.addEventListener('play',function(){MusicPlayer.play();});
    manager.sse = sse;

    manager.event = function(event,data){
        data = data || "";
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