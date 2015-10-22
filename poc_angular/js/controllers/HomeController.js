function HomeController($scope,$route,$location,HostServ){
    $scope.title='Home';
    $scope.connectedHost = HostServ.getHost();
    $scope.changeHost = function(newHost){
        HostServ.setHost(newHost);
    }
    $scope.route = $route;
    /*$scope.routeParams = $routeParams;
     $scope.location = $location;*/
}