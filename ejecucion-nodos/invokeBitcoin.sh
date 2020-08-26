#!/bin/bash

baseport=8330
offset=$1
rpcp=$(($baseport+$offset*2+1))
cport=$(($baseport+$offset*2))
scriptdir=/Users/mpizzagali/Tesis/btc-core/ejecucion-nodos
runningdir=$scriptdir/.exec-results/data-$1
rm -rf $runningdir
mkdir -p $runningdir
#bash $(dirname $(readlink -f $0))/bitcoindo.sh $1 stop 2>/dev/null
#Note: -daemon makes the process run on background
/Users/mpizzagali/Tesis/btc-core/src/bitcoind -regtest -daemon -port=$cport -rpcport=$rpcp -rpcuser=a -rpcpassword=b -datadir=$runningdir ${@:2}
