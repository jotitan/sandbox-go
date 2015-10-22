angular.module('angular-ace-editor', [])
    .directive('aceEditor', ['$rootScope', function ($rootScope) {

        var lastId = 0;

        return {
            restrict: 'A',
            require:'?ngModel',
            link: function (scope, element, attrs,ngModel) {
                var id = 'ace-editor-' + (++lastId);
                // set unique id
                element.attr('id', id);

                // emit object out
                var editor = ace.edit(id);
                // When object is well loaded
                ngModel.$render = function(){
                    editor.setValue(ngModel.$viewValue)
                }

                editor.on('blur',function(){
                    ngModel.$setViewValue(editor.getValue())
                })
                $rootScope.$emit('aceEditor.init', editor);
            }
        };
    }]);
