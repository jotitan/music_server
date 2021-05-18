// Remote control for little remote
var RemoteControlManager = {
    manager:null,
    div:null,
    divSelect:null,
    url:'',
    init:function(idDiv,idSelect,musicUrl){
        this.url = musicUrl || '';
        this.div = $('#' + idDiv);
        this.divSelect = $('#' + idSelect);
        $('.play',this.div).bind('click',()=>{
            this.manager.event('play');
            this.setIsPlaying(true);
        });

        $('.pause',this.div).bind('click',()=>{
            this.manager.event('pause');
            this.setIsPlaying(false);
        });

        $('.previous',this.div).bind('click',()=>this.manager.event('previous'));

        $('.next',this.div).bind('click',() =>this.manager.event('next'));

        $('.vup',this.div).bind('click',() =>this.manager.event('volumeUp'));
        $('.vdown',this.div).bind('click',() =>this.manager.event('volumeDown'));

        this.divSelect.bind('change',e=>this._connect($(e.target).val()));
        this.initVolumeSlider();
        Share.getShares(data=> {
            if(data == null || data.length === 0){
                return this.divSelect.hide();
            }
            if(data.length === 1){
                // Auto select the on_self._connect(data[0].Id);dxmàb
                this._connect(data[0].Id);
                return this.divSelect.hide();
            }else{
                this.divSelect.select()
                this.divSelect.empty().append('<option>...</option>');
                data.forEach(s=>this.divSelect.append('<option value="' + s.Id + '">' + s.Name + '</option>'));
            }
        });

        this.timer.init();
        return this;
    },
    initVolumeSlider:function(){
        $(".volume-slider").roundSlider({
            radius: 80,
            circleShape: "pie",
            sliderType: "min-range",
            value: 50,
            startAngle: 315,
            editableTooltip: false,
            mouseScrollAction: true,
            handleShape: "dot",
            circleShape: "half-top",
            width: 33,
            step:5,
            change:e=>this.manager.event('volume',e.value)
        });
    },
    // Manage time counter
    timer:{
        totalPanel:null,
        progressPanel:null,
        currentTime:0,
        currentTotal:0,
        interval:null,
        init(){
            this.progressPanel = $('.time-progress > .current-position',this.div);
            this.totalPanel = $('.time-progress > .total-length',this.div);
        },
        reset(){
            this.set(0,true);
        },
        set(time,start,stop){
            this.currentTime = time;
            // Avoid override total
            if(this.currentTotal != 0 && this.currentTime > this.currentTotal){
                clearInterval(this.interval);
                this.interval = null;
                return;
            }
            this.progressPanel.html(MusicPlayer._formatTime(this.currentTime));
            if(start === true){
                // Start interval and update time
                if(this.interval != null){
                    clearInterval(this.interval);
                }
                this.interval = setInterval(()=>this.set(this.currentTime+1,false),1000);
            }
            if(stop ===true && this.interval != null){
                clearInterval(this.interval);
                this.interval = null;
            }
        },
        setTotal(total){
            this.currentTotal = parseInt(total);
            this.totalPanel.html(MusicPlayer._formatTime(total));
        }
    },
    _connect:function(id){
        this.manager = CreateRemote(id,this);
        this.divSelect.hide();
        this.div.show();
    },
    updateMusic:function(music) {
        $('.title',this.div).html(music.title + " - " + music.artist);
        $('.cover > img',this.div).attr('src',music.cover);
        this.timer.setTotal(music.length);
    },
    updateVolume:function(value) {
        $(".volume-slider").roundSlider('option', 'value', value);
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
};

// Create a simple and light remote cntrol (only play, pause, next and previous). No music read
function CreateRemote(id,target){
    if(target == null){
        return;
    }
    var manager = {id:id};
    var sse = new EventSource('/share?id=' + id + '&device=' + MusicPlayer.device.getName());
    sse.addEventListener('close',()=>manager.disable());
    sse.addEventListener('playMusic',function(response){
        target.timer.setTotal(response.length);
        target.timer.reset();
        manager.loadMusic(JSON.parse(response.data).id);
        target.setIsPlaying(true);
    });

    sse.addEventListener('pause',(response)=>{
        var data = JSON.parse(response.data);
        target.setIsPlaying(false);
        target.timer.set(data.position,false,true);
    });
    sse.addEventListener('volume',(response)=>{
        target.updateVolume(response.data);
    });
    sse.addEventListener('play',(response)=>{
        var data = JSON.parse(response.data);
        target.setIsPlaying(true);
        target.timer.set(data.position,true,false);
    });

    sse.addEventListener('playlist',response=>{
        var data = JSON.parse(response.data);
        if(data.current!==-1){
            var idMusic = data.ids[data.current];
            manager.loadMusic(idMusic);
            target.setIsPlaying(data.playing);
            target.timer.set(data.position,data.playing,!data.playing);
            // Set volume
            target.updateVolume(data.volume);
        }
    });

    manager.sse = sse;

    manager.loadMusic = function(id){
        ajax({
            url:`${basename}musicInfo?id=${id}`,
            success:data=>target.updateMusic(JSON.parse(data))
        })
    };
    manager.event = function(event,data){
        data = data == null ? "" : data;
        ajax({
            url:basename + 'shareUpdate',
            data:{id:this.id,event:event,data:data}
        });
    };
    manager.disable = function(noclose){
        console.log("CLOSE STREAM");
        this.sse.close();
        this.event('close');
        if(noclose == null ||noclose === false){
            remotePlaylist.close();
        }
    };
    return manager;
}
