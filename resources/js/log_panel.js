/* Explore the data of the cluster */

if(Loader){Loader.toLoad("html/log_panel.html","LogPanel");}

var Logger = {
    info(message){
        if(LogPanel.isVisible()){
            LogPanel.log(message);
        }else{
            console.log(message);
        }
    },
    error(message){
        this.info("ERROR : " + message);
    }
};

var LogPanel = {
    title:"",
    // dataProvider load data, params is a json object
    dataProvider:function(params){console.log("Not implemented")},
    init:function(id){
        $.extend(true,this,Panel) ;
        this.initPanel( $('#idLogPanel'),'<span class="glyphicon glyphicon-hdd"></span> ' + this.title);
        this.div.resizable({minWidth:250});
        $('.trash_log',this.div).bind('click',()=>this.cleanPanel());
        this.div.bind('open',()=>this._open(arguments));
    },
    // Call when first open
    _open:function(){
    },

    log(message){
       $('.content',this.div).prepend(`<span style="display:block">${message}</span>`);
    },
    cleanPanel(){
        // Javascript native, fastest
        $('.content',this.div).empty();
    }
}