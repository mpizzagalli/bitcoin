#cd /home/mgeier/ndecarli/db-4.8.30.NC/build_unix
#mkdir -p build
#BDB_PREFIX=$(pwd)/build
#../dist/configure --disable-shared --enable-cxx --with-pic --prefix=$BDB_PREFIX
#make CC=gcc-7 CXX=g++-7 install -j 8
#make install -j 8

./autogen.sh
./configure #CPPFLAGS="-I/home/mgeier/ndecarli/bdb/include/ -O2" LDFLAGS="-L/home/mgeier/ndecarli/bdb/lib/" --with-gui
make install -j 8
