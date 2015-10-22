
function configRoute(routeProvider){
    routeProvider
        .when('/home',{
            templateUrl:'pages/accueil.html'
        })
        .when('/search',{
            templateUrl:'pages/search.html'
        })
        .when('/searchCriteria',{
            templateUrl:'pages/search-criteria.html'
        })
        .when('/referentiel/organization',{
            templateUrl:'pages/organization.html'
        })
        .when('/referentiel',{
            templateUrl:'pages/referentiel.html'
        })
        .when('/referentiel/others',{
            templateUrl:'pages/others.html'
        })
        .when('/batch', {
        	templateUrl:'pages/batch.html'
        })
        .otherwise({
            redirectTo : 'home'
        });
}
