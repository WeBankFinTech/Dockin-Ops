#!/usr/bin/env python
# -*- coding: utf-8 -*-
# author: johnsonhhu



import sys
import ConfigParser
import logging
import logging.config



def is_sysid(rn):
    if len(rn) == 4 and rn.isdigit():
        return True
    return False


def is_sysname(rn):
    temp = rn.split('-')
    if len(temp) == 2:
        if len(temp[0]) > 0 and temp[0].isalpha():
            if len(temp[1]) > 0:# and temp[1].isalpha():
                return True
    return False


def is_ip(rn):
    temp = rn.split('.')
    result = False
    n = 0
    if len(temp) == 4:
        for i in temp:
            if i.isdigit() and int(i) >= 0 and int(i) <= 255:
                n = n + 1
                if n == 4:
                    result = True
            else:
                result = False
                break
    return result


def isdcn(rn):
    if len(rn) == 3:
        return True
    return False


def is_pod_set_id(rn):
    temp = rn.split('-')
    if is_sysid(temp[0]) and temp[3].isdigit():
        if isdcn(temp[2]):
            return True
    return False


def is_podname(rn):
    temp = rn.split('-')
    # pod_set_id的格式
    if len(temp) == 4:
        if is_pod_set_id("%s-%s-%s-%s" % (temp[0], temp[1], temp[2], temp[3])):
            return True
    # pod_set_id在测试环境的格式
    if len(temp) == 5 and temp[4] == '0':
        if env == 'UAT' or env == 'SIT':
            if is_pod_set_id("%s-%s-%s-%s" % (temp[0], temp[1], temp[2], temp[3])):
                return True
    # 测试环境
    if env == 'UAT' or env == 'SIT':
        if len(temp) > 2:
            if is_sysname("%s-%s" % (temp[0], temp[1])):
                return True
    # 生产环境
    if env == 'PRD':
        if len(temp) == 6:
            if is_sysname("%s-%s" % (temp[0], temp[1])):
                if is_ip("%s.%s.%s.%s" % (temp[2], temp[3], temp[4], temp[5])):
                    return True
    return False


def cmdb_pn(pn):
    if env == 'UAT' or env == 'SIT':
        if pn[-2:] == '-0':
            temp = pn.split('-')
            if len(temp) == 6 and is_ip("%s.%s.%s.%s" % (temp[2], temp[3], temp[4], temp[5])):
                return pn
            else:
                return pn[:-2]
    return pn


def k8s_pn(pn):
    if type(pn) == str:
        return k8s_pn_1(pn)
    elif type(pn) == list:
        if len(pn) == 2:
            return k8s_pn_2(pn[0], pn[1])
    return pn


def k8s_pn_2(pn, ns):
    if env == 'UAT' or env == 'SIT':
        if ns == 'monitoring' or ns == 'kube-system':
            return pn
        else:
            return k8s_pn_1(pn)
    return pn


def k8s_pn_1(pn):
    if env == 'UAT' or env == 'SIT':
        if pn[-2:] != '-0':
            return "%s-0" % pn
        else:
            temp = pn.split('-')
            if len(temp) == 6 and is_ip("%s.%s.%s.%s" % (temp[2], temp[3], temp[4], temp[5])):
                return "%s-0" % pn
    return pn


def rule_args():
    if rule == 'default' or len(rule) == 0:
        return ''
    return " -r %s" % rule


def list_2_string(l):
    if len(l) > 1:
        s = l[0]
        for i in l[1:]:
            s = "%s %s" % (s, i)
        return s
    else:
        return ""


def get_name_space(l):
    flag = 0
    for i in l:
        if flag == 1:
            return i
        if i.upper() == "-N":
            flag = 1
        if i.upper() == "--ALL-NAMESPACE" or i.upper() == "--ALL-NAMESPACES":
            return "all"
    return "null"


def new_logger(log_file_name):
    handler = logging.handlers.TimedRotatingFileHandler(log_file_name, when='D', interval=1, backupCount=10, encoding=None, delay=False, utc=False)
    fmt = '%(asctime)s - %(threadName)s - %(levelname)s - %(funcName)s - %(message)s'
    formatter = logging.Formatter(fmt)
    handler.setFormatter(formatter)
    logger = logging.getLogger('commonlogger')
    logger.addHandler(handler)
    logger.setLevel(logging.DEBUG)
    return logger


sshdocker_logger = new_logger("%s/sshdocker.log" % sys.path[0])
cf = ConfigParser.ConfigParser()
cf.read("%s/commontool.conf" % sys.path[0])



env = cf.get("common", "env").upper()# 环境标识 UAT or PRD
try:
    rule = cf.get("common", "rule").lower()
except ConfigParser.NoOptionError:
    rule = 'default'
opsctl_name = cf.get("common", "opsctl_name").lower()
