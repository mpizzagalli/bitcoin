package main

import (
	"time"
	"net"
	"strconv"
	"os"
	"math/rand"
)

const basePort int = 11000

type PingPacket int64

var Log []byte

func AddLogEntry(TimestampReceived []byte, TimeOfReception []byte, SenderNode uint16) {
	Log = append(Log, TimestampReceived...)
	Log = append(Log, TimeOfReception...)
	Log = append(Log, byte(SenderNode >> 8))
	Log = append(Log, byte(SenderNode))
}

func NodeToPort(n int) int {
	return basePort + n
}

func PortToNode(p int) uint16 {
	return  uint16(p - basePort)
}

func encodePacket(p int64) []byte {
	b := make([]byte, 8)
	b[0] = byte(p >> 56)
	b[1] = byte(p >> 48)
	b[2] = byte(p >> 40)
	b[3] = byte(p >> 32)
	b[4] = byte(p >> 24)
	b[5] = byte(p >> 16)
	b[6] = byte(p >> 8)
	b[7] = byte(p)
	return b
}

func decodePacket(b []byte) (p PingPacket) {
	p = PingPacket(b[7])
	p |= PingPacket(b[6]) << 8
	p |= PingPacket(b[5]) << 16
	p |= PingPacket(b[4]) << 24
	p |= PingPacket(b[3]) << 32
	p |= PingPacket(b[2]) << 40
	p |= PingPacket(b[1]) << 48
	p |= PingPacket(b[0]) << 56
	return p
}

func OpenPort(nodeNumber int) (c *net.UDPConn) {

	var addr net.UDPAddr
	addr.Port = NodeToPort(nodeNumber)
	addr.IP = nil
	c, _ = net.ListenUDP("udp4", &addr)
	return c
}

func ListenIncomingPackets(c *net.UDPConn) {

	logFile, _ := os.Create("pingLogN" + os.Args[2])
	Log = make([]byte, 0, 8192)

	buff := make([]byte, 8)
	var senderAddr *net.UDPAddr
	var receptionTime int64
	lastFlushTime := time.Now().UnixNano()
	for {
		_, senderAddr, _ = c.ReadFromUDP(buff)
		receptionTime = time.Now().UnixNano()
		AddLogEntry(buff, encodePacket(receptionTime), PortToNode(senderAddr.Port))
		if len(Log) >= 8190 || time.Duration(receptionTime - lastFlushTime) > time.Minute * 5 {
			flushBufer(logFile)
			lastFlushTime = receptionTime
		}
	}
}

func flushBufer(logFile *os.File){
	if _, err := logFile.WriteString(string(Log)); err == nil {
		Log = make([]byte, 0, 8192)
	}
}

func main(){

	if len(os.Args) < 3 {
		os.Stderr.WriteString("Missing node number as argument.\n")
		return
	}

	nodeNumber, _ := strconv.ParseInt(os.Args[2], 10, 64)

	c := OpenPort(int(nodeNumber))

	go ListenIncomingPackets(c)

	listeningNodesAmnt, _ := strconv.ParseInt(os.Args[1], 10, 64)

	nodeRndGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	var r int64
	var targetIp []net.IP
	var addr net.UDPAddr
	var b []byte

	listeningNodesAmnt-- // no queremos mandarnos paquetes a nosotros mismos

	sleepRndGen := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		r = nodeRndGen.Int63n(listeningNodesAmnt)

		if r >= nodeNumber {r++}

		if targetIp, _ = net.LookupIP("n"+strconv.FormatInt(r, 10)); len(targetIp) == 0 {
			targetIp, _ = net.LookupIP("localhost")
		}

		addr.Port = NodeToPort(int(r))
		addr.IP = targetIp[0]

		b = encodePacket(time.Now().UnixNano())

		c.WriteToUDP(b, &addr)

		r = sleepRndGen.Int63n(1024)

		time.Sleep(time.Millisecond * (1000 + time.Duration(r)))
	}

}
