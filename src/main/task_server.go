package main
import "task_manager"

func main(){
    server := task_manager.NewManagerServer()
    server.Start()

}
