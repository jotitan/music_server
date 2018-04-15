if(Loader){Loader.toLoad("html/favorites.html","FavoritesPanel");}

var FavoritesPanel = {
    init:function(idDiv,title){
        idDiv = idDiv || '#idFavorites';
        title = title || 'Favorites'
        $.extend(true,this,Panel)
        this.initPanel($(idDiv),'<span class="glyphicon glyphicon-music icon"></span>' + title,true);
        this.div.resizable({minWidth:250});
        $(idDiv).bind('open',()=>this.load());
        $(idDiv).bind('close',()=>$(document).unbind('refresh-favorites'));
        $(document).bind('refresh-favorites',()=>this.load());
    },
    load:function(){
        $.ajax({
            url:'/getFavorites',
            dataType:'json',
            success:(data)=>this.display(data)
        });
    },
    display:function(musics){
        $('.folders',this.div).empty();
        musics.forEach(music=>{
            var line = $('<div class="music"></div>');
            line.append('<span class="glyphicon glyphicon-remove remove" title="Remove"></span>');
            var info = '<span style="margin-left:10px;" class="glyphicon glyphicon-plus add-music info"></span>'
                +'<span class="info">' + MusicPlayer._formatTime(music.infos.time) + '</span>';
            line.append('<span class="details">' + music.infos.artist + ' - ' + music.name + info + '</span>');
            line.data("id",music.id);
            line.draggable({revert:true,helper:'clone'});
            $('.remove',line).bind('click',e=>FavoritesPanel._removeFavorite($(e.currentTarget).closest('.music')));
            $('.folders',this.div).append(line);
        });
        $('.folders .add-music',this.div).bind('click',function(e){
            PlaylistPanel.addMusicFromId($(e.currentTarget).closest('div').data('id'));
        });
        this._updateTotal(musics.length);
    },
    _removeFavorite:function(line){
        $.ajax({
            url:'/setFavorite?id=' + line.data('id') + '&value=false',
            dataType:'json',
            success:data=>{
                if(data.value == false)line.remove();
                FavoritesPanel._updateTotal($('div.music',FavoritesPanel.div).length);
            }
        });
    },
    _updateTotal:function(total){
        $('.total_favorites',this.div).html(total);
    }

}