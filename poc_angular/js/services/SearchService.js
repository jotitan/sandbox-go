function SearchService($http,HostServ){
    this.results = {data:[],headers:[],executionTime:0};

    this.GetResults = function(){
        return this.results;
    }

    this.Execute = function(query){
        console.log("Execute query : ",query);
        var url = HostServ.getBaseUrl() + "search/query?callback=JSON_CALLBACK&query=" + query;
        var _self = this;
        $http.jsonp(url).success(function(results){
            _self.results.data = results.data;
            _self.results.headers = results.headers.map(function(h){return {name:h}});
            _self.results.executionTime= results.executionTime;
        }).error(function(){console.log(arguments)});

    }


}