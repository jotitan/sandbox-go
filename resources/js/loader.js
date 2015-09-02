/* Manage the load of html bloc */

var Loader = {
    modules : [],
    current:0,
    toLoad:function(url){
        this.modules.push(url);
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
            var div = $('<div></div>').load(this.modules[this.current],function(){
                Loader.current++;
                Loader._load(callback);
            });
            $('body').append(div);
        }
    }

}