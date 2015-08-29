package hardwareutil
import (
    "os/exec"
    "regexp"
    "math"
    "strconv"
)

// GetTemperature return the temperature
func GetTemperature()float32{
    cmd := exec.Command("/opt/vc/bin/vcgencmd","measure_temp")
    if data,err := cmd.Output() ; err == nil {
        if reg,e := regexp.Compile("[0-9]{1,2}\\.[0-9]{0,2}") ; e == nil {

            strTemp := reg.FindString(string(data))
            if temperature,e := strconv.ParseFloat(strTemp,32) ; e == nil {
                return math.Ceil(temperature*10)/10
            }
        }
    }
    return 0
}
