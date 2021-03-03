#!/usr/bin/env python
# -*- coding: utf-8 -*-
# author: johnsonhhu

import json
import sys
import os
import prettytable
import commontool as ct, cmdbdocker
from commontool import sshdocker_logger as logger
import wcsssh



class OpsctlCommander():

    def __init__(self, args_list):
        self.command_path = sys.path[0]

        self.args_list = args_list # 输入参数 列表格式
        self.args_string = ct.list_2_string(args_list)

        self.name_space = ct.get_name_space(self.args_list)
        self.command_type = "COMMAND ERROR"
        self.resource_name = ''
        self.exec_cmd = ''

    def print_usage(self):
        print '''
        query container instance
        usage ：wcsctl cmdb <subsys_name|subsys_id|host_ip|pod_ip|pod_name> [dcn_list]              
               
        login container (interactive mode)
        usage ：wcsctl ssh [pod_name|pod_ip]
        
        exec cmd (remote command mode)
        usage ：wcsctl exec <pod_name|pod_ip> "cmd" [-n namespace]
        '''


    def exec_pod(self):
        if ct.is_podname(self.resource_name):
            self.resource_name = ct.k8s_pn(self.resource_name)
        if self.name_space == "null":
            cmd = '%s/%s%s exec %s -- /bin/sh -c "%s" 2>&1' % (self.command_path, ct.opsctl_name, ct.rule_args(), self.resource_name, self.exec_cmd)
        else:
            cmd = '%s/%s%s exec %s -n %s -- /bin/sh -c "%s" 2>&1' % (self.command_path, ct.opsctl_name, ct.rule_args(), self.resource_name, self.name_space, self.exec_cmd)
        logger.debug(cmd)
        os.system(cmd)

    def ssh(self):
        s = wcsssh.WcsSsh(self.resource_name)
        s.run()

    def analysis_subcommand(self):
        # print self.args_list
        if len(self.args_list) > 1:
            if self.args_list[1].upper() == "EXEC":
                if self.name_space == "null" and len(self.args_list) == 4:
                    if ct.is_podname(self.args_list[2]) or ct.is_ip(self.args_list[2]):
                        self.resource_name = self.args_list[2]
                        self.exec_cmd = self.args_list[3]
                        return "EXEC"
                elif self.name_space != "null" and len(self.args_list) == 6:
                    if self.args_list[2].upper() == "-N":
                        i, j = (4, 5)
                    elif self.args_list[3].upper() == "-N":
                        i, j = (2, 5)
                    elif self.args_list[4].upper() == "-N":
                        i, j = (2, 3)
                    else:
                        return "COMMAND ERROR"
                    self.resource_name = self.args_list[i]
                    self.exec_cmd = self.args_list[j]
                    if ct.is_ip(self.resource_name) or ct.is_podname(self.resource_name):
                        return "EXEC"
                    else:return "COMMAND ERROR"
            if self.args_list[1].upper() == "SSH":
                if len(self.args_list) == 3:
                    self.resource_name = self.args_list[2]
                return "SSH"
            if self.args_list[1].upper() == "CMDB":
                return "CMDB"
        return "COMMAND ERROR"

    def cmdb(self, al):
        cd = cmdbdocker.CmdbDocker()
        result = cd.run(al[1:])
        if result["Code"] != 0:
            if result["Code"] == 1 and result["Message"] == "Parameter Error":
                print result["Data"].replace("cmdbdocker", "wcsctl cmdb")
            else:
                print result["Message"]
        else:
            print result["Data"]

    def run(self):
        self.command_type = self.analysis_subcommand()
        if self.command_type == "SSH":
            self.ssh()
        elif self.command_type == "EXEC":
            self.exec_pod()
        elif self.command_type == "CMDB":
            self.cmdb(self.args_list)
        else:
            self.print_usage()



if __name__ == '__main__':
    logger.info(ct.list_2_string(sys.argv))
    oc = OpsctlCommander(sys.argv)
    oc.run()
