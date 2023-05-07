package main

import (
	"encoding/json"
	"os"
	"strings"

	"code.iot.chinamobile.com/huyangyi/studio-coap-client/simulator"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// 检查订阅结果
func checkSubResult(token MQTT.Token, subTopic string) {
	t := token.(*MQTT.SubscribeToken)
	result := t.Result()[subTopic]
	if result == 128 {
		sublog.Errorf("订阅失败, 接入机返回granted qos 128")
		os.Exit(1)
	}
}

// 上行回复handler
func upwardRespHandler(c MQTT.Client, msg MQTT.Message) {
	// Handler
	downlog.Printf("收到回复:\nTopic: %v\nPayload: %v", msg.Topic(), string(msg.Payload()))
	if !strings.Contains(string(msg.Payload()), "success") {
		downlog.Error("上行消息失败!")
	}
}

// 上行事件handler
func upwardEventHandler(c MQTT.Client, msg MQTT.Message) {
	downlog.Printf("收到回复:\nTopic: %v\nPayload: %v", msg.Topic(), string(msg.Payload()))
	if !strings.Contains(string(msg.Payload()), "success") {
		downlog.Error("上行事件失败!")
	}
}

// 命令下发handler
func newPropertySetHandler(pubTopic string) MQTT.MessageHandler {
	return func(c MQTT.Client, msg MQTT.Message) {
		downlog.Printf("收到set命令:\nTopic: %v\nPayload: %v", msg.Topic(), string(msg.Payload()))
		var cmd simulator.SetCommand
		if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
			downlog.Errorf("解析set命令json时发生错误: %v", err)
			return
		}
		// 回复命令
		resp := simulator.SetCommandResp{
			Id:   cmd.Id,
			Code: 200,
			Msg:  "Success! Kupo!",
		}
		respBytes, err := json.Marshal(&resp)
		if err != nil {
			downlog.Errorf("编码响应json时发生错误: %v", err)
			return
		}
		if token := c.Publish(pubTopic, byte(*qos), *retained, respBytes); token.Wait() && token.Error() != nil {
			resplog.Errorf("回复set命令时发生错误: %v", err)
			return
		}
		resplog.Printf("已回复set命令:\nTopic: %v\nPayload:%v", pubTopic, string(respBytes))
	}
}

// clear desired handler
func clearDesiredHandler(c MQTT.Client, msg MQTT.Message) {
	downlog.Printf("收到clear desired回复:\nTopic: %v\nPayload: %v", msg.Topic(), string(msg.Payload()))
	var cmd simulator.SetCommand
	if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
		downlog.Errorf("解析clear desired回复json时发生错误: %v", err)
		return
	}
}

// get desired handler
func getDesiredHandler(c MQTT.Client, msg MQTT.Message) {
	// Handler
	downlog.Printf("收到get desired回复:\nTopic: %v\nPayload: %v", msg.Topic(), string(msg.Payload()))
	var cmd simulator.SetCommand
	if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
		downlog.Errorf("解析get desired回复json时发生错误: %v", err)
		return
	}
}
