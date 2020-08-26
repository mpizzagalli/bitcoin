baseport=8330
offset=$1
rpcp=$(($baseport+$offset*2+1))
bitcoin-cli -regtest -rpcport=$rpcp -rpcuser=a -rpcpassword=b ${@:2}