/* Show log of a server */

if(Loader){Loader.toLoad("html/tasks.html","TasksPanel");}

var TasksPanel = {
    init:function(){
        $.extend(this,Panel)
        this.initPanel($('#idTasks'),'Tasks')
        this.div.resizable()
        this.panel = $('.tasks',this.div);
        var _self = this;
        this.div.bind('close',function(){
            _self.stop();
        })

        this.div.bind('open',function(){
            _self.start();
        })
    },
    stop:function(){

    },
    start:function(){
        // Check every second new data
        this._loadTasks();
        setTimeout(function(){
            TasksPanel._loadTasks();
        },2000)
    },
    _loadTasks:function(){
        $.ajax({
            url:'/allTasks',
            dataType:'json',
            success:function(data){
                TasksPanel._displayTasks(data);
            }
        })
    },
    _displayTasks:function(tasks){
        tasks.forEach(function(t){
            var line = $('div[data-id-task="' + t.Id + '"]',this.panel);
            if(line.length == 0){
                // New one
                line = $('<div class="task-info" data-id-task="' + t.Id + '">ID : ' + t.Id + '</div>')

                this.panel.append(line);
            }else{
                console.log("Already have, update line remove")
            }
            line.data('visited',true);
        },this)
        $('.task-info:not(:data("visited"))',this.panel).remove()
        $('.task-info').removeData('visited');
    }
}