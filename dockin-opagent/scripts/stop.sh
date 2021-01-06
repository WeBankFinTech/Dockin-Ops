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
# stop.sh
###############################

function get_pid {
    local ppid=""
	if [ -f $APP_BIN/sys.pid ]; then
		ppid=$(cat $APP_BIN/sys.pid)
	fi
    echo "$ppid";
}

APP_BIN="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
APP_NAME=$(basename "$APP_HOME")
APP_HOME="$(dirname $APP_BIN)"; [ -d "$APP_HOME" ] || { echo "ERROR! failed to detect APP_HOME."; exit 1;}

${APP_BIN}/watchdog.sh --delete-crontab
pid=$(get_pid)
if [ -z "$pid" ];then
	echo -e "INFO\t the server does NOT start yet, there is no need to execute stop.sh."
	exit 0;
fi
kill -15 $pid;
stop_timeout=30
for no in $(seq 1 $stop_timeout); do
    if ps $PS_PARAM -p "$pid" 2>&1 > /dev/null; then
        if [ $no -lt $stop_timeout ]; then
            echo "[$no] shutdown server ..."
            sleep 1
            continue
        fi
        echo "shutdown server timeout, kill process: $pid"
        kill -9 $pid; sleep 1; break;
    else
        echo "shutdown server ok!"; break;
    fi
done
if [ -f $APP_BIN/sys.pid ]; then
	rm $APP_BIN/sys.pid
fi
