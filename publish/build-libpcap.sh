# see https://github.com/google/gopacket/issues/100

export PCAPV=1.8.1
wget http://www.tcpdump.org/release/libpcap-$PCAPV.tar.gz

mkdir libpcap-arm
mkdir libpcap-i386
#ar xvf libpcap-$PCAPV.tar.gz -C libpcap-arm
tar xvf libpcap-$PCAPV.tar.gz -C libpcap-i386

cd libpcap-arm
#cd libpcap-$PCAPV
#CC=arm-linux-gnueabi-gcc ./configure --host=arm-linux --with-pcap=linux
#make

cd libpcap-$PCAPV
#export CC=gcc
#./configure --host

#
