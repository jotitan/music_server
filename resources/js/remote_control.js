// Remote control for little remote
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
            // Todo verify
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

// Create a simple and light remote control (only play, pause, next and previous). No music read
function CreateRemote(id,target){
    if(target == null){
        return null;
    }
    var manager = {id:id};
    var sse = new EventSource('/share?id=' + id + '&device=' + MusicPlayer.device.name);

    sse.addEventListener('close',function(){
        manager.disable();
    });
    sse.addEventListener('playMusic',function(response){
        manager.loadMusic(JSON.parse(response.data).id);
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