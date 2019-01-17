package main

import (
	"os"
	"io/ioutil"
	"fmt"
	"time"
	"strconv"
)

const timeFormat = "15:04:05.999999"
var nodeNumber int64

func readPingsLogFile() []byte {
	if b, err := ioutil.ReadFile(fmt.Sprintf("%spingLogN%d",os.Args[2], nodeNumber)); err == nil {
		return b
	} else {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse pings log file.\n %s\n", err.Error()))
		return nil
	}
}

func createFile() (scriptFile *os.File) {
	var err error

	if scriptFile, err = os.Create(fmt.Sprintf("%spingLogN%d",os.Args[1], nodeNumber)); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file "+os.Args[1]+"pingLogN\n %s\n", err.Error()))
	}

	return scriptFile
}

func decodePacket(b []byte) (p int64) {
	p = int64(b[7])
	p |= int64(b[6]) << 8
	p |= int64(b[5]) << 16
	p |= int64(b[4]) << 24
	p |= int64(b[3]) << 32
	p |= int64(b[2]) << 40
	p |= int64(b[1]) << 48
	p |= int64(b[0]) << 56
	return p
}

func decodePort(b []byte) (p int32) {
	p = int32(b[1])
	p |= int32(b[0]) << 8
	return p
}

func writeLineToFile(file *os.File, content string) {
	if _, err := file.Write([]byte(content+"\n")); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed writing to %s.\n %s\n", file.Name(), err.Error()))
	}
}

func decodeLog(log []byte, scriptFile *os.File) {
	
	var sentTime time.Time
	var receiveTime time.Time
	var senderPort int32
	maxIndex := len(log)-18
	var delay int64

	for i:=0; i<=maxIndex; i+=18 {
		sentTime = time.Unix(0, decodePacket(log[i:i+8]))
		receiveTime = time.Unix(0, decodePacket(log[i+8:i+16]))
		senderPort = decodePort(log[i+16:i+18])
		delay = (receiveTime.Sub(sentTime).Nanoseconds()+500000)/1000000
		if delay > 1000 {
			fmt.Println(fmt.Sprintf("Ping Alto: %d en paquete del host %d al host %d at %s", delay, senderPort, nodeNumber, receiveTime.Format("02-01T15:04:05.999-07")))
		}
		writeLineToFile(scriptFile, fmt.Sprintf("Packet from node at port %d had a delay of %dms at %s", senderPort, delay, receiveTime.Format("02-01T15:04:05.999-07")))
	}
}

func main(){

	if len(os.Args) < 4 {
		os.Stderr.WriteString("Missing File names as parameters.\n")
		return
	}

	nodeNumber, _ = strconv.ParseInt(os.Args[3], 10, 64)

	log := readPingsLogFile()

	scriptFile := createFile()

	decodeLog(log, scriptFile)
}
