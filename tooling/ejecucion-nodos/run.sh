#!/bin/bash
# Levanta n nodos, simulando lo realizado por el .fog pero localmente.
# Esto no incluye ningun tipo de monitoreo ya que las cosas andan localmente
#
# Parametros:
# - Numero de nodos (sin incluir el del historial) que desee levantar
# - Numero de nodos con selfish mining que desee levantar (este numero debe ser menor al de arriba)
# - Nombre de la corrida (por default se pone test)
# - Modo en que se quiere correr, si es libre en base a los hash rates asignados ("free") o siguiendo una traza dada ("trace")
# - path a la traza dada

print_usage() {
  printf "Usage: bash run.sh -n nodes -s selfishNodes -f runName -m mode -o [-i traceFileIn]\n"
}

# Constantes
gobin=go
# scriptdir=/Users/mpizzagali/Tesis/btc-core/tooling/ejecucion-nodos
scriptdir=/home/mgeier/mpizzagalli/bitcoin/tooling/ejecucion-nodos
datadir=$scriptdir/.exec-results/data-0
addressesdir=$scriptdir/.exec-results
logdir=$scriptdir/logs

#Parametros entrada
numnodes=1
numselfishnodes=0
runName="test"
mode="free"
traceFileIn=""
writeTrace=false

while getopts 'n:s:f:m:oi:h' flag; do
  case "${flag}" in
    n) numnodes=$OPTARG ;;
    s) numselfishnodes=$OPTARG ;;
    f) runName="$OPTARG" ;;
    m) mode="${OPTARG}" ;;
    o) writeTrace=true ;;
    i) traceFileIn="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

stopAllBTCNodes() {
    echo 'Señal de matar a los nodos recibida, matando a los BTC nodes levantados...'
    for (( nodeId=0; nodeId<$numnodes; nodeId++ ))
    do
        bash $scriptdir/bitcoindo.sh $nodeId stop
    done
    sleep 1m

    exit 0
}

startLogAt=$(($numnodes*2400))

runlogdir=$scriptdir/logs/$(date +'%Y%m%d%H%M%S')-$runName

echo "Going to create the folder $runlogdir to save the results inside for nodes $numnodes"

mkdir $runlogdir

echo "Empiezo $numnodes nodo(s) BTC"
for (( nodeId=0; nodeId<$numnodes; nodeId++ ))
do
    if (($nodeId<$numselfishnodes)); then
        bash $scriptdir/invokeBitcoin.sh $nodeId -dificulta=0 -dbcache=3072 -start-log-at=$startLogAt -log-folder=$runlogdir -mining-mode=1
    else
        bash $scriptdir/invokeBitcoin.sh $nodeId -dificulta=0 -dbcache=3072 -start-log-at=$startLogAt -log-folder=$runlogdir -mining-mode=0
    fi
done
sleep 30s # Esperamos a que termine el booteo

echo "Esperamos a que los nodos se estabilicen"
for (( nodeId=0; nodeId<$numnodes; nodeId++ ))
do
    $gobin run $scriptdir/semaphore.go $nodeId
done

echo "Conectamos los nodos cada uno con el otro (por ahora en forma de clique)" # TODO aca agregar algo para no conectar tipo clique
for (( node1Id=0; node1Id<$numnodes-1; node1Id++ ))
do
    for (( node2Id=node1Id+1; node2Id<$numnodes; node2Id++ ))
    do
        echo "Vamos a conectar los nodos $node1Id y $node2Id"
        bash $scriptdir/connectNodes.sh $node1Id $node2Id localhost
    done
done

echo "Comienzo el nodo que genera el historial y lo conecto al resto de los nodos"
bash $scriptdir/invokeBitcoin.sh $numnodes -dificulta=0 -dbcache=3072 -log-folder=$runlogdir -debug-tesis
sleep 30s
echo "Vamos a esperar que el nodo que genera el historial se estabilice"
$gobin run $scriptdir/semaphore.go $numnodes
for (( nodeId=0; nodeId<$numnodes; nodeId++ ))
do
    bash $scriptdir/connectNodes.sh $nodeId $numnodes localhost
done

echo "Genero addresses para darle fondos a cada uno de los nodos"
for (( nodeId=0; nodeId<$numnodes; nodeId++ ))
do
    rm -f $addressesdir/addrN$nodeId
done
for (( nodeId=0; nodeId<$numnodes; nodeId++ ))
do
    bash $scriptdir/bitcoindo.sh $nodeId getnewaddress > $addressesdir/addrN$nodeId
    bash $scriptdir/bitcoindo.sh $nodeId getnewaddress >> $addressesdir/addrN$nodeId
done

echo "Generamos una blockchain inicial"
$gobin run $scriptdir/generateBlockchain.go $numnodes
sleep 30s

echo "Esperamos que los nodos se estabilicen despues de recibir la blockchain inicial"
for (( nodeId=0; nodeId<$numnodes; nodeId++ ))
do
    $gobin run $scriptdir/semaphore.go $nodeId
done

echo "Mato al nodo del historial"
bash $scriptdir/bitcoindo.sh $numnodes stop
sleep 1m

starttime=$(date +%s)

if [ "$mode" = "free" ]; then
    echo "Disparo el motor que genero bloques en cada uno de los nodos"
    nodeshp=$(bc -l <<< "1.0/$numnodes") # TODO: aca cambiar para tener hash rates asimetricos
    for (( nodeId=0; nodeId<$numnodes; nodeId++ ))
    do
        if [ "$writeTrace" = true ] ; then
            traceFileOut=$runlogdir/traceN$nodeId.out
            $gobin run $scriptdir/launcher.go $gobin run $scriptdir/minerEngine.go $nodeId $nodeshp $traceFileOut $starttime
        else
            $gobin run $scriptdir/launcher.go $gobin run $scriptdir/minerEngine.go $nodeId $nodeshp
        fi
        sleep 1s
    done
elif [ "$mode" = "trace" ]; then
    echo "Muevo la traza de entrada a la carpeta de los logs y la renombro trace.in"
    cp $traceFileIn $runlogdir/trace.in
    echo "Disparo el motor que reproduce la traza pasada por parametro"
    if [ "$writeTrace" = true ] ; then
        traceFileOut=$runlogdir/traceGlobal.out
        $gobin run $scriptdir/traceEngine.go $traceFileIn $numnodes $traceFileOut $starttime
    else 
        $gobin run $scriptdir/traceEngine.go $traceFileIn $numnodes
    fi
    echo "Termine la traza"
    # echo "dejo todo levantado para seguir probando"
    stopAllBTCNodes
    exit 0
else 
    echo "Invalid mode quedan los nodos levantados"
fi

# Paro a los nodos BTC ante un SIGINT o un SIGTERM
trap stopAllBTCNodes SIGINT SIGTERM
# sleep infinity
sleep 86400
