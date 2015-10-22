
function SearchController($scope,SearchServ) {
    $scope.query = "";

    if(localStorage["lastQuery"]!=null){
        $scope.query = localStorage["lastQuery"];
    }

    $scope.search = function () {
        localStorage["lastQuery"] = $scope.query;
        SearchServ.Execute($scope.query);
    }
}

function SearchCriteriaController($scope){

}

function OtherDataController($scope,$http,HostServ){
    $scope.pdlKinds = [];
    $scope.operators = [];
    $scope.columns = [];
    $scope.functions = [];
    $http.jsonp(HostServ.getBaseUrl() + "misc/get_pdlkind?callback=JSON_CALLBACK")
        .success(function(data){
            $scope.pdlKinds = data
        })
    $http.jsonp(HostServ.getBaseUrl() + "misc/get_operator?callback=JSON_CALLBACK")
        .success(function(data){
            $scope.operators = data;
        })
    $http.jsonp(HostServ.getBaseUrl() + "misc/get_columns?callback=JSON_CALLBACK")
        .success(function(data){
            $scope.columns = data;
        })
    $http.jsonp(HostServ.getBaseUrl() + "misc/get_functions?callback=JSON_CALLBACK")
        .success(function(data){
            $scope.functions = data;
        })
}

function ResultsController($scope,SearchServ,$templateCache){
    console.log($templateCache)
    $scope.results = SearchServ.GetResults();
    $scope.nbResult = 0;
    $scope.execTime = 0;
    $scope.subList = [];
    $scope.currentDataGroup = [];
    $scope.$watch("results.data",function(){
        $scope.gridOptions.data=$scope.results.data;
        $scope.gridOptions.columnDefs=$scope.results.headers;
        $scope.nbResult = $scope.results.data.length;
        $scope.execTime = $scope.results.executionTime;
    })

    $templateCache.put('ui-grid/personalExpandableRowHeader',
        "<div class=\"ui-grid-row-header-cell ui-grid-expandable-buttons-cell\"><div ng-show=\"getExternalScopes().visible\" class=\"ui-grid-cell-contents\"><i ng-class=\"{ 'ui-grid-icon-plus-squared' : !row.isExpanded, 'ui-grid-icon-minus-squared' : row.isExpanded }\" ng-click=\"grid.api.expandable.toggleRowExpansion(row.entity)\"></i></div></div>"
    );
    $scope.infos = {visible:false};

    $scope.selectField = "";
    $scope.$watch("selectField",function(){
        $scope.groupResults($scope.selectField.name);
    })
    $scope.groupResults = function(field){
        if(field == null){
            return;
        }
        $scope.currentDataGroup = [];// For each column value, list of element
        var  header = [];
        var nb = 0;
        $scope.results.data.forEach(function(f){
            var value = f[field];
            if (value == null){
                return;
            }
            if($scope.currentDataGroup[value] == null){
                nb++;
                $scope.currentDataGroup[value] = [f];
                var head = [];head[field] = value;
                header.push(head);
            }else{
                $scope.currentDataGroup[value].push(f)
            }
        })
        $scope.gridOptions.columnDefs = [{name:field,type:"string"}];
        $scope.gridOptions.data = header;
        $scope.infos.visible = true;
    }

    $scope.gridOptions = {
        expandableRowTemplate:'pages/subsearch.html',
        enableExpandableRowHeader:false,
        expandableRowHeight:150,

        onRegisterApi: function (gridApi) {
            var cellTemplate = 'ui-grid/personalExpandableRowHeader';
            gridApi.core.addRowHeaderColumn( { name: 'rowHeaderCol', displayName: '', width: 30, cellTemplate: cellTemplate} );

            gridApi.expandable.on.rowExpandedStateChanged($scope, function (row) {
                if($scope.selectField.name!="") {
                    var value = row.entity[$scope.selectField.name];
                    var data = $scope.currentDataGroup[value]
                    row.subData = {
                        data:data,
                        columnDefs:$scope.results.headers,
                    }
                }
            })
        }
    }
}