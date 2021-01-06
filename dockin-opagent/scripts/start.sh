#!/bin/bash
#
# Copyright (C) @2021 Webank Group Holding Limited
# <p>
# Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
# in compliance with the License. You may obtain a copy of the License at
# <p>
# http://www.apache.org/licenses/LICENSE-2.0
# <p>
# Unless required by applicable law or agreed to in writing, software distributed under the License
# is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
# or implied. See the License for the specific language governing permissions and limitations under
# the License.
#

################################
# start.sh
###############################

set -e;

function detect_pid_file(){
    for no in $(seq 1 30); do
		if [ -f ${APP_BIN}/sys.pid ]; then
			return 0;
		else
			echo "[$no] detect pid file ..."
			sleep 1s;
		fi
	done
	echo "detect pid file timeout."
	return 1;
}

function get_pid {
	local ppid=""
	if [ -f $APP_BIN/sys.pid ]; then
		ppid=$(cat $APP_BIN/sys.pid)
	fi
	echo "$ppid";
}

APP_BIN="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
APP_HOME="$(dirname $APP_BIN)"; [ -d "$APP_HOME" ] || { echo "ERROR failed to detect APP_HOME."; exit 1;}
APP_NAME=$(basename "$APP_HOME")
APP_START_OPTION=
if [ -f $APP_BIN/server.env ];then
        APP_START_OPTION=`cat $APP_BIN/server.env  | grep APP_START_OPTION:::  | awk -F ':::' {'print $2'}`
fi

# set up the logs dir
APP_LOG=$APP_HOME/logs
if [ ! -d ${APP_LOG} ];then
  mkdir ${APP_LOG}
fi

export APP_HOME=${APP_HOME}
cd $APP_HOME
subsystem=`ls apps`
cd $APP_BIN

pid=$(get_pid)
if [ -n "$pid" ];then
	echo -e "ERROR\t the server is already running (pid=$pid), there is no need to execute start.sh again."
	exit 9;
fi

if [[ "$DOCKER" == true ]]; then
    $APP_HOME/apps/${subsystem} ${APP_START_OPTION} >> $APP_LOG/${APP_NAME}.out 2>&1
else
    $APP_HOME/apps/${subsystem} ${APP_START_OPTION} >> $APP_LOG/${APP_NAME}.out 2>&1 &
fi

echo $! > sys.pid

if detect_pid_file; then
    ./watchdog.sh --add-crontab
fi


