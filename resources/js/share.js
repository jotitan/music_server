// Use to remote control playlist

// when share, create a connexion to server
let Share = {
    original:null,
    emptyManager:{event:function(){},disable:function(){}},
    listShared:[],
    enable:()=>{
        if(MusicPlayer.device.name === "No name"){
            alert("Define a real name for device");
            return;
        }
        // Get unique share id from server
        Share.original = CreateOriginal(PlaylistPanel);
        PlaylistPanel.open();
    },
    disable:()=>{
        Share.original.disable();
        Share.original = null;
        $('.share-button').removeClass('active');
    },
    getShares:callback=>{
        ajax({
            url:basename + 'shares',
            dataType:'json',
            success:function(data){
                callback(data);
            }
        });
    },
    addShare:(id,remotePlaylist)=>Share.listShared[id] = remotePlaylist,
    removeShare:id=>delete(this.listShared[id]),
    init:()=>{
        // Manage original share
        $('.share-button').bind('click',function(){
            if(Share.original!=null){
                Share.disable();
            }else{
                Share.enable();
            }
        });
        // If name existe, enable autoshare
        if(MusicPlayer.device.name !== "No name" && MusicPlayer.device.name !== ""){
            Share.enable();
        }
    }
};

// Create a clone manager (for a specific share id). receive event from original
function CreateClone(id,remotePlaylist){
    // Add behaviour on remotePlaylist, receive event for remoteplaylist
    let manager = {id:id};
    let sse = new EventSource(`/share?id=${id}&device=${MusicPlayer.device.name}`);
    sse.onerror = e=>{
        console.log("Error with share",e)
    };
    // Set receive events behaviour
    sse.addEventListener('add',data => {
        remotePlaylist.addMusicsFromIds({ids:data.data.split(',')},true);
    });
    sse.addEventListener('close',()=>manager.disable());
    sse.addEventListener('playlist',function(data){
        let info = JSON.parse(data.data);
        remotePlaylist._cleanPlaylist(true);
        remotePlaylist.addMusicsFromIds(info,true);
        info.playing ? remotePlaylist._play() : remotePlaylist._pause();
        remotePlaylist.updateVolume(info.volume);
        if(info.radio !== ""){
            remotePlaylist.selectRadio(info.radio);
        }
    });
    sse.addEventListener('remove',data => remotePlaylist._removeMusic(data.data,true));
    sse.addEventListener('cleanPlaylist',(data)=>remotePlaylist._cleanPlaylist(true));
    // Send position and position to check
    sse.addEventListener('playMusic',data => {
        let d = JSON.parse(data.data);
        remotePlaylist.showMusicByPosAndId(d.position,d.id, true);
    });
    // No next or previous, receive only position to play
    sse.addEventListener('pause',()=>remotePlaylist._pause());
    sse.addEventListener('play',()=>remotePlaylist._play());
    sse.addEventListener('volume',data => remotePlaylist.updateVolume(data.data));

    manager.sse = sse;
    manager.event = function(event,data){
        data = data == null ? "" : data;
        ajax({
            url:basename + 'shareUpdate',
            data:{id:this.id,event:event,data:data}
        });
    };
    manager.disable = function(noclose){
        this.sse.close();
        this.event('close');
        if(noclose == null ||noclose === false){
            remotePlaylist.close();
        }
    };
    remotePlaylist.setShareManager(manager);
}

// Create a original manager (for a specific share id). Receive event and apply on player (read music). Send event to clones
function CreateOriginal(playlist){
    let manager = {};
    let sse = new EventSource('/share?device=' + MusicPlayer.device.name);

    sse.onerror = (e)=>{
        console.log('err',e)
        $('.share-button').removeClass('active');
    };

    sse.onopen = ()=>{
        $('.share-button').addClass('active');
    };

    // Set receive events behaviour
    sse.addEventListener('id',data=>manager.id = parseInt(data.data));
    sse.addEventListener('reload',()=>location.href='/?autoshare=true');
    sse.addEventListener('add',data=>playlist.addMusicsFromIds({ids:data.data.split(',').map(v=>parseInt(v))},true));
    sse.addEventListener('playlist',data=>{
        console.log("GGG",data)
        playlist.addMusicsFromIds(JSON.parse(data.data),true)
    });
    /*sse.addEventListener('check-latency',rawData=>{
        var localReceive = Math.round(window.performance.now()*1000000);
        var data = JSON.parse(rawData.data);
        var originalTime = parseInt(data.time);
        ajax({
            url:basename + 'latency',
            data:{id:data.id,local_receive:localReceive,local_push:Math.round(window.performance.now()*1000000),original_time:originalTime}
        })
    });*/
    sse.addEventListener('askPlaylist',()=>{
        let ids = playlist.list.map(function(m){return parseInt(m.id)});
        let data = {
            ids:ids,
            current:playlist.current,
            position:MusicPlayer.player.currentTime,
            playing:!MusicPlayer.isPause(),
            volume:Math.round(MusicPlayer.player.volume*100),
            radio:Radio.currentRadio,
        };
        ajax({
            url:basename + 'shareUpdate',
            data:{id:manager.id,event:'playlist',data:JSON.stringify(data)}
        });
    });
    sse.addEventListener('remove',data=>playlist.removeMusic(data.data));
    sse.addEventListener('radio',data=>Radio.read(data.data));
    sse.addEventListener('stopRadio',()=>MusicPlayer.stop());
    sse.addEventListener('cleanPlaylist',()=>playlist.cleanPlaylist());
    sse.addEventListener('playMusic',data=>{
        var d = JSON.parse(data.data);
        playlist.showMusicByPosAndId(d.position,d.id);
    });
    sse.addEventListener('next',()=>playlist.next());
    sse.addEventListener('previous',()=>playlist.previous());
    sse.addEventListener('pause',()=>playlist.pause());
    sse.addEventListener('play',()=>playlist.play());
    sse.addEventListener('volume',e=>playlist.updateVolume(e.data));
    sse.addEventListener('volumeUp',()=>playlist.volumeUp());
    sse.addEventListener('volumeDown',()=>playlist.volumeDown());
    manager.sse = sse;

    manager.event = function(event,data){
        data = data == null ? "" : data;
        ajax({
            url:basename + 'shareUpdate',
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
