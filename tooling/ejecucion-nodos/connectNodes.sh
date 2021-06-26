baseport=8330
offset=$2
port=$(($baseport+$offset*2))
ip=127.0.0.1

# scriptdir=/Users/mpizzagali/Tesis/btc-core/tooling/ejecucion-nodos
scriptdir=/home/mgeier/mpizzagalli/bitcoin/tooling/ejecucion-nodos
if [ $# -gt 2 ]
then
	ip=$3
fi
bash $scriptdir/bitcoindo.sh $1 addnode ${ip}:${port} add