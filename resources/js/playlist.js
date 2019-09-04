if (Loader) { Loader.toLoad("html/playlist.html", "PlaylistPanel", "css/playlist.css"); }

// Local playlist
var PlaylistPanel = {
    listDiv: null,
    list: [],
    current: -1,
    saveInLS: true,
    isPaused: false,
    shareManager: Share.emptyManager,
    // To avoid load music
    noLoadMusic: false,
    init: function (idDiv, title, noLoad) {
        idDiv = idDiv || '#idPlaylist';
        title = title || 'Playlist'
        $.extend(true, this, Panel)
        this.initPanel($(idDiv), '<span class="glyphicon glyphicon-music icon"></span>' + title, true)
        this.div.resizable()
        this.listDiv = $('.playlist', this.div);
        // Select behaviour
        var _self = this;
        this.listDiv.on("click", 'div', function () {
            $('div', _self.listDiv).removeClass('focused');
            $(this).addClass('focused');
        });
        this.listDiv.on("dblclick", 'div', function (e) {
            window.getSelection().removeAllRanges();
            if (!_self.noLoadMusic) {
                MusicPlayer.load($(this).data("music"));
            }
            _self.setActualPlayed($(this));
            _self.shareCurrent();
            _self.saveCurrent();
        });
        $(document).unbind('delete_event').bind('delete_event', function () {
            // Delete music. Find position element in list
            _self.removeMusic(_self.getFocusedPosition());
        });
        /*$(document).unbind('next_event').bind('next_event',function(){
            _self.next();
            $(document).trigger('show_next_event');
        });
        $(document).unbind('previous_event').bind('previous_event',function(){
            _self.previous();
            $(document).trigger('show_previous_event');
        });*/
        this.listDiv.droppable({
            drop: function (event, ui) {
                var idMusic = ui.draggable.data('id');
                if (idMusic == null) {
                    // Get url if exist to load all, folder case
                    var dataProvider = ui.draggable.data("dataProvider");
                    if(dataProvider != null) {
                        dataProvider(ui.draggable.data("params"), data => _self.addMusicsFromIds({ids: data.filter(m => m.id != null).map(m => parseInt(m.id))}));
                    }
                } else {
                    // Get info from id music
                    _self.addMusicFromId(idMusic);
                }
            }
        });
        $('.trash-playlist', this.div).bind('click', function () {
            _self.cleanPlaylist();
        });
        $('.shuffle-playlist', this.div).bind('click', function () {
            _self.shuffle();
        });
        this.div.on('close', function () {
            ActivePlaylist.set(null);
            _self.cleanPlaylist(true);
            // If share, close it
            _self.shareManager.disable();
        });
        // Load saved playlist
        if (noLoad == true) {
            this.saveInLS = false;
        }
        this.load();
        this.initSearch();
        $(document).bind('focus.' + this.div.attr('id'), () => ActivePlaylist.set(this));
    },
    play: function () {
        MusicPlayer.play();
        this.isPaused = false;
        this.shareManager.event('play',JSON.stringify({position:MusicPlayer.player.currentTime}));
    },
    pause: function () {
        MusicPlayer.pause();
        this.isPaused = true;
        this.shareManager.event('pause',JSON.stringify({position:MusicPlayer.player.currentTime}));
    },
    updateVolume:function(value){
        MusicPlayer.volume.set(value);
    },
    volumeUp: function () {
        MusicPlayer.volume.up();
    },
    volumeDown: function () {
        MusicPlayer.volume.down();
    },
    // return songs around current
    initSearch: function () {
        $('#idSearch,.remoteSearch', this.div).autocomplete({
            source: 'search?size=20&',
            minLength: 2,
            position: { my: "left bottom", at: "left top", collision: "none" },
            focus: function (e, ui) {
                $('#idSearch').val(ui.item.title);
                return false;
            },
            select: function (event, ui) {
                ActivePlaylist.getReal().shareManager.event('add', ui.item.id);
                ActivePlaylist.getReal().add(ui.item);
            }
        }).autocomplete("instance")._renderItem = function (ul, item) {
            return $("<li>")
                .append("<a><b>" + item.title + "</b><br>" + item.artist + " (<font size='-2'>" + item.album + "</font>)</a>")
                .appendTo(ul);
        };
    },
    shuffle: function () {
        var shuffle = [];
        for (var i = 0; i < this.list.length; i++) {
            shuffle.push({ value: Math.random(), index: i });
        }
        shuffle = shuffle.sort(function (a, b) { return a.value < b.value; });
        var shuffleList = [];
        for (var i = 0; i < shuffle.length; i++) {
            shuffleList.push(this.list[shuffle[i].index]);
        }
        this.list = [];
        this.listDiv.empty();
        for (var i = 0; i < shuffleList.length; i++) {
            this.add(shuffleList[i], true, true);
        }
        this.save();
    },
    getSong: function (shift, from) {
        from = from || this.current;
        if (from != -1 && from + shift >= 0 && from + shift < this.list.length) {
            return this.list[from + shift];
        }
        return null;
    },
    getNSongs: function (n) {
        var musics = {};
        if (this.list == null || this.list.length == 0) {
            return musics;
        }
        var c = this.current != -1 ? this.current : 0;
        musics[0] = this.list[c];
        for (var i = c; i <= c + n && i < this.list.length; i++) {
            musics[i - c] = this.list[i];
        }
        for (var i = c; i >= c - n && i >= 0; i--) {
            musics[i - c] = this.list[i];
        }
        return musics;
    },
    cleanPlaylist: function (noShare) {
        this._cleanPlaylist(noShare);
    },
    _cleanPlaylist: function (noShare) {
        if (this.saveInLS && localStorage) {
            delete (localStorage["playlist"]);
            delete (localStorage["current"]);
        }
        this.listDiv.empty();
        this.list = [];
        this.current = -1;
        this.updateTotal();
        if (!noShare) {
            this.shareManager.event('cleanPlaylist');
        }
    },
    hideRadio: function () { },
    // Show music by position. Check if id is same. If not, reload playlist
    showMusicByPosAndId: function (pos, id) {
        if (pos >= this.list.length || this.list[pos].id != id) {
            //error
            console.log("Error");
            return;
        }
        if (!this.noLoadMusic) {
            MusicPlayer.load(this.list[pos]);
        }
        this.showMusicByPosition(pos);
        this.hideRadio();
    },
    showMusicByPosition: function (position) {
        this.current = parseInt(position);
        this._selectLine();
        this.play();
    },
    // No need to propagate share
    playMusic: function (id) {
        this.list.forEach(function (music, i) {
            if (music.id == id) {
                var line = $('>div:nth-child(' + (i + 1) + ')', this.listDiv);
                MusicPlayer.load(music);
                this.setActualPlayed(line);
                this.saveCurrent();
                return;
            }
        }, this);
    },
    setActualPlayed: function (line) {
        if (line.length == 0) { return; }
        $('div', this.listDiv).removeClass('played selected focused');
        line.addClass('played');
        this.current = this.getPlayedPosition();

        var position = Math.max(0, this.listDiv.scrollTop() + line.position().top - this.listDiv.height() / 2);

        if (this.listDiv.scrollTop() != position) {
            this.listDiv.scrollTop(position);
        }
    },
    getPlayedPosition: function () {
        var nb = $('> div', this.listDiv).length
        var afters = $('.played~div', this.listDiv).length
        return nb - afters - 1;
    },
    getFocusedPosition: function () {
        var nb = $('> div', this.listDiv).length;
        var afters = $('.focused~div', this.listDiv).length
        return nb - afters - 1;
    },
    getCurrentMusic: function () {
        if (this.current != -1) {
            return this.list[this.current];
        }
        return null;
    },
    // Return selected or first in list
    getOne: function () {
        var focused = $('div.focused', this.listDiv);
        if (focused.length > 0) {
            this.current = this.getFocusedPosition();
            this.saveCurrent();
            this._selectLine();
            return focused.data('music');
        }
        focused = $('div.played', this.listDiv);
        if (focused.length > 0) {
            this.current = this.getPlayedPosition();
            this.saveCurrent();
            this._selectLine();
            return focused.data('music');
        }
        if (this.list.length > 0) {
            this.current = 0;
            this.saveCurrent();
            this._selectLine();
            return this.list[0];
        }
        return null;
    },
    load: function () {
        if (this.saveInLS && localStorage && localStorage["playlist"] != null) {
            var musics = JSON.parse(localStorage["playlist"]);
            var currentMusic = localStorage["current"];
            this.current = currentMusic != null ? parseInt(currentMusic) : -1;
            musics.forEach(function (m, i) {
                this.add(m, true);
            }, this);
            this._selectLine();
            this.open();
        }
    },
    // Save current playlist and current music in localstorage
    save: function () {
        if (this.saveInLS && localStorage) {
            localStorage["playlist"] = JSON.stringify(this.list);
        }
    },
    updateTotal: function () {
        $('.total_playlist', this.div).html(this.list.length);
    },
    saveCurrent: function () {
        if (this.saveInLS && localStorage) {
            localStorage["current"] = this.current;//this.list[this.current].id;
        }
    },
    removeMusicId: function (id, noShare) {
        this.list.forEach(function (music, i) {
            if (music.id == id) {
                return this.removeMusic(i, noShare);
            }
        }, this);
    },
    removeMusic: function (index, noShare) {
        this._removeMusic(index, noShare);
    },
    // Do the remove of music from list and div
    _removeMusic: function (index, noShare) {
        $('>div:nth-child(' + index + ')', this.listDiv).remove();
        var music = this.list.splice(index - 1, 1)[0];
        this.save();
        this.updateTotal();
        if ((noShare == null || noShare == false)) {
            // Send index to remove, to be sure not to remove duplicates but only the good line
            this.shareManager.event("remove", index);
        }
    },
    // Add many musics from list of id
    addMusicsFromIds: function (datas, noShare) {
        var ids = datas.ids;
        this.current = datas.current != null ? datas.current : this.current;
        var _self = this;
        if (ids.length > 0) {
            $.ajax({
                url: basename + 'musicsInfo?ids=' + JSON.stringify(ids),
                dataType: 'json',
                success: data => {
                    var musics = [];
                    data.forEach(m => musics[m.id] = m);
                    ids.forEach(id => {
                        var music = musics[id];
                        _self.add(music, true, true);
                    });
                    if ((noShare == null || noShare == false)) {
                        _self.shareManager.event('add', ids.join(','));
                    }
                    _self.save();
                    _self.updateTotal();
                    _self._selectLine();
                }
            });
        }
    },
    addMusicFromId: function (id, noShare) {
        if (noShare == null || noShare == false) {
            this.shareManager.event('add', id);
        }
        $.ajax({
            url: basename + 'musicInfo?id=' + id,
            dataType: 'json',
            // No need to create a real Music, just a container with properties, no methods
            success: data => this.add(data)
        })
    },
    addMusicsFromUrl: function (url) {
        var _self = this;
        $.ajax({
            url: url,
            dataType: 'json',
            success: function (data) {
                // If list receive with id in each element, add music
                var ids = data.filter(m =>m.id != null).map(m=>parseInt(m.id));
                _self.addMusicsFromIds({ ids: ids });
            }
        });
    },
    // Add a new music in list
    add: function (music, noSave, noTotal) {
        var line = $('<div class="music"></div>');
        line.append('<span class="glyphicon glyphicon-remove remove" title="Remove"></span>');
        line.append('<span><span class="pc artist">' + music.artist + ' -</span> <span class="music-title">' + music.title + '</span><span class="artist2">' + music.artist + '</span></span>');
        line.append('<span class="time">' + MusicPlayer._formatTime(music.length) + '</span>');
        line.append('<span class="glyphicon glyphicon-play play" title="Play"></span>');
        var _self = this;
        $('.glyphicon-play', line).bind('click', function () {
            _self.setActualPlayed($(this).closest('div'));
            _self.shareCurrent();
            if (!_self.noLoadMusic) {
                MusicPlayer.load(music);
                _self.saveCurrent();
            }
        });
        $('.glyphicon-remove', line).bind('click', function () {
            var nb = _self.listDiv.find('div').length;
            var pos = nb - $(this).parent().find('~div').length;
            _self.removeMusic(pos);
        });
        line.data("music", music);
        this.listDiv.append(line);
        this.list.push(music);
        if (noSave == null || noSave == false) {
            this.save();
        }
        if (noTotal == null || noTotal == false) {
            this.updateTotal();
        }
    },
    _selectLine: function () {
        if (this.current != null && this.current == -1) {
            return
        }
        var line = $('div:nth-child(' + (this.current + 1) + ')', this.listDiv);
        this.setActualPlayed(line);
    },
    // Share current music
    shareCurrent: function () {
        this.shareManager.event('playMusic', JSON.stringify({length:this.list[this.current].length, position: this.current, id: this.list[this.current].id }));
        this.hideRadio();
    },
    next: function (noShare) {
        if (this.current >= this.list.length-1) {
            return;
        }
        this.current++;
        this._selectLine();
        this.saveCurrent();
        if (!noShare) {
            this.shareCurrent();
        }
        if (!this.noLoadMusic) {
            MusicPlayer.load(this.list[this.current]);
        }
        $(document).trigger('show_next_event');
    },
    previous: function (noShare) {
        if (this.current <= 0) {
            return;
        }
        this.current--;
        this._selectLine();
        this.saveCurrent();
        if (!noShare) {
            this.shareCurrent();
        }
        if (!this.noLoadMusic) {
            MusicPlayer.load(this.list[this.current]);
        }
        $(document).trigger('show_previous_event');
    },
    setShareManager: function (manager) {
        this.shareManager = manager;
    }
}

// Playlist of a remote device
var RemotePlaylist = {
    init2: function () {
        // If remote player, shareManager exist for sure
        var _self = this;
        this.div.unbind('close').bind('close', function () {
            // When closing, delete
            _self.div.remove();
            _self.shareManager.disable(true);
            if (ActivePlaylist.get() == _self) {
                ActivePlaylist.set(null);
            }
            Share.removeShare(_self.id);
        });
        if (Radio) {
            Radio.getRadios().forEach(radio => $('.list-radios', this.div).append('<option value="' + radio + '">' + radio + '</option>'))
            $('.list-radios', this.div).bind('change', r => this.readRadio($(r.currentTarget).val()));
        }
    },
    readRadio: function (radio) {
        radio != "" ? this.shareManager.event('radio', radio) : this.shareManager.event('stopRadio');
    },
    updateVolume: function (value) {
        $('.remoteVolume').css('background-image', 'linear-gradient(to right,white 0%,orange ' + value + '%,white 0%');
    },

    // Show music by id, have to find position
    showMusicById: function (id) {
        var position = this.list.findIndex(m => m.id == id);
        if (position != -1) {
            this.showMusicByPosition(position);
        }
    },
    previous: function () {
        this.shareManager.event("previous");
    },
    next: function () {
        this.shareManager.event("next");
    },
    volumeUp: function () {
        this.shareManager.event("volumeUp");
    },
    volumeDown: function () {
        this.shareManager.event("volumeDown");
    },
    // show played music
    showMusic: function (id) {
        this.list.forEach(function (music, i) {
            if (music.id == id) {
                this.play();
                var line = $('>div:nth-child(' + (i + 1) + ')', this.listDiv);
                this.setActualPlayed(line);
                return;
            }
        }, this);
    },
    pause: function () {
        this.isPaused = true;
        this.shareManager.event("pause");
        this._pause();
    },
    _pause: function () {
        MusicPlayer._showPlaying(false);
    },
    hideRadio: function () {
        $('.list-radios', this.div).val("");
    },
    selectRadio: function (radio) {
        $('.list-radios', this.div).val(radio);
    },
    play: function () {
        this.isPaused = false;
        this.shareManager.event("play");
        this._play();
    },
    _play: function () {
        MusicPlayer._showPlaying(true);
    },
    // Override method remove music to just send information to share
    removeMusic: function (index, noShare) {
        this.shareManager.event("remove", index);
    },
    cleanPlaylist: function () {
        this.shareManager.event("cleanPlaylist");
    },

    addMusicsFromList:function(ids){
        this.shareManager.event("add", ids);
    },
    // Override add from url to avoid many add request to send only big one
    addMusicsFromUrl: function (url) {
        var _self = this;
        $.ajax({
            url: url,
            dataType: 'json',
            success: function (data) {
                //Send to share the list
                var ids = data.filter(m=>m.id != null).map(m=>parseInt(m.id)).join(',');
                _self.addMusicsFromList(ids);
            }
        });
    }
}

var ActivePlaylist = {
    previous:null,
    current:null,
    set(playlist){
        if(this.current === playlist){return;}
        this.previous = this.current;
        this.current = playlist;
    },
    get(){
        return this.current ==null ? this.getDefault() : this.current;
    },
    // Return the current playlist but not the same as caller
    getReal:function(caller){
        if(this.current !=null && caller === this.current){
            // return previous
            return this.previous == null ? this.getDefault() : this.previous;
        }
        return this.get();
    },
    getDefault:function(){
        PlaylistPanel.open();
        return PlaylistPanel;
    }
}

function connectToShare() {
    var sharePanel = $.extend({}, PlaylistPanel, RemotePlaylist);
    sharePanel.noLoadMusic = true;
    sharePanel.list = [];
    var id = 'idRemotePlaylist_' + new Date().getTime();
    Share.addShare(id, sharePanel);
    $('body').append($('#idRemotePlaylist').clone().attr('id', id));

    sharePanel.init('#' + id, 'Remote playlist', true);
    sharePanel.init2();

    // Show shares
    Share.getShares(function (data) {
        if (data.length == 1) {
            // only one share, direct connect
            sharePanel.open();
            //currentPlaylist = sharePanel;
            CreateClone(data[0].Id, sharePanel);
            $('.title > span:first', sharePanel.div).html('Remote playlist for ' + data[0].Name);
        } else {
            data.forEach(function (s) {
                var share = $('<div> => ' + s.Name + ' : <button>Connect</button></div>');
                $('button', share).bind('click', function () {
                    sharePanel.listDiv.empty();
                    $('.title > span:first', sharePanel.div).html('Remote playlist for ' + s.Name);
                    CreateClone(s.Id, sharePanel);
                });
                sharePanel.listDiv.append(share);
                sharePanel.open();
            });
        }
    });
}
