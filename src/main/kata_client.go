package main
import (
    "net/http"
    "fmt"
    "logger"
    "io/ioutil"
    "time"
    "sync"
)

func main(){
    logger.GetLogger().Info("Begin")
    begin := time.Now()
    nb := 10000
    c := make(chan struct{},250)
    wg := sync.WaitGroup{}
    wg.Add(nb)
    for i := 0 ; i < nb ; i++ {
        go func(value int) {
            c <- struct {}{}
            if resp, err := http.Get(fmt.Sprintf("http://localhost:8081/request?value=%d", value)); resp.StatusCode == 200 {
                data, _ := ioutil.ReadAll(resp.Body)
                logger.GetLogger().Info("Get results", string(data))
            }else {
                if err == nil {
                    logger.GetLogger().Error("Get special status code",resp.StatusCode)
                }else{
                    logger.GetLogger().Fatal("Panic stop")
                }
            }
            <-c
            wg.Done()
        }(i)
    }
    wg.Wait()
    logger.GetLogger().Info("Time",time.Now().Sub(begin))
}
