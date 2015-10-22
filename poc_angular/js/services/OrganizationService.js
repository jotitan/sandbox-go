function OrganizationService($http,HostServ){
    this.results = {entities:[],books:[],portfolios:[]};

    this.GetResults = function(){
        return this.results;
    }

    this._get = function(url,dataName){
        var _self = this;
        $http.jsonp(HostServ.getBaseUrl() + url)
            .success(function(data){
                _self.results[dataName] = data;
            });
    }

    this.updateEntities = function(date){
        this.results.books = [];
        this.results.portfolios = [];
        this._get('organization/get_all_entities?callback=JSON_CALLBACK&date=' + date,"entities");
    }

    this.updateBooks = function(date,entity){
        this.results.portfolios = [];
        this._get('organization/get_books_by_entities?callback=JSON_CALLBACK&date='
            + date + '&entities=' + entity,"books");
    }

    this.updatePortfolios = function(date,book){
        this._get('organization/get_portfolios_by_books?callback=JSON_CALLBACK&date=' + date + "&books=" + book,'portfolios');
    }


}