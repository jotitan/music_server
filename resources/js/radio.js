if(Loader){Loader.toLoad("html/radio.html","RadioPanel");}

// Sounds radio from website ecouter-en-direct. Read radio into hidden iframe
var Radio = {
    url:"https://ecouter-en-direct.com/",
    radios:{
        "FIP":"http://direct.fipradio.fr/live/fip-midfi.mp3",
        "France Musique":"http://direct.francemusique.fr/live/francemusique-midfi.mp3",
        "Radio Classique":"http://radioclassique.ice.infomaniak.ch/radioclassique-high.mp3",
        "Rire et Chansons":"http://cdn.nrjaudio.fm/audio1/fr/30401/mp3_128.mp3?origine=fluxradios",
        "RMC":"https://rmcinfo.cdn.dvmr.fr/rmcinfo",
        "RTL2":"http://streaming.radio.rtl2.fr/rtl2-1-44-96",
        "TSF JAZZ":"http://tsfjazz.ice.infomaniak.ch/tsfjazz-high.mp3"},
    currentRadio:"",
    stopRadio:function(){
        this.currentRadio = "";
    },
    getRadios:function(){
        var radios = [];
        for(var r in this.radios){
            radios.push(r);
        }
        return radios;
    },
    read:function(radio){
        if(this.radios[radio] == null){
            return;
        }
        this.currentRadio = radio;
        MusicPlayer.loadUrl(this.radios[radio],{artist:'Radio',title:radio});
        MusicPlayer.controls.setTitle({artist:'Radio',title:radio});        
    }    
}

var RadioPanel = {
    init:function(idDiv,title){
        idDiv = idDiv || '#idRadio';
        title = title || 'Radios'
        $.extend(true,this,Panel)
        this.initPanel($(idDiv),'<span class="glyphicon glyphicon-music icon"></span>' + title,true);
        this.div.resizable({minWidth:250});
        this.load();
        $('span.play',this.div).on('click',(e)=>Radio.read($(e.target).data('radio')));
        this.div.on('close',()=>MusicPlayer.stop());
    },
    load:function(){
        Radio.getRadios().forEach(radio=>{
            var line = $('<div class="music"></div>');
            line.append('<span></span><span><span class="music-title">' + radio + '</span></span>');
            line.append('<span data-radio="' + radio + '" class="glyphicon glyphicon-play play" title="Play"></span>');
            $('.content',this.div).append(line)
        });
        
    }
}