package main

import (
	"os"
	"io/ioutil"
	"fmt"
	"time"
)

const timeFormat = "15:04:05.999999"

func readPingsLogFile() []byte {
	if b, err := ioutil.ReadFile(os.Args[2]); err == nil {
		return b
	} else {
		os.Stderr.WriteString(fmt.Sprintf("Failed to parse pings log file.\n %s\n", err.Error()))
		return nil
	}
}

func createFile() (scriptFile *os.File) {
	var err error

	if scriptFile, err = os.Create(os.Args[1]); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Failed to create file "+os.Args[1]+"\n %s\n", err.Error()))
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

	for i:=0; i<=maxIndex; i+=18 {
		sentTime = time.Unix(0, decodePacket(log[i:i+8]))
		receiveTime = time.Unix(0, decodePacket(log[i+8:i+16]))
		senderPort = decodePort(log[i+16:i+18])
		writeLineToFile(scriptFile, fmt.Sprintf("Packet from node at port %d sent at %s and received at %s", senderPort, sentTime.Format(timeFormat), receiveTime.Format(timeFormat)))
	}
}

func main(){

	if len(os.Args) < 3 {
		os.Stderr.WriteString("Missing File names as parameters.\n")
		return
	}

	log := readPingsLogFile()

	scriptFile := createFile()

	decodeLog(log, scriptFile)
}
