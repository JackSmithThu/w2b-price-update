#!/bin/bash
CURDIR=$(cd $(dirname $0); pwd)
CONF_DIR=$CURDIR/conf/
CONF_FILE="${CONF_DIR}/config.json"

if [ ! -f $CONF_FILE  ]; then
    echo "there is no config.json exist at: ${CONF_FILE}"
    exit -1
fi

SERVICE_NAME=$(cat $CONF_FILE | grep '"service_name"' | awk -F':' '{print $2}' | grep -Eo "(\w+|\.)+")

if [ -z "$SERVICE_NAME" ]; then
    echo "Need service_name in ${CONF_FILE}"
    exit -1
fi
LOG_DIR=$CURDIR/logs/
PORT=$(cat $CONF_FILE | grep '"port"' | awk -F':' '{print $2}' | grep -Eo "(\w+|\.)+")
args_count=$#
if [ $args_count -ge 2 ] ; then
    if [ $1 == "--log_dir" ] ; then
        LOG_DIR=$2
    else
        echo "log_dir error, usage: $0 --log_dir dirname --port portnum"
        exit -1
    fi
    if [ $args_count -ge 4 ] ; then
        if [ $3 == "--port" ] ; then
            PORT=$4
        else
            echo "port error, usage: $0 --log_dir dirname --port portnum"
            exit -1
        fi
    fi
fi

args="-service_name=$SERVICE_NAME -conf_dir=$CONF_DIR -log_dir=$LOG_DIR"
if [ "X$PORT" != "X" ]; then
    args+=" -port=$PORT"
fi

echo "$CURDIR/bin/${SERVICE_NAME} $args"

exec $CURDIR/bin/${SERVICE_NAME} $args
