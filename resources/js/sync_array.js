// Implementation of a pseudo synchronized array
function SyncArray(){
    this.array = new Array();
    // Store the current lock. If 2 at same time, only the first in list do the job
    this._syncList = new Array();

    // Sync delete
    this.delete = function(value){
        // First if an existing delete is running
        if(this._syncList.length > 0){
            // Have to wait 10ms and start again
            setTimeout(()=>this.delete(value),10);
            return;
        }else{
            // Add is value inside and check after if size is always 1 with my entry
            this._syncList.push(value);
            try{
                if(this._syncList.length == 1 && this._syncList[0] == value){
                    // Do the real delete and remove the lock after
                    // Delete this.array[value]
                    this._syncList = [];
                }else{
                    // Check if the first element is mine, if true, continue and remove all others entries
                
                    if(this._syncList.length > 1 && this._syncList[0] == value){
                        this._syncList = [];
                    }else{
                        setTimeout(()=>this.delete(value),10);
                    }
                }
            }catch(e){
                // An error will be thrown if syncList is get empty during check. Just relaunch delete
                setTimeout(()=>this.delete(value),10);
            }
        }
    }

    this.push = function(value){
        this.array.push(value);
    }

    this.pop = function(){
        return this.array.pop();
    }
    this.get = function(){
        return this.array;
    }

}