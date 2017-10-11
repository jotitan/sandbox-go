package main
import (
    "os"
    "strconv"
    "task_manager"
    "logger"
)

/* Launch a node server used to treat tasks sended by task manager */
func main(){
    if len(os.Args) != 4 {
        logger.GetLogger().Info("Need following parameters : urlManager prefix (d:\\) capacity")
        os.Exit(1)
    }
    urlManager := os.Args[1]
    prefix := os.Args[2]
    capacity := 1
    if value,err := strconv.ParseInt(os.Args[3],10,32) ; err == nil {
        capacity = int(value)
    }
    task_manager.LaunchNodeServer(urlManager, prefix, capacity)
}
