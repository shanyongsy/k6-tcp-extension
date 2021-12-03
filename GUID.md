# Message access guidelines
## xk6-tcp modification
1. Add a message sending function with the ***same name*** as the sending message in tcpSendMsg.go
```go
func (client *Client) G2CHeartbeatMessage(data *[]byte) bool {
	var msg = pb.G2CHeartbeatMessage{}

	err := proto.Unmarshal(*data, &msg)
	if err != nil {
		log.Printf("%v Err: %v\n", lib.GetFuncName(), err)
		return false
	}

	return msg.Milli != 0
}
```
2. Add a message processing function with the ***same name*** as the received message in tcpRecvMsg.go
```go
func (client *Client) G2CHeartbeatMessage(data *[]byte) bool {
	var msg = pb.G2CHeartbeatMessage{}

	err := proto.Unmarshal(*data, &msg)
	if err != nil {
		log.Printf("%v Err: %v\n", lib.GetFuncName(), err)
		return false
	}

	return msg.Milli != 0
}
```
3. Add the corresponding relationship between ***receiving messages adn sending messages*** in tcpRecvMsg.go
```go
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
```
## JS Script modification
1. Add the message information to be ***sent*** and the expected weight in js/gateway/unit/msg_unit.js
```js
const msgArray = [
    { weight: 10, name: 'C2GTestMessage' },
    { weight: 10, name: 'C2GHeartbeatMessage' },
];
```
## Run resault
[report link](https://d7n9vj8ces.feishu.cn/docs/doccnPgdV0NLXyBZTl6WFEY0y1g)
