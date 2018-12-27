// Use to remote control playlist

// when share, create a connexion to server
var Share = {
    original:null,
    emptyManager:{event:function(){},disable:function(){}},
    listShared:[],
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
    addShare:function(id,remotePlaylist){
        this.listShared[id] = remotePlaylist;
    },
    removeShare:function(id){
        delete(this.listShared[id]);
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
        // Detect auto share
        if(location.search.indexOf('autoshare=true')!=-1){
            Share.enable();
        }
    }
}

// Create a clone manager (for a specific share id). receive event from original
function CreateClone(id,remotePlaylist){
     // Add behaviour on remotePlaylist, receive event for remoteplaylist

    var manager = {id:id};
    var sse = new EventSource('/share?id=' + id + '&device=' + MusicPlayer.device.name);

     // Set receive events behaviour
     sse.addEventListener('add',function(data){
         remotePlaylist.addMusicsFromIds({ids:data.data.split(',')},true);
     });
     sse.addEventListener('close',function(data){
          manager.disable();
      });
     sse.addEventListener('playlist',function(data){
         var info = JSON.parse(data.data);
         remotePlaylist._cleanPlaylist(true);
         remotePlaylist.addMusicsFromIds(info,true);
         info.playing ? remotePlaylist._play() : remotePlaylist._pause();         
         remotePlaylist.updateVolume(info.volume);
         if(info.radio != ""){
            remotePlaylist.selectRadio(info.radio);            
         }
     });
     sse.addEventListener('remove',function(data){remotePlaylist._removeMusic(data.data,true);});
     sse.addEventListener('cleanPlaylist',function(data){remotePlaylist._cleanPlaylist(true);});
     // Send position and position to check
     sse.addEventListener('playMusic',function(data){
        var d = JSON.parse(data.data);
        remotePlaylist.showMusicByPosAndId(d.position,d.id);
     });
     // No next or previous, receive only position to play
     sse.addEventListener('pause',function(){remotePlaylist._pause();});
     sse.addEventListener('play',function(){remotePlaylist._play();});
     sse.addEventListener('volume',function(data){remotePlaylist.updateVolume(data.data)});

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
        this.event('close'); ''
        if(noclose == null ||noclose == false){
            remotePlaylist.close();
        }
    }
    remotePlaylist.setShareManager(manager);
}

// Create a original manager (for a specific share id). Receive event and apply on player (read music)
function CreateOriginal(playlist){
    var manager = {};
    var sse = new EventSource('/share?device=' + MusicPlayer.device.name);

     // Set receive events behaviour
     sse.addEventListener('id',function(data){
         manager.id = parseInt(data.data);
     });
     sse.addEventListener('reload',function(){
        location.href='/?autoshare=true';
     });
     sse.addEventListener('add',function(data){
          playlist.addMusicsFromIds({ids:data.data.split(',').map(v=>parseInt(v))},true);
      });
     sse.addEventListener('playlist',function(data){
         playlist.addMusicsFromIds(JSON.parse(data.data),true);
     });
     sse.addEventListener('askPlaylist',function(data){
         var ids = playlist.list.map(function(m){return parseInt(m.id)});
         var data = {
             ids:ids,
             current:playlist.current,
             playing:!MusicPlayer.isPause(),
             volume:Math.round(MusicPlayer.player.volume*100),
             radio:Radio.currentRadio
        };
         $.ajax({
             url:'/shareUpdate',
             data:{id:manager.id,event:'playlist',data:JSON.stringify(data)}
         });
     });
     sse.addEventListener('remove',function(data){playlist.removeMusic(data.data);});
     sse.addEventListener('radio',function(data){Radio.read(data.data);});
     sse.addEventListener('stopRadio',function(data){MusicPlayer.stop();});
     sse.addEventListener('cleanPlaylist',function(data){playlist.cleanPlaylist();});
     sse.addEventListener('playMusic',function(data){
        var d = JSON.parse(data.data);
        playlist.showMusicByPosAndId(d.position,d.id);
        //playlist.playMusic(data.data);
    });
     sse.addEventListener('next',function(){playlist.next();});
     sse.addEventListener('previous',function(){playlist.previous();});
     sse.addEventListener('pause',function(){playlist.pause();});
     sse.addEventListener('play',function(){playlist.play();});
     sse.addEventListener('volumeUp',function(){playlist.volumeUp();});
     sse.addEventListener('volumeDown',function(){playlist.volumeDown();});
    manager.sse = sse;

    manager.event = function(event,data){
        data = data == null ? "" : data;
        $.ajax({
            url:'/shareUpdate',
            data:{id:this.id,event:event,data:data}
        });
    };
    manager.disable = function(){
        MusicPlayer.controls.setShareManager(Share.emptyManager);
        playlist.setShareManager(Share.emptyManager);
        this.sse.close();
        this.event('close');
    };
    MusicPlayer.controls.setShareManager(manager);
    playlist.setShareManager(manager);
    return manager;
}