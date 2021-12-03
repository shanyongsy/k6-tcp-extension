package tcp

import (
	"log"
	"time"

	"git.shiyou.kingsoft.com/shanyong/xk6-tcp/lib"
	"git.shiyou.kingsoft.com/shanyong/xk6-tcp/pb"
	"google.golang.org/protobuf/proto"
)

func (client *Client) C2GVerifyMessage() error {
	// println(lib.GetFuncName())
	cmd := lib.GetFuncName()
	msg := &pb.C2GVerifyMessage{Uuid: client.uuid, Token: client.token}

	body, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("%v, %v.\n", cmd, err)
		return err
	}

	err = client.sendMsgToGateway(&cmd, &body)
	if err != nil {
		log.Fatalf("%v, %v.\n", cmd, err)
		return err
	}

	return nil
}

func (client *Client) C2GTestMessage() error {
	// println(lib.GetFuncName())
	cmd := lib.GetFuncName()
	msg := &pb.C2GTestMessage{Test: time.Now().String()}

	body, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("%v, %v.\n", cmd, err)
		return err
	}

	err = client.sendMsgToGateway(&cmd, &body)
	if err != nil {
		log.Printf("%v, %v.\n", cmd, err)
		return err
	}

	return nil
}

func (client *Client) C2GHeartbeatMessage() error {
	// println(lib.GetFuncName())
	cmd := lib.GetFuncName()
	msg := &pb.C2GHeartbeatMessage{Milli: time.Now().UnixMicro() / int64(time.Millisecond)}

	body, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("%v, %v.\n", cmd, err)
		return err
	}

	err = client.sendMsgToGateway(&cmd, &body)
	if err != nil {
		log.Printf("%v, %v.\n", cmd, err)
		return err
	}

	return nil
}
