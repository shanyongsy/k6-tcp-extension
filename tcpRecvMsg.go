package tcp

import (
	"log"

	"git.shiyou.kingsoft.com/shanyong/xk6-tcp/lib"
	"git.shiyou.kingsoft.com/shanyong/xk6-tcp/pb"
	"google.golang.org/protobuf/proto"
)

// 返回收发消息的对应关系
func msgMapping(recvMsg string) string {
	var ret string = ""
	switch recvMsg {
	case "G2CVerifyMessage":
		ret = "C2GVerifyMessage"
	case "G2CTestMessage":
		ret = "C2GTestMessage"
	case "G2CHeartbeatMessage":
		ret = "C2GHeartbeatMessage"
	}
	return ret
}

func (client *Client) G2CVerifyMessage(data *[]byte) bool {

	var msg = pb.G2CVerifyMessage{}

	err := proto.Unmarshal(*data, &msg)
	if err != nil {
		log.Printf("%v Err: %v\n", lib.GetFuncName(), err)
		return false
	}

	return msg.Result == 0
}

func (client *Client) G2CHeartbeatMessage(data *[]byte) bool {
	var msg = pb.G2CHeartbeatMessage{}

	err := proto.Unmarshal(*data, &msg)
	if err != nil {
		log.Printf("%v Err: %v\n", lib.GetFuncName(), err)
		return false
	}

	return msg.Milli != 0
}

func (client *Client) G2CTestMessage(data *[]byte) bool {
	var msg = pb.G2CTestMessage{}

	err := proto.Unmarshal(*data, &msg)
	if err != nil {
		log.Printf("%v Err: %v\n", lib.GetFuncName(), err)
		return false
	}

	return msg.Result == 0
}

func (client *Client) G2CNotifyMessage(data *[]byte) bool {
	var msg = pb.G2CNotifyMessage{}

	err := proto.Unmarshal(*data, &msg)
	if err != nil {
		log.Printf("%v Err: %v\n", lib.GetFuncName(), err)
		return false
	}

	return msg.Message != ""
}
