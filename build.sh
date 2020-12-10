#!/bin/bash

RUN_NAME="ares.script.w2b_price_update"

mkdir -p output/bin output/conf

cp script/bootstrap.sh output 2>/dev/null
chmod u+x output/bootstrap.sh 
cp conf/* output/conf/

go build -o output/bin/${RUN_NAME}
chmod u+x output/bin/${RUN_NAME}
