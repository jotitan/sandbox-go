var paramsDatePicker = {
    format:"yyyymmdd",
    autoclose:true,
    language:"fr",
    daysOfWeekDisabled:0,
    calendarWeeks:true
}

// Use a service to make request
function OrganizationController($scope,OrganizationServ){
    $scope.results = OrganizationServ.GetResults();
    $scope.selected = {date:"",entity:"",book:""};

    $('#idDatePicker').datepicker(paramsDatePicker);

    $scope.$watch('selected.date',function(){
        if ($scope.selected.date!="") {
            $scope.selected.entity = "";
            $scope.selected.book = "";
            OrganizationServ.updateEntities($scope.selected.date);
        }
    })

    $scope.$watch('selected.entity',function(){
        if ($scope.selected.date!="" && $scope.selected.entity!="") {
            $scope.selected.book = "";
            OrganizationServ.updateBooks($scope.selected.date,$scope.selected.entity);
        }
    })

    $scope.$watch('selected.book',function(){
        if ($scope.selected.date!="" && $scope.selected.entity!="" && $scope.selected.book!="") {
            OrganizationServ.updatePortfolios($scope.selected.date,$scope.selected.book);
        }
    })
}