/* Explore the data of the cluster */

if(Loader){Loader.toLoad("html/explorer.html");}

var ExplorerManager  = {
    // Store explorers by name
    explorers:[],
    // Save a new explorer
    register:function(name,title,dataProvider){
        var panel = $.extend({},Explorer);
        panel.title = title;
        panel.init(name);
        panel.dataProvider = dataProvider;
        this.explorers[name] = panel;
        return panel;
    },
    open:function(name,params){
        return this.explorers[name].open(params);
    }
}

var Explorer = {
    // Path of folders
    breadcrumb:null,
    // Div where folders are displayed
    panelFolder:null,
    // Current path displayed
    currentPath :"",

    currentTypeLoad:"",
    fctClick(id){ActivePlaylist.getReal(this).addMusicFromId(id)},
    // Store the position of scroll on home (to restore when returning to home). Reset on genre. Update when click on folder
    scrollPosition:0,
    title:"",
    // dataProvider load data, params is a json object
    dataProvider:function(params){console.log("Not implemented")},
    init:function(id){
        $.extend(true,this,Panel) ;
        var clone = $('#idExplorePanel').clone();
        $('body').append(clone);
        clone.attr('id',id);
        this.initPanel($('#' + id),'<span class="glyphicon glyphicon-hdd"></span> ' + this.title);
        this.div.resizable({minWidth:250});
        this.breadcrumb = $('.breadcrumb',this.div)
        var _self = this;
        this.breadcrumb.on('click','li',function(){
            // delete nexts
            $(this).find('~').remove();
            _self.loadPath($(this).data('params'),"",true);
        });
        this.panelFolder = $('.folders',this.div);

        $('.switch',this.div).bind('click',()=>this.changeZoom());

        this.div.bind('open',()=>this._open(arguments));
        this._loadGenres();
        $('.info-folders > span.filter > :text',this.div).bind('keyup',e=>{
            var value = $(e.target).val();
            if (value.length <=2){
                $('>span',this.panelFolder).show();
                e.stopPropagation();
                return;
            }
            if (value.length > 2){
                // Fitler results
                $('>span:not([data-idx*="' + value + '"])',this.panelFolder).hide()
                $('>span[class^="' + value + '"]',this.panelFolder).show()
            }
            e.stopPropagation();
        });
    },
    // Load genres and add behaviour to list
    _loadGenres:function(){
        var select = $('.info-folders > span.filter > select.genres',this.div);
        $('option:not(:first)',select).remove();
        $.ajax({
            url:'/genres',
            dataType:'json',
            success:function(genres){
                for(var i in genres){
                    select.append('<option value="' + genres[i] + '">' + genres[i] + '</option>');
                };
            }
        });
        var _self = this;
        select.bind('change',function(){
            _self.reloadPath({genre:$(this).val()});
        });
    },
    // Reload same data (maybe new data or different urlServer
    reloadPath:function(params){
        //params.path = this.currentPath;
        this.loadPath(params,"",true);
    },
    loadPath:function(params,display,noAddBC){
        $('.info-folders > span.filter > :text',this.div).val("");
        //this.currentPath = params.path;
        // Add element in breadcrumb
        var currentScrollPosition = this.panelFolder.scrollTop();
        if(!noAddBC){
            this.addBreadcrumb(params,display);
        }
        if(this.currentTypeLoad == ""){
            var url = "";
        }
        this.dataProvider(params,data=>{
            this.display(data);
            if(JSON.stringify(params) == "{}"){
                // Restore home scroll
                this.panelFolder.scrollTop(this.scrollPosition);
            }
            this.scrollPosition = currentScrollPosition;
        });
    },
    changeZoom:function(){
        if ($('.folders',this.div).hasClass('block')){
            $('.folders',this.div).removeClass('block').addClass('line');
        }else{
            $('.folders',this.div).removeClass('line').addClass('block');
        }
    },
    // Call when first open
    _open:function(){
        this.scrollPosition = 0;
        this.breadcrumb.empty();
        if(this.title!=null){
            this.div.find('.title>span:first').html(this.title);
        }
        this.loadPath({},"Home");
        $('.info-folders > span.filter > select.genres',this.div).val("");
    },
    addBreadcrumb:function(params,display){
        if(JSON.stringify(params) === "{}"){
            // Reset breadcrumb
            this.breadcrumb.empty();
        }
        display = display || 'Empty';
        var li= $('<li><a href="#">' + display + '</a></li>');
        li.data('params',params);
        this.breadcrumb.append(li);
    },
    cleanPanel(){
        // Javascript native, fastest
        var parent = this.panelFolder.parent();
        parent.get(0).removeChild(this.panelFolder.get(0))
        this.panelFolder = $('<div class="folders line"></div>');
        parent.append(this.panelFolder);
    },
    display:function(data,noEmpty){
        if(!noEmpty) {
            this.cleanPanel();
        }
        for(var file in data){
            var name = "";
            var url = "";
            var params = {};
            if(Number(file) == file) {
                // Case when {}
                name = data[file].name;
                url = data[file].url;
                if (url != null) {
                    var rawParams = url.split("=");
                    params[rawParams[0]] = rawParams[1];
                }else{
                    params["id"] = data[file].id;
                }
            }
            var info = "";
            // Hidden fields
            var hidden = {}
            if(data[file].infos!=null){
                // List of data info
                for(var field in data[file].infos){
                    var value = data[file].infos[field];
                    switch(field){
                        case "time" : info += '<span class="info">' + MusicPlayer._formatTime(value) + '</span>';break;
                        case "favorite" :
                            info += '<span class="info ' + (value != "true" ? 'not-':'') + 'favorite"></span>';
                            break;
                        case "hidden":
                            var split = value.split("=");
                            hidden[split[0]] = split[1];
                            break;
                        default : info += '<span class="info">' + value + '</span>';
                    }
                };
            }
            if(data[file].info!=null){
                info = '<span class="info">' + MusicPlayer._formatTime(data[file].info) + '</span>';
            }
            var addOption = '<span style="margin-right:10px;" class="glyphicon glyphicon-plus add-music info-left"></span>';
            // Info json with name and either url (param after url) or id
            var span = $('<span class="music" data-idx="' + name.toLowerCase() + '" data-url="' + url + '">'
                + addOption + '<span class="music-name">' + name + '</span>' +info + '</span>');
            span.data('hidden',hidden);
            this.improveSpanActions(span);

            // If url, sub folder exist. Otherwise, final element, can add to playlist
            if(url != null){
                // Can add all data in playlist if drag and drop
                var dragStart = false;
                span.data('params',params);
                span.data("dataProvider",this.dataProvider);
                span.draggable({axis:'x',cancel:'.add-music',delay:300,revert:true,helper:'clone',start:()=>{dragStart=true},stop:()=>{dragStart=false}});
                $('.add-music',span).bind('mouseup',e=>{
                    var params = $(e.currentTarget).closest('.music').data('params');
                    this.dataProvider(params,data=>
                        ActivePlaylist.getReal(this).addMusicsFromIds({ids:data.filter(m =>m.id != null).map(m=>parseInt(m.id))})
                    );
                    e.stopPropagation();
                });
                span.bind('mouseup',e=>{
                    if(dragStart){return;}
                    var line = $(e.currentTarget);
                    // extract parameters from url
                    this.loadPath(line.data('params'),line.text());
                });
            }else{
                // Last element, display server where data is
                span.data("id",data[file].id)
                // Dbl click to playlist
                if(this.fctClick){
                    $('.add-music',span).bind('click',e=>this.fctClick($(e.currentTarget).closest('.music').data('id')));
                    span.bind('dblclick',e=>this.fctClick($(e.currentTarget).data('id')));
                }
            }
            // Add favorite behaviour
            $('.favorite,.not-favorite',span).bind('click',e=>this._changeFavorite($(e.target)));
            this.panelFolder.append(span);
        }
        $('.info-folders > span.counter',this.div).html('' + this.panelFolder.find('>span').length + ' - ');
    },
    // Implement to add content
    improveSpanActions(){},
    size:function(){
        return parseInt($('.info-folders > span.counter',this.div).html());
    },
// Update favorite of music
    _changeFavorite:function(span){
        var line = span.closest('.music');
        var id = line.data('id');
        var favorite = !span.hasClass('favorite');
        $.ajax({
            url:'/setFavorite?id=' + id + "&value=" + favorite,
            dataType:'json',
            success:function(data){
                if(data.error == null){
                    span.removeClass('favorite not-favorite').addClass(data.value ? 'favorite':'not-favorite');
                    $(document).trigger('refresh-favorites');
                }
            }
        });
    },
    setInfoKey:function(key,span){
        span.attr('data-trigger','focus').attr('tabindex','0').popover({
            title:span.text(),
            html:true,
            content:function(){
                var data = ClusterAction.whereIsKey(key,null,true);
                var str = '<div><div style="font-weight:bold">Master : ' + ClusterAction.getAlias(data[0]) + '</div>';
                str+='<div>Replica(s) : ';
                for(var i = 1 ; i < data.length ; i++){
                    str+=ClusterAction.getAlias(data[i]) + ' ';
                }
                return str + '</div></div>';
            }

        })
    }
}