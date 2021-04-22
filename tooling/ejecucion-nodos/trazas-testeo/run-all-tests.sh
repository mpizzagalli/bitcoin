#!/bin/bash

scriptdir=/Users/mpizzagali/Tesis/btc-core/tooling/ejecucion-nodos

# bash $scriptdir/run.sh -n 2 -s 1 -f test-caso-1 -m trace -i $scriptdir/trazas-testeo/caso-1-shh.in

for filename in $scriptdir/trazas-testeo/*.in; do
    date
    echo bash $scriptdir/run.sh -n 2 -s 1 -f test-$(basename $filename .in) -m trace -o -i $filename
    bash $scriptdir/run.sh -n 2 -s 1 -f test-$(basename $filename .in) -m trace -o -i $filename
done