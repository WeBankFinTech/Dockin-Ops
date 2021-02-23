#!/usr/bin/env python
# -*- coding: utf-8 -*-
# author: johnsonhhu

import os
import json
import sys
import prettytable
import commontool as ct
from commontool import sshdocker_logger as logger


class CmdbDocker():

    def __init__(self):
        self.command_path = sys.path[0]
        self.cmd = ''
        self.resource_type = ''
        self.resource_name = ''
        self.dcn_list = []
        self.result = {"Code": -1,
                       "Message": "Unknown Error",
                       "Data": "Unknown Error"
                       }
        self.usage = '''
        用法 : cmdbdocker < 子系统名称 | 子系统编号 | 母机ip | 容器ip | 容器名称 > [ dcn列表 ]

               dcn列表支持模糊匹配，可以使用 "1C1" "1D" "1D,1E" "1" ... 等形式
        
        示例 : cmdbdocker gns-query 1C2
        示例 : cmdbdocker 5036 1C1
        示例 : cmdbdocker 10.106.11.11
        示例 : cmdbdocker gns-query-10-106-11-11
        '''

    def analysis_resource_name(self):
        if ct.is_sysid(self.resource_name):
            self.resource_type = 'SYSID'
        elif ct.is_sysname(self.resource_name):
            self.resource_type = 'SYSNAME'
        elif ct.is_ip(self.resource_name):
            self.resource_type = 'IP'
        elif ct.is_podname(self.resource_name):
            self.resource_type = 'PODNAME'
        else:
            self.result = {"Code": 1,
                           "Message": "Parameter Error",
                           "Data": self.usage
                           }

    def analysis_dcn_list(self):
        if self.result["Code"] == 1 and self.result["Message"] == "Parameter Error":
            pass
        else:
            for i in range(0,len(self.dcn_list)):
                if len(self.dcn_list[i]) > 3:
                    self.result = {"Code": 1,
                                   "Message": "Parameter Error",
                                   "Data": self.usage
                                   }
                    break
                self.dcn_list[i] = self.dcn_list[i].upper()

    def analysis_input(self, args):
        if len(args) == 2 or len(args) == 3:
            self.resource_name = args[1]
            if len(args) == 3:
                self.dcn_list = args[2].split(",")
            self.analysis_resource_name()
            self.analysis_dcn_list()
        else:
            self.result = {"Code": 1,
                           "Message": "Parameter Error",
                           "Data": self.usage
                           }

    def generate_prettytable(self):
        if self.result['Code'] == 0:
            if self.result['Data'] is None:
                self.result["Code"] = 2
                self.result["Message"] = "No Container App Instance found"
            else:
                table = prettytable.PrettyTable()
                table.field_names = ['NameSpace',
                                     'SubSysId',
                                     'SubSysName',
                                     'Dcn',
                                     'PodIp',
                                     'PodName',
                                     'Version',
                                     'CPU',
                                     'MEM',
                                     'State',
                                     'HostIp',
                                     'Cluster']
                for i in self.result['Data']:
                    for n in i:
                        if len(i[n]) == 0:
                            i[n] = '-'
                    row = [i['Namespace'].lower(),
                           i['SubSysId'],
                           i['SubSysName'],
                           i['Dcn'],
                           i['PodIp'],
                           i['PodName'],
                           i['Version'],
                           i['LimitCpu'],
                           i['LimitMem'],
                           i['State'],
                           i['HostIp'],
                           i['ClusterId']]
                    if len(self.dcn_list) == 0:
                        table.add_row(row)
                    else:
                        for d in self.dcn_list:
                            if d in i['Dcn']:
                                table.add_row(row)
                                break
                table.border = False
                table.align = 'l'
                table.sortby = 'SubSysName'
                self.result["Data"] = table

    def run(self, args):
        self.analysis_input(args)
        if self.result["Code"] == 1 and self.result["Message"] == "Parameter Error":
            pass
        else:
            cmd = '%s/%s%s list' % (self.command_path, ct.opsctl_name, ct.rule_args())
            if self.resource_type == 'SYSID':
                cmd = '%s subsysId %s' % (cmd, self.resource_name)
            elif self.resource_type == 'SYSNAME':
                cmd = '%s subsysName %s' % (cmd, self.resource_name.lower())
            elif self.resource_type == 'IP':
                cmd = '%s ip %s' % (cmd, self.resource_name)
            elif self.resource_type == 'PODNAME':
                cmd = '%s podName %s' % (cmd, ct.k8s_pn(self.resource_name.lower()))
            logger.debug(cmd)
            self.result = json.loads(os.popen(cmd).read())
            logger.debug(self.result)
            self.generate_prettytable()
        return self.result



if __name__ == '__main__':
    logger.info(ct.list_2_string(sys.argv))
    cd = CmdbDocker()
    result = cd.run(sys.argv)
    if result["Code"] != 0:
        if result["Code"] == 1 and result["Message"] == "Parameter Error":
            print result["Data"]
        else:
            print result["Message"]
    else:
        print result["Data"]