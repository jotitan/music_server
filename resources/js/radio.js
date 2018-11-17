if(Loader){Loader.toLoad("html/radio.html","RadioPanel");}

// Sounds radio from website ecouter-en-direct. Read radio into hidden iframe
var Radio = {
    url:"https://ecouter-en-direct.com/",
    radios:["rmc","rtl2","tsf-jazz","radio-fip"],
    read:function(radio){
        MusicPlayer.pause();
        MusicPlayer.controls.setTitle({artist:'Radio',title:radio.toUpperCase()});
        $('#idRadioFrame').attr('src',this.url + radio + "/");
    },
    stop:function(){
        $('#idRadioFrame').remove();
        MusicPlayer.controls.setTitle({artist:'',title:''});
    },
    init:function(){
        var iframe = '<iframe id="idRadioFrame" width=400 height=400 allow="autoplay" style="display:none" src="' + this.url + '"></iframe>';
        $('body').append(iframe);
    }
}
$(function(){Radio.init()})

var RadioPanel = {
    init:function(idDiv,title){
        idDiv = idDiv || '#idRadio';
        title = title || 'Radios'
        $.extend(true,this,Panel)
        this.initPanel($(idDiv),'<span class="glyphicon glyphicon-music icon"></span>' + title,true);
        this.div.resizable({minWidth:250});
        this.load();
        $('span.play',this.div).on('click',(e)=>Radio.read($(e.target).data('radio')));
        this.div.on('close',()=>Radio.stop());
    },
    load:function(){
        Radio.radios.forEach(radio=>{
            var line = $('<div class="music"></div>');
            line.append('<span></span><span><span class="music-title">' + radio.toUpperCase() + '</span></span>');
            line.append('<span data-radio="' + radio + '" class="glyphicon glyphicon-play play" title="Play"></span>');
            $('.content',this.div).append(line)
        });
        
    }
}