/* Base of explorer panel, extended by explorer and my playlists */
var BaseExplorer = {
    // Path of folders
    breadcrumb:null,
    // Div where folders are displayed
    panelFolder:null,
    // Current path displayed
    currentPath :"",
    
    currentTypeLoad:"",
    fctClick:null,
    // Store the position of scroll on home (to restore when returning to home). Reset on genre. Update when click on folder
    scrollPosition:0,
    initBaseExplorer:function(idPanel, title){
        $.extend(true,this,Panel) ;
        this.initPanel($(`#${id}`),`<span class="glyphicon glyphicon-hdd"></span> ${title}`);
        this.div.resizable({minWidth:250});
        this.breadcrumb = $('.breadcrumb',this.div)
        var _self = this;
        this.breadcrumb.on('click','li',function(){
            // delete nexts
            $(this).find('~').remove();
            _self.loadPath($('a',$(this)).data('path'),"",true);
        });
        this.panelFolder = $('.folders',this.div);

        $('.switch',this.div).bind('click',()=>this.changeZoom());

        this.div.bind('open',()=>this._openBase(arguments));
        // Add filter on elements displayed
        $('.info-folders > span.filter > :text',this.div).bind('keyup',e=>{
            var value = $(this).val().toLowerCase();
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
    addClickBehave:function(fct){
       this.fctClick = fct;
    },
    // Reload same data (maybe new data or different urlServer
    reloadPath:function(){
        this.loadPath(this.currentPath,"",true);
    },
    // Load data from source (must be override in implementation)
    loadData:function(path, success){
        alert("Must be override");
    },
    loadPath:function(path,display,noAddBC){
        $('.info-folders > span.filter > :text',this.div).val("");
        this.currentPath = path;
        // Add element in breadcrumb
        var currentScrollPosition = this.panelFolder.scrollTop();
        if(!noAddBC){
            this.addBreadcrumb(path,display);
        }
        if(this.currentTypeLoad == ""){
            var url = "";
        }
        this.loadData(path,data=>{
            this.display(data);
            if(path == ""){
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
    _openBase:function(){
        this.scrollPosition = 0;
        this.breadcrumb.empty();
        this.loadPath("","Home");
    },
    addBreadcrumb:function(path,display){
        display = display || path;
        this.breadcrumb.append('<li><a href="#" data-path="' + path + '">' + display + '</a></li>');
    },
    display:function(data){
        this.panelFolder.empty();
        var _self = this;
        var nb = 0;
        for(var file in data){
            var name = "";
            var url = "";
            if(Number(file) == file){
                // Case when {}
                name = data[file].name;
                url = data[file].url;
            }else{
                // Normal case of map
                name = file;
                url = "path=" + Explorer.currentPath + file + "/"
            }
            
            var info = "";
            if(data[file].infos!=null){
                // List of data info
               for(var field in data[file].infos){
                    var value = data[file].infos[field];
                    switch(field){
                        case "time" : info += '<span class="info">' + MusicPlayer._formatTime(value) + '</span>';break;
                        case "favorite" : 
                            info += '<span class="info ' + (value != "true" ? 'not-':'') + 'favorite"></span>';    
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
            // If url, sub folder exist. Otherwise, final element, can add to playlist
            if(url != null){
                // Can add all data in playlist if drag and drop
                var dragStart = false;
                span.data("url_drag",this.urlServer + (this.urlServer.indexOf("?") == -1 ? '?':'&') + url);
                span.draggable({axis:'x',cancel:'.add-music',delay:300,revert:true,helper:'clone',start:()=>{dragStart=true},stop:()=>{dragStart=false}});                
                $('.add-music',span).bind('mouseup',e=>{                    
                    var url = $(e.currentTarget).closest('.music').data('url_drag');
                    PlaylistPanel.addMusicsFromUrl(url);
                    e.stopPropagation();
                });
                span.bind('mouseup',function(){
                    if(dragStart){return;}
                    Explorer.loadPath($(this).data('url'),$(this).text());
                });
            }else{
                // Last element, display server where data is
                span.data("id",data[file].id)
                //span.draggable({revert:true,helper:'clone'})
                // Dbl click to playlist
                if(this.fctClick){
                    $('.add-music',span).bind('click',e=>Explorer.fctClick($(e.currentTarget).closest('.music').data('id')));
                    span.bind('dblclick',e=>Explorer.fctClick($(e.currentTarget).data('id')));
                }
            }
            // Add favorite behaviour
            $('.favorite,.not-favorite',span).bind('click',e=>Explorer._changeFavorite($(e.target)));
            this.panelFolder.append(span);
        }
        $('.info-folders > span.counter',this.div).html('' + this.panelFolder.find('>span').length + ' - ');
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