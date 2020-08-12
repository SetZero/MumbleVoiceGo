package mumble

import (
	"MumbleSound/src/mumble/static/mumbleproto"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/hraban/opus.v2"
	"io"
	"os"
	"time"
)

const (
	Version             uint16 = 0
	UDPTunnel                  = 1
	Authenticate               = 2
	Ping                       = 3
	Reject                     = 4
	ServerSync                 = 5
	ChannelRemove              = 6
	ChannelState               = 7
	UserRemove                 = 8
	UserState                  = 9
	BanList                    = 10
	TextMessage                = 11
	PermissionDenied           = 12
	ACL                        = 13
	QueryUsers                 = 14
	CryptSetup                 = 15
	ContextActionModify        = 16
	ContextAction              = 17
	UserList                   = 18
	VoiceTarget                = 19
	PermissionQuery            = 20
	CodecVersion               = 21
	UserStats                  = 22
	RequestBlob                = 23
	ServerConfig               = 24
	SuggestConfig              = 25
)

var voicePackageCounter = 0

func sendProtobufData(conn *tls.Conn, data []byte, message uint16) {
	packageType := make([]byte, 2)
	length := make([]byte, 4)

	binary.BigEndian.PutUint16(packageType, message)
	binary.BigEndian.PutUint32(length, uint32(len(data)))

	writeableData := append(packageType[:], append(length[:], data...)...)
	fmt.Println(writeableData)
	_, err := conn.Write(writeableData)
	if err != nil {
		fmt.Println("There was an error sending the Version!")
	}
}

func sendData(conn *tls.Conn, data protoreflect.ProtoMessage, message uint16) {
	out, err := proto.Marshal(data)
	if err != nil {
		fmt.Println("Failed to create a Version Message: ", err)
	}
	sendProtobufData(conn, out, message)
}

func readData(conn *tls.Conn) {
	packageType := int(binary.BigEndian.Uint16(readBytes(conn, 2)))

	length := readBytes(conn, 4)
	readableLength := int(binary.BigEndian.Uint32(length))

	payload := readBytes(conn, readableLength)
	routeData(packageType, payload)
}

func readBytes(conn *tls.Conn, bytes int) []byte {
	data := make([]byte, bytes)
	_, err := io.ReadFull(conn, data)
	if err != nil {
		fmt.Printf("Failed to Read %d Bytes", bytes)
	}
	return data
}

func StartConnection() {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", "nooblounge.net:64738", conf)
	if err != nil {
		fmt.Print("There was an error while trying to connect to the server!")
	}
	defer conn.Close()

	fmt.Println("Starting Data Exchange")
	exchangeVersion(conn)
	sendAuth(conn)
	initPing(conn)
	doVoiceStuff(conn)
	joinChannel(conn, 1069) //1069
	for {
		readData(conn)
	}
}

func exchangeVersion(conn *tls.Conn) {
	os := "GoClient"
	release := "1.3.0"
	osversion := "11"
	v := uint32((1 << 16) | (3 << 8))
	version := &mumbleproto.Version{Os: &os, Release: &release, OsVersion: &osversion, Version: &v}

	sendData(conn, version, Version)
}

func sendAuth(conn *tls.Conn) {
	opus := true
	versionlist := []int32{-2147483637, -2147483632}
	username := "GoClient"
	auth := &mumbleproto.Authenticate{Opus: &opus, CeltVersions: versionlist, Username: &username}

	sendData(conn, auth, Authenticate)
}

func initPing(conn *tls.Conn) chan struct{} {
	ticker := time.NewTicker(15 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				timestamp := uint64(time.Now().Unix())
				ping := &mumbleproto.Ping{Timestamp: &timestamp}
				sendData(conn, ping, Ping)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return quit
}

func joinChannel(conn *tls.Conn, channel int) {
	channelId := uint32(channel)
	channelChange := &mumbleproto.UserState{ChannelId: &channelId}
	sendData(conn, channelChange, UserState)
}

func doVoiceStuff(conn *tls.Conn) {
	ticker := time.NewTicker(50 * time.Millisecond)

	f, err := os.Open("test.raw")
	if err != nil {
		fmt.Println("There was an error while reading the pcm data")
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				sendVoiceData(conn, f)
			}
		}
	}()
}

func sendVoiceData(conn *tls.Conn, f *os.File) {
	const sampleRate = 48000
	const channels = 1

	header := []byte{byte(0x80)}
	counterVar := make([]byte, 0)
	counterVar = makeVarInt(uint64(voicePackageCounter), counterVar)
	voicePackageCounter += 1

	enc, err := opus.NewEncoder(sampleRate, channels, opus.AppVoIP)
	if err != nil {
		fmt.Println("There was an error creating the encoder")
	}

	const bufferSize = 2880

	pcm := make([]int16, bufferSize)
	err = binary.Read(f, binary.LittleEndian, &pcm)
	fmt.Println(err)
	if err == io.EOF {
		f.Seek(0, 0)
	}

	frameSize := len(pcm)
	frameSizeMs := float32(frameSize) / channels * 1000 / sampleRate
	switch frameSizeMs {
	case 2.5, 5, 10, 20, 40, 60:
		break
	default:
		fmt.Printf("Illegal frame size: %d bytes (%f ms)", frameSize, frameSizeMs)
		return
	}

	payload := make([]byte, bufferSize)
	n, err := enc.Encode(pcm, payload)
	if err != nil {
		fmt.Println("There was an error while encoding the data!")
	}
	payload = payload[:n]
	length := uint64(len(payload) & 0x1FFF)
	payload = append(makeVarInt(length, make([]byte, 0)), payload...)
	full := append(header, append(counterVar[:], payload...)...)
	sendProtobufData(conn, full, UDPTunnel)
}

func routeData(packageType int, payload []byte) {
	var data protoreflect.ProtoMessage

	switch uint16(packageType) {
	case Version:
		data = &mumbleproto.Version{}
	case CryptSetup:
		data = &mumbleproto.CryptSetup{}
	case ChannelState:
		data = &mumbleproto.ChannelState{}
	case UserState:
		data = &mumbleproto.UserState{}
	case ServerSync:
		data = &mumbleproto.ServerSync{}
	}

	if data == nil {
		return
	}
	if err := proto.Unmarshal(payload, data); err != nil {
		fmt.Println("Unable to Unmarshal")
	}
	fmt.Println(data)
}

func makeVarInt(i uint64, buff []byte) []byte {
	if i < 0x80 {
		// Need top bit clear
		buff = append(buff, byte(i))
	} else if i < 0x4000 {
		// Need top two bits clear
		buff = append(buff, byte((i>>8)|0x80))
		buff = append(buff, byte(i&0xFF))
	} else if i < 0x200000 {
		// Need top three bits clear
		buff = append(buff, byte((i>>16)|0xC0))
		buff = append(buff, byte((i>>8)&0xFF))
		buff = append(buff, byte(i&0xFF))
	} else if i < 0x10000000 {
		// Need top four bits clear
		buff = append(buff, byte((i>>24)|0xE0))
		buff = append(buff, byte((i>>16)&0xFF))
		buff = append(buff, byte((i>>8)&0xFF))
		buff = append(buff, byte(i&0xFF))
	} else if i < 0x100000000 {
		// It's a full 32-bit integer.
		buff = append(buff, 0xF0)
		buff = append(buff, byte((i>>24)&0xFF))
		buff = append(buff, byte((i>>16)&0xFF))
		buff = append(buff, byte((i>>8)&0xFF))
		buff = append(buff, byte(i&0xFF))
	} else {
		// It's a 64-bit value.
		buff = append(buff, 0xF4)
		buff = append(buff, byte((i>>56)&0xFF))
		buff = append(buff, byte((i>>48)&0xFF))
		buff = append(buff, byte((i>>40)&0xFF))
		buff = append(buff, byte((i>>32)&0xFF))
		buff = append(buff, byte((i>>24)&0xFF))
		buff = append(buff, byte((i>>16)&0xFF))
		buff = append(buff, byte((i>>8)&0xFF))
		buff = append(buff, byte(i&0xFF))
	}
	return buff
}
