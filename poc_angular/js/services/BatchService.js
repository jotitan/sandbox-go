function BatchService($http,HostServ){

    this.GetBatches = function($scope){
        console.log("Execute GetBatches");
        var url = HostServ.getBaseUrl() + "batch/get_catalogue?callback=JSON_CALLBACK";
        var _self = this;
        $http.jsonp(url).success(function(results){
            return $scope.batches = results;
        }).error(function(){console.log(arguments)});

    }
    
    


}