/* Explore the data of the cluster */

if(Loader){Loader.toLoad("html/explorer.html","Explorer");}

var Explorer = {
    breadcrumb:null,
    panelFolder:null,
    currentPath :"",
    currentTypeLoad:"",
    urlServer:"",
    baseServer:"",
    fctClick:null,
    // Store the position of scroll on home (to restore when returning to home). Reset on genre. Update when click on folder
    scrollPosition:0,
    init:function(){
        $.extend(true,this,Panel) ;
        this.initPanel($('#idExplorePanel'),'<span class="glyphicon glyphicon-hdd"></span> Explore');
        this.div.resizable({minWidth:250});
        this.breadcrumb = $('.breadcrumb',this.div)
        var _self = this;
        this.breadcrumb.on('click','li',function(){
            // delete nexts
            $(this).find('~').remove();
            _self.loadPath($('a',$(this)).data('path'),"",true);
        });
        this.panelFolder = $('.folders',this.div);

        $('.switch',this.div).bind('click',function(){
            Explorer.changeZoom();
        })

        this.div.bind('open',function(){
           Explorer._open(arguments);
        });
        this._loadGenres();
        $('.info-folders > span.filter > :text',this.div).bind('keyup',function(e){
            var value = $(this).val().toLowerCase();
            if (value.length <=2){
                $('>span',Explorer.panelFolder).show();
                e.stopPropagation();
                return;
            }
            if (value.length > 2){
                // Fitler results
                $('>span:not([data-idx*="' + value + '"])',Explorer.panelFolder).hide()
                $('>span[class^="' + value + '"]',Explorer.panelFolder).show()
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
            _self.urlServer = _self.baseUrl + "?genre=" + $(this).val();
            _self.reloadPath();
        });
    },
    addClickBehave:function(fct){
       this.fctClick = fct;
    },
    // Reload same data (maybe new data or different urlServer
    reloadPath:function(){
        this.loadPath(this.currentPath,"",true);
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
        $.ajax({
            url:this.urlServer + (this.urlServer.indexOf("?") == -1 ? '?':'&') + path,
            dataType:'json',
            success:function(data){
                Explorer.display(data);
                if(path == ""){
                    // Restore home scroll
                    Explorer.panelFolder.scrollTop(Explorer.scrollPosition);
                }
                Explorer.scrollPosition = currentScrollPosition;
            }
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
        if(arguments[0][1] == null){return;}
        this.scrollPosition = 0;
        this.breadcrumb.empty();
        this.urlServer = arguments[0][1];
        this.baseUrl = arguments[0][1];
        var title = arguments[0][2];
        if(title!=null){
            this.div.find('.title>span:first').html(title);
        }
        this.loadPath("","Home");
        $('.info-folders > span.filter > select.genres',this.div).val("");
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
                    var value = data[file].infos[field]
                    if(field == "time"){
                        info += '<span class="info">' + MusicPlayer._formatTime(value) + '</span>';
                    }else{
                        info += '<span class="info">' + value + '</span>';
                    }
                };
            }
            if(data[file].info!=null){
                info = '<span class="info">' + MusicPlayer._formatTime(data[file].info) + '</span>';
            }
            // Info json with name and either url (param after url) or id
            var span = $('<span data-idx="' + name.toLowerCase() + '" data-url="' + url + '">' + name + info + '</span>');
            // If url, sub folder exist. Otherwise, final element, can add to playlist
            if(url != null){
                // Can add all data in playlist if drag and drop
                var dragStart = false;
                span.data("url_drag",this.urlServer + (this.urlServer.indexOf("?") == -1 ? '?':'&') + url);
                span.draggable({delay:300,helper:'clone',start:function(){dragStart=true;},stop:function(){dragStart=false;}})
                span.bind('mouseup',function(){
                    if(dragStart){return;}
                    Explorer.loadPath($(this).data('url'),$(this).text());
                });
            }else{
                // Last element, display server where data is
                span.data("id",data[file].id)
                span.draggable({revert:true,helper:'clone'})
                // Dbl click to playlist
                if(this.fctClick){
                    span.bind('dblclick',function(){
                        Explorer.fctClick($(this).data('id'));
                    });
                }
            }
            this.panelFolder.append(span);
        }
        $('.info-folders > span.counter',this.div).html('' + this.panelFolder.find('>span').length + ' - ');
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