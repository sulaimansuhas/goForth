package pktparser

import (
	// "archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	// "strings"
)


type dnsmessage struct{
	Header Header
	Questions Questions
	Answers ResourceRecord
	Authority ResourceRecord
	Additional ResourceRecord

}

type Header struct{
	Id uint16
	Info uint16
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}
type Questions struct{
	QName string
	Qinfo questionInfo
}

type questionInfo struct {
	QType uint16
	QClass uint16
}

type ResourceRecord struct{
	Name string
	RRInfo ResourceRecordInfo
}

type ResourceRecordInfo struct{
	Type uint16
	Class uint16
	Ttl uint32

}

//func Serialize(message *dnsmessage) (output []byte){
//	
//}

func Deserialize(input []byte) (Header, Questions, ResourceRecord){

	label_pos_map:= make(map[int]string)
	r := bytes.NewReader(input)
	var head Header
	var question Questions
	var answer ResourceRecord
	var pos int
	//var qs questions
	if err:= binary.Read(r,binary.BigEndian,&head);err!=nil{
		fmt.Printf("Reading into header struct resulted in error %v", err)
		return head,question,answer
	}
	pos = int(r.Size()) - r.Len()
	question.QName, _ = decodeDomainName(r, label_pos_map, "", true) //Need to handle error here
	label_pos_map[pos]=question.QName

	if err:= binary.Read(r,binary.BigEndian,&question.Qinfo);err!=nil{
		fmt.Printf("Reading into question info struct resulted in error %v", err)
		return head,question,answer
	}

	pos = int(r.Size()) - r.Len()
	answer.Name, _ = decodeDomainName(r, label_pos_map, "", true)
	label_pos_map[pos]=answer.Name
	if err:= binary.Read(r,binary.BigEndian,&answer.RRInfo);err!=nil{
		fmt.Printf("Reading into answer resource record info struct resulted in error %v", err)
		return head,question,answer
	}

	return head,question,answer
	
}


//Assuming if we get a pointer in the compression scheme that also indicates the end of the label 
func decodeDomainName(input *bytes.Reader, label_map map[int]string, decoded_string string, first bool) (string, error){
	var byte1 byte
	var byte2 byte
	var err error
	if byte1,err= input.ReadByte();err!=nil{
		fmt.Printf("Reading length octet resulted in error %v", err)
		return "", err
	}

	if byte1 == 0 {
		return decoded_string,nil

	}
	if !first{
		decoded_string = decoded_string + "."
	}
	if byte1 > 63 {
		byte1 = byte1 & byte(63) 
		if byte2,err= input.ReadByte();err!=nil{
			fmt.Printf("Reading length octet resulted in error %v", err)
			return "", err
		}
		combined := uint16(byte1) << 8 | uint16(byte2)
		decoded_string = decoded_string + label_map[int(combined)]
		return decoded_string,nil
	}

	stringbuf := make([]byte, byte1)
	if _,err:= input.Read(stringbuf);err!=nil{
		fmt.Printf("Reading word octet resulted in error %v", err)
		return "", err
	}

	decoded_string = decoded_string + fmt.Sprintf("%s", stringbuf)

	return decodeDomainName(input, label_map, decoded_string, false)
}
