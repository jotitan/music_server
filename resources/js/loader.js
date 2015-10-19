/* Manage the load of html bloc */

var Loader = {
    modules : [],
    current:0,
    toLoad:function(url,initFct){
        this.modules.push({url:url,fct:initFct});
    },
    launch:function(callback){
        this.current = 0;
        this._load(callback);
    },
    _load:function(callback){
        if(this.current >= this.modules.length){
            if(callback){
                callback();
            }
        }else{
            var element =this.modules[this.current];
            var div = $('<div></div>').load(element.url,function(){
                if(element.fct){
                    window[element.fct].init()
                }
                Loader.current++;
                Loader._load(callback);
            });
            $('body').append(div);
        }
    }

}