function createMyPlaylistsPanel(){
    var myPlaylists = ExplorerManager.register("myPlaylists","My playlists",getMyPlaylistsProvider((playlistId,order)=>{
        myPlaylists.currentPlaylist=playlistId;
        myPlaylists.lastOrder=order;
    }));

    $('.filter',myPlaylists.div).remove();
    $('.title > span.glyphicon-list-alt',myPlaylists.div).after('<span class="glyphicon glyphicon-plus close"></span>');
    $('.glyphicon-plus',myPlaylists.div).on('click',()=>myPlaylists.newPlaylist());

    $(document).bind('focus.' + myPlaylists.id, () => ActivePlaylist.set(myPlaylists));
    // Save the working playlist
    myPlaylists.currentPlaylist = null;
    // Last order in the playlist
    myPlaylists.lastOrder = 0;

    myPlaylists.div.on('close', () =>{
        if (ActivePlaylist.get() === myPlaylists) {
            ActivePlaylist.set(null);
        }
    });
    // Receive list of musics from explorer
    myPlaylists.addMusicsFromIds = params=>{
        if(this.currentPlaylist == null){
            return;
        }
        // Add in database first
        var musics = params.ids.map(id=>{return {musicId:parseInt(id),playlistId:this.currentPlaylist,order:++this.lastOrder}});
        var orders = musics.map(m=>m.order);
        DatabaseAccess.addMusicsToPlaylist(musics,results=>{
            this.addMusicsInPlaylist(results,orders);
        });

    };
    // Receive a music from explorer
    myPlaylists.addMusicFromId = id => {
        if(this.currentPlaylist == null){
            return;
        }
        // Add in database first
        var music = {musicId:parseInt(id),playlistId:this.currentPlaylist,order:++this.lastOrder};
        DatabaseAccess.addMusicToPlaylist(music,()=>{
            this.addMusicsInPlaylist([music],[music.order]);
        });
    };

    myPlaylists.addMusicsInPlaylist = function(rawMusics,orders){
        var ids = rawMusics.map(m=>m.musicId);
        ajax({
            url: basename + 'musicsInfo?short=true&ids=' + JSON.stringify(ids),
            dataType: 'json',
            success: data => {
                var musics = [];
                data.forEach(m => musics[m.id] = m);
                var counter = 0;
                this.display(
                    ids.map(id=>{
                        musics[id].name = musics[id].infos.artist + " - " + musics[id].name;
                        musics[id].infos.hidden = "order=" + orders[counter++];
                        delete musics[id].infos.artist;
                        return musics[id];
                    }),true);
            }
        });
    };
    myPlaylists.improveSpanActions = (span)=>{
        var deleteOption = $('<span style="margin-right:10px;" class="glyphicon glyphicon-remove remove remove-music info-left"></span>')
        if(myPlaylists.currentPlaylist == null) {
            // Create delete button for playlist

            deleteOption.on('mouseup', (e) =>{
                var playlistId = $(e.target).closest('.music').data('params').id;
                DatabaseAccess.deletePlaylist(playlistId,() => $(e.target).closest('.music').remove());
                e.stopPropagation();
            })
        }else {
            deleteOption.on('mouseup', (e) => {
                var order = $(e.target).closest('.music').data('hidden').order;
                DatabaseAccess.deleteMusicFromPlaylist(myPlaylists.currentPlaylist, order, () => {
                    // Remove element from div
                    $(e.target).closest('.music').remove();
                });

            });
        }
        span.find('.add-music').before(deleteOption);
    };

    myPlaylists.newPlaylist = ()=>{
        var playlistName = prompt("New playlist name");
        DatabaseAccess.newPlaylist(playlistName,()=>{
            myPlaylists.loadPath({},"Home");
        });
    }
}

function getMyPlaylistsProvider(updateStatus){
    return (params,success)=>{
        if(params.id === undefined){
            DatabaseAccess.loadPlaylists(success);
            updateStatus(null);
        }else {
            // Load detail of a playlist
            DatabaseAccess.loadMusicOfPlaylist(parseInt(params.id), (musics,lastOrder) => {
                    var ids = musics.map(m=>m.musicId);
                    updateStatus(params.id, lastOrder);
                    var counter = 0;
                    ajax({
                        url: basename + 'musicsInfo?short=true&ids=' + JSON.stringify(ids),
                        dataType: 'json',
                        success: data => success(data.map(m => {
                            m.name = m.infos.artist + " - " + m.name;
                            m.infos.hidden = "order=" + musics[counter++].order;
                            delete m.infos.artist;
                            return m
                        }))
                    })
                }
            );
        }
    }
}
// Two tables : playlist (id, name) and music of playlist (playlistId, musicId)
let DatabaseAccess = {
    db:null,
    init(){
        window.indexedDB = window.indexedDB || window.mozIndexedDB || window.webkitIndexedDB || window.msIndexedDB;
        if(window.indexedDB == null){
            throw "Impossible to open database";
        }
        this.createDatabase();
        return this;
    },
    createDatabase() {
        var openRequest = window.indexedDB.open("MyPlaylists", 1);
        openRequest.onerror = () => alert("Impossible to open dabase");
        openRequest.onupgradeneeded = (e) => {
            this.db = e.target.result;
            var playlist = this.db.createObjectStore("playlist", {keyPath: "id", autoIncrement: true});
            playlist.createIndex("name", "name", {unique: true});

            var musicOfPlaylist = this.db.createObjectStore("musicPlaylist", {keyPath: ["playlistId","order"]});
            // Field musicId must be save also
            musicOfPlaylist.createIndex("playlistId","playlistId");
        };
        openRequest.onsuccess = (e) => this.db = e.target.result;
    },
    // Result : []id,lastOrder
    loadMusicOfPlaylist(playlistId,success){
        var store = this.db.transaction("musicPlaylist").objectStore("musicPlaylist");
        var index = store.index("playlistId");
        var request = index.getAll(playlistId.toString());
        request.onsuccess=e=>success(e.target.result.sort((a,b)=>a.order - b.order),this.computeLastOrder(e.target.result))
    },
    computeLastOrder(musics){
        return musics.length === 0 ? 0 : Math.max(...musics.map(m=>m.order));
    },
    loadPlaylists(success){
        var transaction = this.db.transaction(["playlist"]);
        var store = transaction.objectStore("playlist");
        var request = store.getAll();
        request.onsuccess=e=>success(e.target.result.map(r=>{return {name:r.name,url:"id="+r.id}}));
        request.onerror=()=>success([]);
    },
    newPlaylist(name,success){
        var transaction = this.db.transaction(["playlist"],"readwrite");
        var store = transaction.objectStore("playlist");
        var result = store.add({"name":name});
        result.onsuccess = success;
        result.onerror = ()=>alert("Impossible to add playlist with name " + name);
    },
    deletePlaylist(id,success){
        // Load playlist and remove all music after
        this.loadMusicOfPlaylist(id,musics=>{
            var transaction = this.db.transaction(["musicPlaylist","playlist"],"readwrite");
            if(musics.length === 0){
                var delRequest = transaction.objectStore("playlist").delete(parseInt(id));
                delRequest.onsuccess = success;
                delRequest.onerror = ()=>alert("Impossible to delete this playlist")
                return;
            }
            var musicPlaylistStore = transaction.objectStore("musicPlaylist");
            var counter = 0;
            var totalDone = 0;
            musics.map(m=>[m.playlistId,m.order]).forEach(m=>{
                var request = musicPlaylistStore.delete(m);
                request.onsuccess = ()=>{
                    totalDone++;
                    if(++counter === musics.length){
                        // Delete playlist in database
                        var delRequest = transaction.objectStore("playlist").delete(parseInt(id));
                        delRequest.onsuccess = success;
                        delRequest.onerror = ()=>alert("Impossible to delete this playlist")
                    }else{
                        if(totalDone === musics.length){
                            // End but an error previously append
                            alert("Impossible to delete this playlist, too many errors " + (totalDone - counter) + " / " + totalDone);
                        }
                    }

                };
                request.onerror = ()=>{
                    if(++totalDone === musics.length) {
                        alert("Impossible to delete this playlist, too many errors " + (totalDone - counter) + " / " + totalDone);
                    }
                }
            });
        });
    },
    deleteMusicFromPlaylist(idPlaylist,order,success){
        var store = this.db.transaction("musicPlaylist","readwrite").objectStore("musicPlaylist");
        var req = store.delete([idPlaylist,parseInt(order)]);
        req.onsuccess = success;
        req.onerror = ()=>alert("Impossible to delete music");
    },
    moveMusicInPlaylist(idPlaylist,idMusic){

    },
    addMusicToPlaylist(music,success){
        var store = this.db.transaction("musicPlaylist","readwrite").objectStore("musicPlaylist");
        var r = store.add(music);
        r.onsuccess = ()=>{
            success(music);
        };
        r.onerror = ()=>{
            alert("Impossible to add");
        }
    },
    // musics is a list with {musicId,order}
    addMusicsToPlaylist(musics, success){
        var store = this.db.transaction("musicPlaylist","readwrite").objectStore("musicPlaylist");
        var results = {musics:[],counter:0,ko:0};
        musics.forEach(music=>{
            var r = store.add(music);
            r.onsuccess = ()=>{
                results.counter++;
                results.musics.push(music);
                if(results.counter === musics.length){
                    success(results.musics);
                }
            };
            r.onerror = ()=>{
                results.counter++;
                results.ko++;
                if(results.counter === musics.length){
                    alert("Impossible to add : " + results.ko + " errors");
                }
            }
        });
    },
    // Extract
    extractAll(success){
        var transaction = this.db.transaction(["musicPlaylist","playlist"]);
        var storePlaylist = transaction.objectStore("playlist")
        var storeMusics = transaction.objectStore("musicPlaylist")
        storePlaylist.getAll().onsuccess = (e)=>{
            var results = {}
            e.target.result.forEach(p=>{
                results[p.id] = {name:p.name,musics:[]};
            });
            storeMusics.getAll().onsuccess = (e2)=>{
                e2.target.result.forEach(m=>{
                    results[m.playlistId].musics.push({id:m.musicId,order:m.order});
                });
                success(results)
            }
        }
    }
}.init();