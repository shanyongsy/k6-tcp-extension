package tcp

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"

	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"google.golang.org/protobuf/proto"

	"git.shiyou.kingsoft.com/shanyong/xk6-tcp/pb"
	"github.com/google/uuid"
)

func init() {
	modules.Register("k6/x/tcp", new(TCP))
}

// TCP is the k6 tcp extension.
type TCP struct{}

// Client is the TCP Client wrapper.
type Client struct {

	// conn
	conn net.Conn

	// address and port, for example, 127.0.0.1:8000.
	connStr string

	// The last error from this struct.
	lastErr error

	// The function pointer used to receive the message needs to be passed in JS.
	jsCallBackHandler JSCallBack

	// The uuid of client
	uuid string

	// The token of client
	token string

	// Information used to cache incoming and outgoing messages
	msgData MsgDataManager
}

// XClient represents the Client constructor (i.e. `new tcp.Client()`) and
// returns a new TCP client object.
func (r *TCP) XClient(ctxPtr *context.Context) interface{} {
	rt := common.GetRuntime(*ctxPtr)
	return common.Bind(rt, &Client{}, ctxPtr)
}

// Create new tcp.
// A new TCP link must be used, otherwise it will cause the same conn to communicate.
//func (tcp *Client) Create() *Client {
//	return new(Client)
//}

// Send the received message to JS. Com through this function.
// cmd - msg id
// sus -resault
type JSCallBack func(cmd *string, sus bool)

// To init all things.
func (client *Client) Connect(addr string, handler JSCallBack) error {
	client.connStr = addr
	client.uuid = uuid.New().String()
	client.token = client.uuid
	client.jsCallBackHandler = handler
	client.conn, client.lastErr = net.Dial("tcp", client.connStr)
	if client.lastErr != nil {
		client.conn = nil
		log.Printf("client connect err. set conn nil.\n")
		return client.lastErr
	} else {
		go client.loopRecv()
	}

	return nil
}

func (client *Client) TearDown() {
	defer func() {
		log.Printf("%v, TearDown\n", client.uuid)

		if err := recover(); err != nil {
			log.Printf("TearDown err. client=%v\n, err is %v.\n", client.uuid, err)
		}
	}()

	if client.conn != nil {
		client.conn.Close()
	}
}

func (client *Client) GetID() string {
	return client.uuid
}

// Send msg by this function.
// func (client *Client) WriteStr(data string) error {

// 	if client.conn == nil {
// 		return errors.New("call Write function, but conn is nil.")
// 	}

// 	_, err := client.conn.Write([]byte(data))
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// Send msg by this function.
// func (client *Client) WriteStrLn(data string) error {
// 	return client.WriteStr(fmt.Sprintln(data))
// }

// Get msg by this function.
func (client *Client) loopRecv() {
	defer func() {
		log.Printf("client %v, loopReadConn out.\n", client.uuid)

		if err := recover(); err != nil {
			log.Printf("loopReadConn err. client=%v\n, err is %v.\n", client.uuid, err)
		}
	}()

	for {

		conn := client.conn

		if conn == nil {
			break
		}

		lenBuf := make([]byte, 2)
		_, err := io.ReadFull(conn, lenBuf)
		if err != nil {
			client.conn = nil
			if err == io.EOF {
				log.Printf("client %v quit...set conn nil\n", client.uuid)
			} else {
				log.Printf("client %v read length error: %v\n", client.uuid, err.Error())
			}
			break
		}
		pkgLen := binary.BigEndian.Uint16(lenBuf[0:])
		buf := make([]byte, pkgLen)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			client.conn = nil
			log.Printf("read package error: %v. set client %v conn nil.\n", err.Error(), client.uuid)
			break
		}

		onRecvMsg(client, &buf)

		//time.Sleep(time.Duration(1) * time.Millisecond)
	}
}

func (client *Client) toJsCallBack(cmd string, sus bool) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("toJsCallBack err. client=%v, cmd=%v, sus=%v, call=%v, %v\n", client.uuid, cmd, sus, client.jsCallBackHandler, err)
		}
	}()

	// log.Printf("toJsCallBack. client=%v, cmd=%v, sus=%v\n", client.uuid, cmd, sus)

	// if client.jsCallBackHandler == nil {
	// 	log.Printf("toJsCallBack handler is nil. client=%v, cmd=%v, sus=%v, call=%v\n", client.uuid, cmd, sus, client.jsCallBackHandler)
	// }

	// if client.jsCallBackHandler != nil {
	// 	client.jsCallBackHandler(&cmd, sus)
	// }
}

func (client *Client) sendMsgToGateway(cmd *string, body *[]byte) error {

	frameMsg := &pb.FrameMessage{Cmd: *cmd, Body: *body}
	msg := &pb.GatewayPack{GatewayMessage: frameMsg, BusinessMessage: nil}

	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	buffer := make([]byte, len(data)+2)

	binary.BigEndian.PutUint16(buffer[0:], uint16(len(data)))
	copy(buffer[2:], data)
	_, err = client.conn.Write(buffer)

	return err
}

func onRecvMsg(client *Client, data *[]byte) {
	var cmd string = ""
	var sus bool = false

	msg := &pb.GatewayPack{}
	err := proto.Unmarshal(*data, msg)
	if err != nil {
		log.Printf("onRecveMsg Err: %v\n", err)
		return
	}

	if msg == nil || msg.GetGatewayMessage() == nil {
		log.Println("onRecveMsg Err: msg == nil || msg.GetGatewayMessage()")
		return
	}

	if msg.GetGatewayMessage().Cmd == "" {
		log.Println("onRecveMsg Err: msg.GetGatewayMessage().Cmd == ''")
		return
	}

	cmd, sus = dealRecvMsg(client, msg.GatewayMessage)

	// log.Printf("onRecveMsg client=%v conn=%v cmd=%v sus=%v\n", client.uuid, client.conn.LocalAddr(), cmd, sus)

	retCmd := msgMapping(cmd)

	if retCmd != "" {
		client.msgData.StoreRecvMsg(&retCmd, sus)

		client.toJsCallBack(retCmd, sus)
	}
}

// 返回 - param1:G2C cmdID, param2:sus
func dealRecvMsg(client *Client, msg *pb.FrameMessage) (string, bool) {

	cmd := msg.Cmd

	value := reflect.ValueOf(client)
	f := value.MethodByName(cmd)
	if !f.IsValid() || f.Kind() != reflect.Func {
		log.Printf("onDealMsg, No handler found for message %v.\n", cmd)
		return cmd, false
	}

	in := make([]reflect.Value, 1)
	in[0] = reflect.ValueOf(&msg.Body)
	ret := f.Call(in)

	if ret == nil || len(ret) != 1 || ret[0].Kind() != reflect.Bool {
		log.Printf("The processing function of message %v is incorrectly written.\n", cmd)
		return cmd, false
	}

	return cmd, ret[0].Interface().(bool)
}

func (client *Client) SendMsgToServer(cmd string) error {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("SendMsg err. client=%v, cmd=%v, err=%v\n", client.uuid, cmd, err)
		}
	}()

	_, bool := client.NeedWaitServerMsg()
	if bool {
		//errMsg := fmt.Sprintf("Client[%v]  wait msg %v .\n", client.uuid, cmdWait)
		//log.Println(errMsg)
		return nil
	}

	value := reflect.ValueOf(client)
	f := value.MethodByName(cmd)
	if !f.IsValid() || f.Kind() != reflect.Func {
		errMsg := fmt.Sprintf("SendMsg error, No handler found for message %v.\n", cmd)
		log.Println(errMsg)
		return errors.New(errMsg)
	}

	if client.conn == nil {
		return errors.New("client conn is nil")
	}

	in := make([]reflect.Value, 0)
	out := f.Call(in)
	if len(out) >= 1 && out[0].Interface() != nil {
		err := out[0].Interface().(error)
		if err != nil {
			client.conn = nil
			log.Printf("%v, set client conn nil. client=%v\n", err, client.uuid)
			return err
		}
	}

	client.msgData.StoreSendMsg(&cmd)
	return nil
}

func (client *Client) GetMsgData() []MsgMiniData {
	return client.msgData.GetData()
}

// 是否正在等待消息返回
func (client *Client) NeedWaitServerMsg() (string, bool) {
	return client.msgData.NeedWaitServerMsg()
}

func (client *Client) ConnValid() bool {
	return client.conn != nil
}

func (client *Client) CloseConn() {
	if client.conn != nil {
		client.conn.Close()
		client.conn = nil
	}
}
