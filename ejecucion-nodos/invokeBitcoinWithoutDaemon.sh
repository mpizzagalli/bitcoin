#!/bin/bash

baseport=8330
offset=$1
rpcp=$(($baseport+$offset*2+1))
cport=$(($baseport+$offset*2))
runningdir=$(dirname $(readlink -f $0))/.exec-results/data-$1
rm -rf $runningdir
mkdir -p $runningdir
#bash $(dirname $(readlink -f $0))/bitcoindo.sh $1 stop 2>/dev/null
#Note: -daemon makes the process run on background
bitcoind -regtest -port=$cport -rpcport=$rpcp -rpcuser=a -rpcpassword=b -datadir=$runningdir ${@:2}
