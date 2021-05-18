// Manage a progress bar to show music current time


var MusicProgressBar = {
    //canvas:null,
    width:0,
    height:0,
    // Total time in seconds
    totalTime:0,
    currentTime:0,
    //colors:{},
    visible:false,
    init:function(id){
        this.div = $('#' + id);
        //this.canvas = $('canvas','#' + id).get(0).getContext('2d');
        this.width = $('#' + id).width();
        this.height = $('#' + id).height();
        //this.colors = {fcts:this._defineColorEvolution("74D0F1","22427C"),startColor:this._getColorInt("74D0F1")};
    },
    setVisible:function(visible){
        this.visible = visible;
    },
    load:function(total){
        total = total | 300;
        this.totalTime = total;
        this.currentTime = 0;
        this.redraw();
    },
    updateStr:function(strTime){
        this.update(parseInt(strTime.substring(0,2))*60 + parseInt(strTime.substring(3,5)));
    },
    update:function(time){
        this.currentTime = time;
        this.redraw();
    },
    redraw:function(){
        if(!this.visible){
            return;
        }
        //this.canvas.clearRect(0,0,this.width,this.height);
        if(this.totalTime<=0){
            return;
        }
        var percent = Math.round(this.currentTime*100/this.totalTime);
        this.div.css('background','linear-gradient(to right,#74D0F1 0%,#22427C ' + percent + '%,rgba(255,0,0,0) 0%)')
        /*var currentColor = this.colors.startColor;
        for(var i = 0 ; i < this.width*percent ; i++){
            this.canvas.fillStyle=this._getColorHexa(currentColor);
            this.canvas.fillRect(i,0,1,this.height-1);
            currentColor = this._changeColor(currentColor,this.colors.fcts);
        }*/
    } 
    /*_changeColor:function(color,fcts){
         return {r:fcts.r(color.r),g:fcts.g(color.g),b:fcts.b(color.b)};
    },*/
    /*_getColorInt:function(hexa){
        return {r:this._toInt(hexa.substring(0,2)),g:this._toInt(hexa.substring(2,4)),b:this._toInt(hexa.substring(4,6))};
    },*/
    /*_getColorHexa:function(color){
        return "#" + this._toHexa(color.r) + "" + this._toHexa(color.g) + "" + this._toHexa(color.b);
    },*/
    /*_defineColorEvolution:function(from,to){
        var colorFrom = this._getColorInt(from);
        var colorTo = this._getColorInt(to);

        var delta = {r:colorFrom.r - colorTo.r,g:colorFrom.g - colorTo.g,b:colorFrom.b - colorTo.b};
        var step = {r:Math.abs(delta.r) / this.width,g:Math.abs(delta.g) / this.width,b:Math.abs(delta.b) / this.width};
        var fct = {
            r:delta.r < 0 ? function(val){return val+step.r} : function(val){return val-step.r},
            g:delta.g < 0 ? function(val){return val+step.g} : function(val){return val-step.g},
            b:delta.b < 0 ? function(val){return val+step.b} : function(val){return val-step.b}
        }
        return fct;
    },
    _toHexa:function(nb){
        var hex = Math.round(nb).toString(16);
        return (hex.length == 1 ? "0" : "") + hex;
    },
    _toInt:function(hex){
        return parseInt(hex,16);
    }*/
}