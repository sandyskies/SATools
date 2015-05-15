#-*-coding:utf-8-*-
#!/usr/bin/env python
import paramiko
import sys
import os
import Queue
import threading

GLOBAL_QUEUE = Queue.Queue()
GLOBAL_LOCK = threading.Lock()
SSH_PORT = 22

class thread_main(threading.Thread):
    def __init__(self, cmd, log_file, user, passwd, pkey):
        self.cmd = cmd
        self.user = user
        self.passwd = passwd
        self.log_file = log_file
        self.pkey = pkey
        self.run()


    def ssh_connect(self, ip):
        client = paramiko.SSHClient()
        client.load_system_host_keys()
        client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        pkeyfile = os.path.expanduser(self.pkey)
        mykey = paramiko.RSAKey.from_private_key_file(pkeyfile)
        if self.passwd != '':
            client.connect(ip, port=SSH_PORT, username=self.user,password=self.passwd, timeout=2)
        else:
            client.connect(ip,port=SSH_PORT, username=self.user, timeout=2, pkey=mykey)
        return client
    def exec_cmd(self,eachip):
        try:
            client = self.ssh_connect(eachip)
            i, o, e = client.exec_command(self.cmd)
            error = e.read()
            if error:
                print "ip:%s errors happen:%s" %(eachip, error)
                log = error
                log = eachip + ":\n" + log + "\n"
            else:
                log = o.read()
                log = eachip + ":\n" + log + "\n"
            GLOBAL_LOCK.acquire()
            f = open(self.log_file, "a")
            f.write(log)
            f.close()
            GLOBAL_LOCK.release()
            client.close()
        except Exception, err:
            print "ip:%s errors happen: %s" %(eachip, err)
    def run(self):
        while True:
            if GLOBAL_QUEUE.empty():
                break
            ip = GLOBAL_QUEUE.get()
            self.exec_cmd(ip)
            GLOBAL_QUEUE.task_done()


class remote_command:
    def __init__(self, ip_file, log_file, threads, cmd, user='root', passwd='', pkey=''):
        self.ip_file = ip_file
        self.log_file = log_file
        self.threads = threads
        self.cmd = cmd
        self.user = user
        self.passwd = passwd
        self.pkey = pkey

    def get_ip_list(self, ip_file):
        with open(ip_file,'r') as fp:
            allip = fp.readlines()
            fp.close()
        return allip

        
    def start(self):
        allip = self.get_ip_list(self.ip_file)
        for ip in allip:
            if ip.strip():
                GLOBAL_QUEUE.put(ip.strip())
        for thread in range(self.threads):
            thread_main(self.cmd, self.log_file, self.user, self.passwd, self.pkey)

        GLOBAL_QUEUE.join()
        os._exit(0)


if __name__ == '__main__':
    #example
    c = remote_command("./ip.txt", log_file="./test.log",threads=5, cmd='ls',user='root', passwd='123')
    c.start()

    
