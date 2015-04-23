A http server wrote in golang.
=======

A http server wrote in golang, for excuting shell script and return result in json format with access control.
---

explantion for control.ini:
----
[default]
> 
;allowcmd defines the command we can use for shell script. If it reprents as all, then no limits with the command
> 
allowcmd = all
> 
;allowcmd = ls sleep awk grep zcat cat echo cut sed head sort uniq
> 
;allowkey defines the private access key, only the person who has this key can execute the script.
> 
allowkey = 3b93a59d36cf4b2a1s
> 
;allowip defiles the client ip which can fullfill the request.
> 
allowip = 127.0.0.1 
> 



Start:
---
Usage of ./server:
> 
  -c="./control.ini": Control file
> 
  -l="./cmd_server.log": Log file
> 
  -s=":8080": Listen address and port
> 
  -t=120: Exec timeout
> 

go run server.go -l="/tmp/server.log" -c="./control.ini"  -s=127.0.0.1:8081 -t=20



example:
---
> 
curl -X POST  --data-urlencode 'cmd=ls /tmp' -d "key=3b93a59d36cf4b2a194b4b3617f1f41c"  "http://127.0.0.1:8081/cmd"
> 

{"code":0,"message":"ok","stdout":"fcitx-socket-:0\ngedit.mingjie6.3633868441\ngo-build951207863\nhsperfdata_mdm\nhsperfdata_mingjie6\nicedteaplugin-mdm-UoQVqQ\nlibgksu-WvQEdZ\nmintUpdate\norbit-mingjie6\nplugtmp\nproxy.log\npulse-PKdhtXMmr18n\nserver.log\nsni-qt_sogou-qimpanel_2559-R53OMT\nsogou-qimpanel:0.pid\nsogou-qimpanelmingjie6\nssh-4tqQyYaluIDQ\nVMwareDnD\nvmware-mingjie6\nvmware-root\n","stderr":""}


