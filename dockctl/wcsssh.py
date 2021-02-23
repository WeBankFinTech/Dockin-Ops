#!/usr/bin/env python
# -*- coding: utf-8 -*-
# author: johnsonhhu

import os
import json
import sys
import commontool as ct, cmdbdocker


class WcsSsh():

    def __init__(self, resource_name):
        self.command_path = sys.path[0]
        self.user_name = ''
        self.user_pwd = ''
        self.rule = ct.rule
        self.token = ''
        self.subsystem = ""
        self.dcn = ""
        self.namespace = ""
        self.pod_list_all = ""
        self.pod_list_dcn = ""
        self.cmdb_string_table = ""
        self.chosen = -1
        self.resource_name = resource_name.lower()
        self.flag = 'common'
        if len(self.resource_name) > 0:
            if ct.is_podname(self.resource_name) or ct.is_ip(self.resource_name):
                self.flag = 'once'

    def analysis_input(self, args):
        pass

    def check_user(self):
        cmd = '%s/%s auth%s -u %s -p %s' % (self.command_path, ct.opsctl_name, ct.rule_args(), self.user_name, self.user_pwd)
        result = json.loads(os.popen(cmd).read())
        if result['Code'] == 0 and result['Message'] == 'Success' and len(result['Data']) > 0:
            print "login with user %s success" % self.user_name
            self.token = result['Data']
        else:
            print result['Message']
        return

    def input_user(self):
        self.user_name = os.popen("who am i | awk '{print $1}'").read()[:-1]
        self.user_pwd = "wcsssh"

    def input_namespace(self):
        while True:
            namespace = raw_input("input namespace : ")
            if len(namespace) > 0:
                self.namespace = namespace
                break

    def input_dcn(self):
        dcn = raw_input("input dcn : ")
        self.dcn = dcn
        self.query_dcn()

    def run_ssh_pod(self):
        if self.flag == 'number':
            if len(self.pod_list_dcn) > 0:
                cmd = '%s/%s ssh %s --access-token %s' % (self.command_path, ct.opsctl_name, self.get_pod_from_pod_list(self.pod_list_dcn, self.chosen), self.token)
            else:
                cmd = '%s/%s ssh %s --access-token %s' % (self.command_path, ct.opsctl_name, self.get_pod_from_pod_list(self.pod_list_all, self.chosen), self.token)
        if self.flag == 'pod' or self.flag == 'once':
            if ct.is_podname(self.resource_name):
                self.resource_name = ct.k8s_pn(self.resource_name)
            cmd = '%s/%s ssh %s --access-token %s' % (self.command_path, ct.opsctl_name, self.resource_name, self.token)
        #print cmd
        os.system(cmd)

    def string_table_add_index(self, string_table):
        count = 0
        result_string = ""
        for i in string_table.split('\n'):
            if count == 0:
                result_string = "Index  %s" % i
            else:
                k = 4
                if count >= 10:
                    k = 3
                elif count >= 100:
                    k = 2
                elif count >= 1000:
                    k = 1
                elif count >= 10000:
                    k = 0
                result_string = "%s\n%s  %s" % (result_string, "%d%s" % (count, " " * k),i)
            count = count + 1
        return result_string

    def get_pod_from_pod_list(self, p_string, p_num):
        return p_string.split('\n')[p_num].split()[2]

    def query_cmdb(self):
        cd = cmdbdocker.CmdbDocker()
        result = cd.run(["cmdb",self.subsystem])
        if result["Code"] == 0 and result["Data"] is not None:
            self.cmdb_string_table = result["Data"].get_string(fields=["State", "Dcn", "PodIp", "PodName", "HostIp", "Version"])
            self.pod_list_all = self.string_table_add_index(self.cmdb_string_table)
            self.pod_list_dcn = ""
            self.dcn = ""
        else:
            print "No Container App Instance found"

    def query_dcn(self):
        result_count = 0
        is_data = 0
        result_string = ""
        for i in self.cmdb_string_table.split('\n'):
            if is_data == 0:
                is_data = 1
                result_string = i
            else:
                if i.split()[0].lower() == self.dcn.lower():
                    result_count = result_count + 1
                    result_string = "%s\n%s" % (result_string, i)
        if result_count == 0:
            self.dcn = ""
            self.pod_list_dcn = ""
        else:
            self.pod_list_dcn = self.string_table_add_index(result_string)

    def display_pod(self):
        #print self.pod_list_dcn
        #print self.pod_list_all
        if len(self.pod_list_dcn) > 0:
            print self.pod_list_dcn
        elif len(self.pod_list_all) > 0:
            print self.pod_list_all

    def input_common_command(self):
        print "valid input : pod_name | pod_ip | subsystem_name | \"exit\""
        if len(self.pod_list_all) > 0:
            print "valid input : choose container : index number | choose dcn : \"dcn\""
        while True:
            input_cmd = raw_input("input your chosen : ").strip()
            if len(input_cmd) == 0:
                continue
            elif input_cmd.lower() == "dcn":
                if len(self.pod_list_all) > 0:
                    self.flag = "dcn"
                    break
            elif input_cmd.lower() == "exit":
                self.flag = "exit"
                break
            elif ct.is_sysname(input_cmd):
                self.subsystem = input_cmd
                self.flag = "subsys"
                break
            elif input_cmd.isdigit():
                if input_cmd > 0:
                    if (self.dcn == "" and len(self.pod_list_all.split('\n')) > int(input_cmd)) or (self.dcn != "" and len(self.pod_list_dcn.split('\n')) >= int(input_cmd)):
                        self.chosen = int(input_cmd)
                        self.flag = "number"
                        break
            elif ct.is_podname(input_cmd) or ct.is_ip(input_cmd):
                self.resource_name = input_cmd.lower()
                self.flag = 'pod'
                break

    def run(self):
        self.input_user()
        self.check_user()
        while True:
            if self.flag == 'common':
                self.display_pod()
                self.input_common_command()
            if self.flag == 'exit':
                break
            if self.flag == 'subsys':
                self.query_cmdb()
                self.flag = 'common'
            if self.flag == 'dcn':
                self.input_dcn()
                self.flag = 'common'
            if self.flag == 'number' or self.flag == 'pod':
                self.run_ssh_pod()
                self.flag = 'common'
            if self.flag == 'once':
                self.run_ssh_pod()
                self.flag = 'exit'
        return
