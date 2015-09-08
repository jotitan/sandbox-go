/* Show tasks from server */

if(Loader){Loader.toLoad("html/tasks.html","TasksPanel");}

var TasksPanel = {
    tasksSSE:null,
    init:function(){
        $.extend(this,Panel)
        this.initPanel($('#idTasks'),'<span class="glyphicon glyphicon-list-alt icon"></span>Tasks')
        this.div.resizable()
        this.panel = $('.tasks > tbody',this.div);
        var _self = this;
        this.div.bind('close',function(){
            _self.stop();
        })

        this.div.bind('open',function(){
            _self.start();
        })
    },
    stop:function(){
       this.panel.empty();
       if(this.tasksSSE!=null){
        this.tasksSSE.close();
        this.tasksSSE = null;
       }
    },
    start:function(){
        // Use SSE to get all tasks
        this.tasksSSE = new EventSource('/allTasksAsSSE')
        this.tasksSSE.onmessage = function(data){
            TasksPanel._displayTasks(JSON.parse(data.data))
        }
        this.tasksSSE.onerror = function(){
            TasksPanel.stop();
            console.log("=>Error")
        }
    },
    _getStatus:function(status){
        var results = ["",""];
        switch(status){
            case 0:results[0]="New";results[1] = "blue";break;
            case 1:results[0]="Running";results[1] = "orange";break;
            case 2:results[0]="End";results[1] = "green";break;
            default:results[0]="Error";results[1] = "red";break;
        }
        return results;
    },
    _displayTasks:function(tasks){
        tasks.forEach(function(t){
            var line = $('tr[data-id-task="' + t.Id + '"]',this.panel);
            var status = TasksPanel._getStatus(t.Status);
            var time = 0;
            if (new Date(t.StartTime).getTime() > 0){
                var time = (new Date(t.EndTime) - new Date(t.StartTime))/1000;
                if (time < 0){
                    var time = (new Date() - new Date(t.StartTime))/1000;
                }
                time = Math.round(time*10)/10;
            }
            if(line.length == 0){
                // New one
                line = $('<tr class="task-info" data-id-task="' + t.Id + '"><td>' + t.Id +'</td>' +
                '<td>' + t.TypeTask + '</td><td class="time">' + time + ' s</td>' +
                '<td class="status" style="color:' + status[1] + '">' + status[0] + '</td></tr>')
                this.panel.prepend(line);
            }else{
                $('td.status',line).css('color',status[1]).html(status[0]);
                $('td.time',line).html(time + " s");
                // Update, go up
                this.panel.prepend(line)
            }
        },this)
    }
}