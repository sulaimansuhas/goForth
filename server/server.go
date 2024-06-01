package main

import (
	"dnsf/pktparser"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type cached_record struct{
	time_stamp time.Time
	ttl int
	answer []byte

}


//TODO:
//Better Naming
//Multi thread the two different sockets
func main(){
	var head pktparser.Header
	var question pktparser.Questions
	var answer pktparser.ResourceRecord
	id_to_ip := make(map[uint16]net.UDPAddr)
	cached_answers := make(map[string]cached_record)

	messageBuffer := make([]byte, 1024)
	defaults := net.UDPAddr{
		IP: net.ParseIP("127.0.0.1"),
		Port: 8080,
	}
	google_addr:= net.UDPAddr{
		IP: net.ParseIP("8.8.8.8"),
		Port: 53,
	}
	connection, err := net.DialUDP("udp", nil, &google_addr)
	server, err := net.ListenUDP("udp", &defaults)
	defer server.Close()
	if err != nil {
		fmt.Printf("Error  %v on Socket %d", err, defaults.Port)
	}
	fmt.Println("Server running......")
	for {
		n, addr, _:= server.ReadFromUDP(messageBuffer)

		fmt.Printf("Read %d bytes from %s.\n",n, addr.String())

		fmt.Printf("Message: %s\n", messageBuffer[:n])
		headerBuffer := make([]byte, n)
		copy(headerBuffer,messageBuffer[:n])

		head,question,answer= pktparser.Deserialize(headerBuffer)

		id_to_ip[head.Id]=*addr
		val, exists := cached_answers[question.QName]
		if exists{
			if (time.Since(val.time_stamp)).Seconds() <= float64(val.ttl) {
				fmt.Printf("Found cached value, ttl is %d, timestamp is %s, current time is %s\n", val.ttl, val.time_stamp.Format("15:04:05.000"), time.Now().Format("15:04:05.0000"))
				temp:= make([]byte, len(val.answer))
				newid := make([]byte,2)
				binary.BigEndian.PutUint16(newid, head.Id)
				copy(temp, newid)
				copy(temp[2:],val.answer[2:])
				add := (id_to_ip[head.Id])
				server.WriteToUDP(temp, &(add))
				continue
			}
		}
		_,err = connection.Write(headerBuffer)
		fmt.Printf("Forwarded UDP\n")
		n,_,err = connection.ReadFromUDP(messageBuffer)
		copy(headerBuffer,messageBuffer[:n])
		head,question,answer= pktparser.Deserialize(headerBuffer)
		cached_answer := cached_record{
			time_stamp: time.Now(),
			ttl: int(answer.RRInfo.Ttl),
		}
		cached_answer.answer = make([]byte, n)
		copy(cached_answer.answer, headerBuffer)
		cached_answers[question.QName] = cached_answer

		add := (id_to_ip[head.Id])
		server.WriteToUDP(headerBuffer, &(add))
		
	}


}
