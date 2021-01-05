package simulator

import (
	"errors"
	"fmt"

	"github.com/dustin/go-coap"
)

type CoapClient struct {
	ServerHost string
	ServerPort int
	ProductID  string
	DeviceKey  string
	DeviceName string
	Token      []byte
	AuthMsg    AuthMessage
	Logger
	*coap.Conn
}

func (c *CoapClient) Dial(serverHost string, serverPort int) error {
	serverAddr := fmt.Sprintf("%s:%d", serverHost, serverPort)
	conn, err := coap.Dial("udp", serverAddr)
	if err != nil {
		return err
	}
	c.Conn = conn
	c.Printf("Successfully dialed in coap server: %v", serverAddr)
	return nil
}

const (
	AUTH_MD5    = "md5"
	AUTH_SHA1   = "sha1"
	AUTH_SHA256 = "sha256"
)

func (c *CoapClient) Register() ([8]byte, error) {
	// 生成鉴权请求payload(json格式)
	authPl, err := AuthJson(c.AuthMsg, c.ProductID, c.DeviceName, c.DeviceKey)
	if err != nil {
		return [8]byte{}, err
	}
	// 发送鉴权请求
	authReq := coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.POST,
		MessageID: uint16(RandomMsgID()),
		Payload:   authPl,
	}
	c.addBaseOptions(&authReq)
	authReq.AddOption(coap.URIPath, "log_in")
	resp, err := c.Send(authReq)
	if err != nil {
		return [8]byte{}, err
	}
	c.Debugf("鉴权请求Payload:\n"+"%+v", string(authReq.Payload))
	c.PrintRequest("注册鉴权", authReq)
	if resp == nil {
		panic(errors.New("未收到鉴权失败后的响应，程序结束..."))
	}
	c.PrintResponse("鉴权ACK", resp)
	if resp.Code != coap.Created {
		return [8]byte{}, fmt.Errorf("鉴权失败, 接入机返回Code: %v", resp.Code)
	}

	var token [8]byte

	for i := range resp.Payload {
		token[i] = resp.Payload[i]
	}
	return token, nil
}

func (c *CoapClient) addBaseOptions(msg *coap.Message) {
	msg.AddOption(coap.URIHost, c.ServerHost)
	msg.AddOption(coap.URIPort, c.ServerPort)
	msg.AddOption(coap.URIPath, "$sys")
	msg.AddOption(coap.URIPath, c.ProductID)
	msg.AddOption(coap.URIPath, c.DeviceName)
	msg.AddOption(coap.ContentFormat, coap.AppJSON)
}

func (c *CoapClient) Deregister(payload []byte) error {
	req := coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.DELETE,
		MessageID: uint16(RandomMsgID()),
		Payload:   payload,
	}
	c.addBaseOptions(&req)
	req.AddOption(coap.URIPath, "log_out")
	err := c.Transmit(req)
	if err != nil {
		return err
	}
	c.Debugf("登出请求Payload:\n"+"%+v", string(req.Payload))
	c.PrintRequest("登出", req)

	return nil
}

func (c *CoapClient) SendHeartBeat(msgID int) error {
	// 生成保活请求payload(json格式)
	hb, err := AuthJson(c.AuthMsg, c.ProductID, c.DeviceName, c.DeviceKey)
	if err != nil {
		return err
	}
	// 发送保活请求
	authReq := coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.POST,
		MessageID: uint16(msgID),
		Payload:   hb,
	}
	c.addBaseOptions(&authReq)
	authReq.AddOption(coap.URIPath, "keep_alive")

	err = c.Transmit(authReq)
	if err != nil {
		return err
	}
	c.PrintRequest("保活", authReq)
	c.Debugf("保活请求Payload:\n%v", string(authReq.Payload))
	return nil
}

func (c *CoapClient) UpwardData(t coap.COAPType, token [8]byte, msgID int, jsonData []byte) error {
	req := coap.Message{
		Type:      t,
		Code:      coap.POST,
		MessageID: uint16(msgID),
		Token:     token[:],
		Payload:   jsonData,
	}
	c.addBaseOptions(&req)
	req.AddOption(coap.URIPath, "thing")
	req.AddOption(coap.URIPath, "property")
	req.AddOption(coap.URIPath, "post")
	req.AddOption(coap.Accept, coap.AppJSON)
	err := c.Transmit(req)

	if err != nil {
		return err
	}
	c.PrintRequest("数据点", req)
	return nil
}

func (c *CoapClient) UpwardEvent(t coap.COAPType, token [8]byte, msgID int, jsonData []byte) error {
	req := coap.Message{
		Type:      t,
		Code:      coap.POST,
		MessageID: uint16(msgID),
		Token:     token[:],
		Payload:   jsonData,
	}
	c.addBaseOptions(&req)
	req.AddOption(coap.URIPath, "thing")
	req.AddOption(coap.URIPath, "event")
	req.AddOption(coap.URIPath, "post")
	req.AddOption(coap.Accept, coap.AppJSON)
	err := c.Transmit(req)

	if err != nil {
		return err
	}
	c.PrintRequest("事件", req)
	return nil
}

func (c *CoapClient) GetDesired(t coap.COAPType, token [8]byte, msgID int, jsonData []byte) error {
	req := coap.Message{
		Type:      t,
		Code:      coap.POST,
		MessageID: uint16(msgID),
		Token:     token[:],
		Payload:   jsonData,
	}
	c.addBaseOptions(&req)
	req.AddOption(coap.URIPath, "thing")
	req.AddOption(coap.URIPath, "property")
	req.AddOption(coap.URIPath, "desired")
	req.AddOption(coap.URIPath, "get")
	req.AddOption(coap.Accept, coap.AppJSON)
	err := c.Transmit(req)

	if err != nil {
		return err
	}
	c.PrintRequest("获取期望值", req)
	return nil
}

func (c *CoapClient) ClearDesired(t coap.COAPType, token [8]byte, msgID int, jsonData []byte) error {
	req := coap.Message{
		Type:      t,
		Code:      coap.POST,
		MessageID: uint16(msgID),
		Token:     token[:],
		Payload:   jsonData,
	}
	c.addBaseOptions(&req)
	req.AddOption(coap.URIPath, "thing")
	req.AddOption(coap.URIPath, "property")
	req.AddOption(coap.URIPath, "desired")
	req.AddOption(coap.URIPath, "delete")
	req.AddOption(coap.Accept, coap.AppJSON)
	err := c.Transmit(req)

	if err != nil {
		return err
	}
	c.PrintRequest("清除期望值", req)
	return nil
}
func (c *CoapClient) RespondToCommand(t coap.COAPType, token []byte, msgID int, jsonData []byte) error {
	req := coap.Message{
		Type:      t,
		Code:      coap.Content,
		MessageID: uint16(msgID),
		Token:     token[:],
		Payload:   jsonData,
	}
	c.addBaseOptions(&req)
	req.AddOption(coap.URIPath, "thing")
	req.AddOption(coap.URIPath, "property")
	req.AddOption(coap.URIPath, "set_reply")
	req.AddOption(coap.Accept, coap.AppJSON)
	err := c.Transmit(req)

	if err != nil {
		return err
	}
	c.PrintRequest("回复set命令", req)
	c.Debugf("回复set命令Payload: %v", string(req.Payload))
	return nil
}

func (c *CoapClient) PrintRequest(reqType string, m coap.Message) {
	c.WithField("上行类型", reqType).Printf("<<<<< 已发送消息: Type: %v, Code: %v, MsgID: %v, URIPATH: %v\n", m.Type, m.Code, m.MessageID, m.PathString())
}

func (c *CoapClient) PrintResponse(msgType string, m *coap.Message) {
	c.WithField("下行类型", msgType).Printf(">>>>> 已收到平台消息: Type: %v, Code: %v, MsgID: %v\n", m.Type, m.Code, m.MessageID)
}
