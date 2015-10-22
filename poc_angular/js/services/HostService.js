function HostService(){
    this.host = "";

    if(localStorage){
        if(localStorage["host"]!=""){
            this.host = localStorage["host"];
        }
    }

    this.setHost = function(host){
        this.host = host;
        localStorage["host"] = this.host;
    }

    this.getHost = function(){
        return this.host;
    }

    this.getBaseUrl= function(){
        return this.host + "/rest/";
    }
}