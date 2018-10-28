var FREQUENCE_REFRESH = 3000

/* Panel mecanism (from all windows herit) */

var moveOptions = {can:true};

/* Represent a panel. Extend this. Don't use directly */
var Panel = {
    div:null,
    // Saved css info about panel
    cssInfo:null,
    name:"",
    id:"",
    initPanel : function(div,name,singleton){
        var _self = this
        this.name = name;
        this.id = div.attr('id')
        this.div = div;
        this.div.draggable({handle:".title",containment:"window",disabled:!moveOptions.can});
        this.singleton = singleton;
        this.div.offset(PanelPositionManager.get());

        $('.title > span:first',this.div).after('<span class="glyphicon glyphicon-minus close minimize_button"></span>');
        if($('.title',this.div).data('nomaximize') == null){
            $('.title > span:first',this.div).after('<span class="glyphicon glyphicon-resize-full close maximize_button"></span>');
        }
        $('.title > span:first',this.div).after('<span class="glyphicon glyphicon-remove close close_button"></span>');

        $('.close_button',this.div).bind('click',function(){
            _self.close()
        })
        $('.minimize_button',this.div).bind('click',function(){
            WindowsNavManager.doAction(_self)
        });
        if($('.maximize_button',this.div).length > 0){
            $('.maximize_button',this.div).bind('click',function(){
                _self.toggleMaximize($(this));
            });
            $('.title',this.div).bind('dblclick',function(e){
                if(e.target.nodeName == "DIV"){
                    _self.toggleMaximize($('.maximize_button',_self.div));
                }
            });
        }
        this.div.bind('mousedown',function(){
            WindowsNavManager.setActive(_self);
            $(document).trigger('focus.' + $(this).attr('id'));
        });

        // Split screen
        if($('.half-screen',this.div).length > 0 && $('.half-screen',this.data).data('split')!=""){
            var _self = this;
            var other = window[$('.half-screen',this.div).data('split')];
            $('.half-screen',this.div).after('<span class="glyphicon glyphicon glyphicon-eye-open close end-half-screen" style="display:none"/>');
            $('.half-screen',this.div).unbind('click').bind('click',function(e){
                // Check if explorer is open
                if (other.isVisible()){
                    other.halfMaximize(_self);
                    $(this).hide();
                    $('.end-half-screen',_self.div).show();
                }
                e.stopPropagation();
            });

            $('.end-half-screen',this.div).unbind('click').bind('click',function(){
                $(this).hide();
                $('.half-screen',_self.div).show();
                other.endHalfMaximize(_self);
            });
        }
    },
    show:function(){
        this.div.show()
    },
    hide:function(){
        this.div.hide()
    },
    isVisible:function(){
        return this.div.is(":visible")
    },
    open:function(){
        this.div.trigger('open',arguments);
        this.show();
        WindowsNavManager.add(this);
    },
    close:function(){
        WindowsNavManager.remove(this);
        PanelPositionManager.closeOne();
        this.div.trigger('close');
        this.hide()
    },
    toggleMaximize:function(button){
          if(this.div.hasClass('maximize-panel-size')){
             this.unmaximize();
             button.removeClass('glyphicon-resize-small').addClass('glyphicon-resize-full');
             this.unblockAction();
         }else{
             this.maximize();
             button.removeClass('glyphicon-resize-full').addClass('glyphicon-resize-small');
             this.blockAction();
         }
    },
    maximize:function(){
        this._css.save(this.div);
        this._css.remove(this.div,'maximize-panel-size');
    },
    unmaximize:function(){
        this._css.restore(this.div);
    },
    blockAction:function(){
        this.div.draggable('disable');
        this.div.resizable('disable');
    },
    unblockAction:function(){
        this.div.draggable('enable');
        this.div.resizable('enable');
    },
    // place two panels at each part of screen (left and right)
    halfMaximize:function(otherPanel){
        this._css.save(this.div);
        otherPanel._css.save(otherPanel.div);

        this.blockAction();
        otherPanel.blockAction();

        this._css.remove(this.div,'half-left-panel');
        otherPanel._css.remove(otherPanel.div,'half-right-panel');
    },
    endHalfMaximize:function(otherPanel){
        this.unblockAction();
        otherPanel.unblockAction();

        this._css.restore(this.div);
        otherPanel._css.restore(otherPanel.div);
    },
    _css:{
        info:null,
        save:function(div){
            this.info = {
                width:div.width(),
                height:div.height(),
                top:div.position().top,
                left:div.position().left,
                boxShadow:div.css('box-shadow')
            };
        },
        restore:function(div){
           if(this.info == null){
               return;
           }
           div.removeClass('maximize-panel-size');
           div.removeClass('half-right-panel');
           div.removeClass('half-left-panel');
           div.css({
               left:this.info.left,
               top:this.info.top,
               width:this.info.width,
               height:this.info.height,
               boxShadow:this.info.boxShadow
           });
           this.info = null;
        },
        remove:function(div,classToAdd){
            div.css({left:'',top:'',width:'',height:'',boxShadow:'0px 0px 0px 0px'})
            div.removeClass('normal-panel-size');
            div.addClass(classToAdd);
        }
    }
}

function CloneDiv(id,prefix){
    var div = $('#' + id).clone();
    div.attr('id',prefix + '_' + new Date().getTime()).css('display','');
    $('body').append(div);
    return div;
}

// Give the position of the new panel
var PanelPositionManager = {
    total:0,
    nb:0,
    top:100,
    left:100,
    get:function(){
        this.total++;
        this.nb++;
        return {top:this.top + 40*this.total,left:this.left + 30*this.total};
    },
    closeOne:function(){
        if(--this.nb <=0){
            this.total = 0;
            this.nb = 0;
        }
    }
};

/* Use to manage window display in the taskbar. Manipulate Panel */
var WindowsNavManager = {
    div:null,
    init:function(id){
        this.div = $('#' + id)
        $('ul',this.div).sortable({
            cancel:false,
            axis:"x",
            distance:10,
            helper:'clone'
        });
    },
    // Panel contain variable name, method hide, show and isVisible
    add:function(panel){
        if(panel.singleton){
            // Check button exist
            if($('ul button[data-inner-id="' + panel.id + '"]',this.div).length > 0){
                return
            }
        }
        var button = $("<button class=\"btn btn-default task-button active\">" + panel.name + "</button>")
        button.data("panel",panel).attr("data-inner-id",panel.id)
        panel.div.data("button",button).attr("data-inner-id",panel.id)
        button.bind('click',function(){
            WindowsNavManager.doAction($(this).data("panel"))
        })
         $('ul',this.div).append(button.wrap("<li></li>"))
        this.setActive(panel)
    },
    remove:function(panel){
        this.div.find('button[data-inner-id="' + panel.id + '"]').remove()
        //this.div.find('button[data-inner-id="' + panel.id + '"]').remove()
    },
    setActive:function(panel){
        $('button.task-button',this.div).removeClass('selected');
        $('.float-panel.active').removeClass('active')
        panel.div.addClass('active')
        if(panel.div.data('button').length > 0){
            panel.div.data('button').removeClass('inactive').addClass('selected active');
        }
    },
    doAction:function(panel){
        if(panel == null){
            return
        }
        var button = this.div.find('button[data-inner-id="' + panel.id + '"]')
        if(panel.isVisible()){
            //panel.hide();
            //button.removeClass('active').addClass('inactive')
            // If not active, set active instead
            if(button.hasClass('selected')){
                panel.hide();
                button.removeClass('active').addClass('inactive')
            }else{
                this.setActive(panel);
            }
        }else{
            this.setActive(panel);
            panel.show();
        }
    }
}

/* Show nodes status in the task bar */
var MiniStatusViewer = {
    canvas : null,
    nodes:[],
    mapId:[],
    toRefresh:false,   // Indicate to refresh widget
    init:function(canvasId){
        if($('#' + canvasId).length == 0){
            return;
        }
        this.canvas = $('#' + canvasId).get(0).getContext('2d')
        this.canvas.width = parseInt($('#' + canvasId).attr('width'))
    },
    update:function(id,up){
        this.toRefresh |= this.nodes[this.mapId[id]].setUp(up);
    },
    add:function(id,up){
        if(this.mapId[id]!=null){
            return this.update(id,up);
        }
        var pos = this.nodes.length;
        var ui = new NodeUI(this.canvas.width,Math.floor(this.nodes.length/2),this.nodes.length%2);
        this.nodes.push(ui);
        this.mapId[id] = pos;
        this.toRefresh = true;
    },
    refresh:function(){
        if(!this.toRefresh){
            return;
        }
        if(this.canvas == null){
            return;
        }
        this.toRefresh = true;
        this.canvas.clearRect(0,0,100,25)
        var toDelete = [];
        this.nodes.forEach(function(node,i){
            if(node == null){return;}
            if (!node.draw(this.canvas)){
                toDelete.push(i)
            }
            node.updated = false;
        },this)
        this.deleteNodes(toDelete)
    },
    // Delete old nodes, must shift all nodes staying
    deleteNodes:function(positions){
        return;
        positions.forEach(function(pos){
            this.nodes[pos] = null;
        },this)
    }
}

function NodeUI(canvasSize,pos,line){
    this.up = true;
    this.pos = canvasSize - (15 + pos*12);
    this.line = 1 + line*12;
    this.updated = true;    // To know if node already exist in cluster

    this.setUp = function(up){
        this.updated = true;
        var change = this.up != up
        this.up = up
    }

    this.draw = function(canvas){
        canvas.fillStyle = this.up ? "green":"red"
        canvas.fillRect(this.pos,this.line,10,10)
        return this.updated;
    }
}