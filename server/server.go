package main
import "net/http"
import "log"
import "os/exec"
import "os"
import "net"
//import "errors"
import "encoding/json"
import "time"
import "io/ioutil"
import "flag"
import "github.com/larspensjo/config"
import "strings"
import "fmt"

//var e error e = errors.New("Not Found")

var logger *log.Logger


type MyHandler struct {
     AllowIps []string
     AllowCmds []string
     AllowKey string
    }

type Response struct{
    Code int `json:"code"`
    Message string `json:"message"`
    Stdout string `json:"stdout"`
    Stderr string `json:"stderr"`
    /*
    b, err := json.Marshal(s)
    if err != nil {
        fmt.Println("json err:", err)
    }
    */
}


/*type Request struct{
    Cmd string
}*/
func FormatResult(res *Response, code int, message string, stdout string ,stderr string){
        res.Code = code
        res.Message = message
        res.Stdout = stdout
        res.Stderr = stderr
}

func in_array(val string, array []string) (exists bool, index int) {
    exists = false
    index = -1;

    for i, v := range array {
        if val == v {
            index = i
            exists = true
            return
        }
    }

    return
}

func ParseCmd(cmd string)(binary [1024]string){
    cmd = strings.Replace(cmd,"||","|",-1)
    cmd = strings.Replace(cmd,"&&","|",-1)
    cmd = strings.Replace(cmd,";","|",-1)
    binaryArgs := strings.Split(cmd, "|")
    for index, value := range binaryArgs{
        v := strings.Trim(value," ")
        cmdArray := strings.Split(v, " ")
        binary[index] = cmdArray[0]
    }
    return

}



func CheckSecurity(cmd string, key string, ip string, h *MyHandler)(ok bool, msg string){
    if key != h.AllowKey{
        ok = false
        msg = "Wrong Key"
        return
    }
    if exists, _ := in_array(ip, h.AllowIps); !exists{
        ok = false
        msg = "Ip is not allowed here"
        return
    }
    if h.AllowCmds[0] == "all"{
       ok = true
       msg = "ok"
       return
    }
    if cmd == ""{
        ok = false
        msg = "Command is nil"
        return
    }else{
        cmds := ParseCmd(cmd)
        for _, value := range cmds{
            if value == ""{
                continue
            }
            if e,_ := in_array(value, h.AllowCmds); !e{
                ok = false
                msg = "cmd is Not allowed: " + value
                return
            }
        }
    }
    msg = "ok"
    ok = true
    return
}

func(h *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
   // result = make(map[string]interface{}) may this do well?
    var result Response
   // var req Request
    if r.Method != "POST" {
        FormatResult(&result, 403, "Only POST method is allowed here.", "", "")

    }else if r.RequestURI != "/cmd" {
        FormatResult(&result,404, "Path is not found.","","")

    }else{
        /*body, _ := ioutil.ReadAll(r.Body)
        err := json.Unmarshal(body, &req)
        if err != nil {
            FormatResult(&result, 400, "Error occur while decode Post data as json.", "", "")
            goto Final
        } */
        r.ParseForm()
        fullCommand := r.Form.Get("cmd")
        key := r.Form.Get("key")
        clientIp, _ ,_ := net.SplitHostPort(r.RemoteAddr)
        logger.Println("From: " + clientIp + " Command: " + fullCommand)
        ok, msg := CheckSecurity(fullCommand, key, clientIp, h)
        if !ok{
            FormatResult(&result, 400, msg, "", "")
            goto Final
        }
        out, code , e := Execshell(fullCommand)
        if e != nil{
            FormatResult(&result, code, "Error occur while execute shell script:" + e.Error(), "", "")
            goto Final
        }
        if code == 0{
            FormatResult(&result, code, "ok", out, "")
        }else{
            FormatResult(&result, code, "ok", "", out)
        }

    }

Final:
    js, js_err := json.Marshal(result)
    if js_err != nil{
        logger.Println(js_err.Error())
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
}

func Execshell(myCmd string)(result string,code int,e error){
    result = ""
    e = nil
    code = 0
    cmd := exec.Command("/bin/sh", "-c", myCmd)
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        logger.Println("StdoutPipe: " + err.Error())
        code = -1
        e = err
        return
    }
    stderr, err := cmd.StderrPipe()
    if err != nil {
        logger.Println("StderrPipe: ", err.Error())
        code = -2
        e = err
        return
    }
    if err := cmd.Start(); err != nil {
        logger.Println("Start: ", err.Error())
        code = -3
        e = err
        return
    }
    bytesErr, err := ioutil.ReadAll(stderr)
    if err != nil {
        logger.Println("ReadAll stderr: ", err.Error())
        code = -5
        e = err
        return
    }
    if len(bytesErr) != 0 {
        logger.Println("stderr is not nil: %s", string(bytesErr))
        code = 2
        result = string(bytesErr)
        return
    }
    bytes, err := ioutil.ReadAll(stdout)
    if err != nil {
        logger.Println("ReadAll stdout: ", err.Error())
        code = -6
        e = err
        return
    }
    if err := cmd.Wait(); err != nil {
        logger.Println("Wait: ", err.Error())
        code = -4
        e  = err
        return
    }
    result = string(bytes)
    return
}

func main(){
    listen := flag.String("s",":8080","Listen address and port")
    confFile := flag.String("c","./control.ini","Control file")
    logFile := flag.String("l","./cmd_server.log","Log file")
    to := flag.Int("t",120,"Exec timeout")
    flag.Parse()
    timeout := time.Duration(time.Duration(*to)*time.Second)
    timeoutMsg := "Your request has timed out"
    logFilefd, err := os.OpenFile(*logFile,os.O_RDWR|os.O_CREATE|os.O_APPEND,0666)
    if err!=nil{
        fmt.Println("Error while Opening log file")
        os.Exit(1)
    }
    defer logFilefd.Close()
    logger = log.New(logFilefd,"\n",log.Ldate|log.Ltime)//|log.Llongfile
    c, _ := config.ReadDefault(*confFile)
    allowCommand, _ := c.String("default","allowcmd")
    allowCommandList := strings.Split(allowCommand," ")
    allowKey,_ := c.String("default","allowkey")
    allowIp,_ := c.String("default","allowip")
    allowIpList := strings.Split(allowIp," ")
    rootHandler := &MyHandler{AllowIps: allowIpList, AllowCmds: allowCommandList, AllowKey: allowKey}
    handler := http.TimeoutHandler(rootHandler, timeout, timeoutMsg)
    s := &http.Server{
        Addr:           *listen,
        Handler:        handler,
        ReadTimeout:   600 * time.Second,
        WriteTimeout:   600 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }
    logger.Println(s.ListenAndServe())
}
