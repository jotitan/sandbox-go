function BatchController($scope, $modal, BatchServ) {
	$scope.batches = [];
	$scope.selectedbatch = ""

	BatchServ.GetBatches($scope);

	/*
	 * $scope.$watch('selected.batch', function() { if ($scope.selected.batch !=
	 * "") { console.log($scope.selected.batch); $scope.gridOptions.data =
	 * $scope.selected.batch.options; $scope.options =
	 * $scope.selected.batch.options; } })
	 */

	$scope.open = function(batch) {
		$scope.selectedbatch = batch
		var modalInstance = $modal.open({
			templateUrl : 'myModalContent.html',
			controller : 'ModalInstanceCtrl',
			size : 'lg',
			resolve : {
				batch : function() {
					return $scope.selectedbatch;
				}
			}
		});

	};
}

function ModalInstanceCtrl($scope, $modalInstance, batch) {
	$scope.gridOptions = {};
	$scope.batch = batch;
	console.log($scope.batch);
	$scope.gridOptions = {
		data : $scope.batch.options,
		columnDefs : [
				{
					name : 'name',
					displayName : 'Technical Name',
					enableCellEdit : false
				},
				{
					name : 'description',
					displayName : 'Description',
					enableCellEdit : false
				},
				{
					name : 'isMandatory',
					displayName : 'is Mandatory',
					visible : false
				},
				{
					name : 'defaultValue',
					displayName : 'Value',
					enableCellEdit : true,
					enableCellEditOnFocus : true,
					cellClass : function(grid, row, col, rowRenderIndex,
							colRenderIndex) {
						if (!grid.getCellValue(row, col) && row.entity.isMandatory) {
							return 'red';
						}
					}
				} ]
	};

	$scope.gridOptions.onRegisterApi = function(gridApi) {
		// set gridApi on scope
		$scope.gridApi = gridApi;
		gridApi.edit.on.afterCellEdit($scope, function(rowEntity, colDef,
				newValue, oldValue) {
			$scope.$apply();
		});
	};

	$scope.ok = function() {
		$modalInstance.close();
	};

	$scope.cancel = function() {
		$modalInstance.dismiss('cancel');
	};
};
