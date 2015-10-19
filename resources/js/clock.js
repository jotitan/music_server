/* Define a clock */

function Timer(div){
    this.div = div

    this.run = function(){
        this.show()
        var _self = this;
         var interval = setInterval(function(){
            if(new Date().getSeconds() == 0){
                // Run every minutes
                clearInterval(interval)
                _self.show()
                setInterval(function(){
                    _self.show()
                },60000)
            }
         },1000)
    },
    this.show = function(){
        var date = new Date();
        time = ((date.getHours()<10)?"0":"") + date.getHours() + ":" + ((date.getMinutes()<10)?"0":"") + date.getMinutes()
        this.div.html(time)
    }


    this.run();
}