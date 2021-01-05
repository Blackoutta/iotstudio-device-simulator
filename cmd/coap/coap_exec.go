package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"code.iot.chinamobile.com/huyangyi/onenetstudio-coap-client/simulator"
	"github.com/dustin/go-coap"
	"github.com/sirupsen/logrus"
)

// flags
var (
	serverHost = flag.String("host", "10.12.31.100", "OneNET Studio Coap接入机地址")
	serverPort = flag.Int("port", 5683, "OneNET Studio Coap接入机端口")
	productID  = flag.String("pid", "H445PYs3Gm", "设备的ProductID，可在设备详情页找到")
	deviceKey  = flag.String("dk", "n9VryY2R78mB+euzd5cx9/Xrr7mlF6En0yGX5xWDJz8=", "设备密钥，可在设备详情页找到")
	deviceName = flag.String("dn", "coap_d_1", "设备名称，可在设备详情页找到")
	logLevel   = flag.String("l", "info", "日志等级，可选项：debug, info, error. debug能看到设备发送的json payload信息, 比较占空间，故默认设置为info")
	connFile   = flag.String("cf", "", "除了使用flag来指定连接信息外，设备也可以使用配置文件的方式来进行连接。")
)

// general configs
var (
	ll         int
	workDir    string
	controller simulator.Switch
	config     simulator.Config
	logger     *logrus.Logger
)

func init() {
	flag.Parse()

	// 如果用户使用了-cf flag并指定了配置文件
	if *connFile != "" {
		fmt.Printf("检测到使用配置文件: %v\n", *connFile)
		var conn simulator.Connection
		simulator.ReadConfig(*connFile, &conn)
		*serverHost = conn.Host
		*serverPort = conn.Port
		*productID = conn.ProductId
		*deviceKey = conn.DeviceKey
		*deviceName = conn.DeviceName
		*logLevel = conn.LogLevel
	}

	// 打印配置信息
	fmt.Println("=======================================")
	fmt.Println("host:", *serverHost)
	fmt.Println("port:", *serverPort)
	fmt.Println("productID:", *productID)
	fmt.Println("device key:", *deviceKey)
	fmt.Println("device name:", *deviceName)
	fmt.Println("=======================================")

	switch *logLevel {
	case "debug":
		ll = 5
	case "info":
		ll = 4
	case "error":
		ll = 2
	default:
		ll = 4
	}

	// 初始化workdir
	workDir = simulator.GetWorkDir()

	// 初始化controller
	go func() {
		for {
			simulator.ReadConfig(workDir+"/data/switch.json", &controller)
			time.Sleep(5 * time.Second)
		}
	}()

	// 初始化logger
	logger = &logrus.Logger{
		Out: os.Stdout,
		Formatter: &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		},
		Level: logrus.Level(ll),
	}

	// 初始化设备配置
	simulator.ReadConfig(workDir+"/config.json", &config)
	fmt.Printf("%+v\n", config)
}

func main() {
	// 监听系统Signal
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGQUIT)

	// 初始化设备
	c := simulator.CoapClient{
		ServerHost: *serverHost,
		ServerPort: *serverPort,
		ProductID:  *productID, DeviceKey: *deviceKey,
		DeviceName: *deviceName,
		Logger:     logger,
		AuthMsg: simulator.AuthMessage{
			Lt: int64(config.Lt),
			SasToken: simulator.SasToken{
				Version: "2018-10-31",
				Et:      time.Now().AddDate(0, 3, 0).Unix(),
				Method:  simulator.AUTH_SHA1,
			},
		},
	}

	// ============== 设备活动开始================//

	// 建立UDP连接
	if err := c.Dial(*serverHost, *serverPort); err != nil {
		logger.Printf("Failed to dial coap server: %v", err)
		os.Exit(-1)
	}

	// 发送鉴权请求
	logger.Println("发送设备注册鉴权请求...")
	token, err := c.Register()
	if err != nil {
		logger.Fatalf(err.Error())
	}

	logger.Println("鉴权通过! 设备已上线!")
	logger.Println("接入机返回Token:", token)

	// 监听平台命令及回复
	go func() {
	Loop:
		for {
			// 接收响应
			msg, err := c.Receive()
			if msg == nil {
				continue
			}
			if err != nil {
				logger.Printf("监听命令发生错误: %v", err)
				continue
			}

			// 根据响应的Type来制定不同动作
			if msg.Type == coap.Acknowledgement {
				switch msg.Code {
				case coap.Changed:
					c.PrintResponse("ACK Changed", msg)
					continue Loop
				case coap.Content:
					c.PrintResponse("ACK Content", msg)
					if len(msg.Payload) > 0 {
						c.Debugf("收到期望值Payload: %v\n", string(msg.Payload))
					}
					continue Loop
				case coap.Deleted:
					c.PrintResponse("ACK DELETED", msg)
					c.Println("设备已正常下线, 程序结束...")
					os.Exit(0)
				default:
					c.PrintResponse("失败ACK", msg)
					continue Loop
				}
			}

			path := msg.PathString()
			if len(path) >= 5 {
				// 如果收到的是property set命令
				if strings.Contains(path, "property/set") {
					c.PrintResponse("Set命令", msg)
					c.Debugf("收到set命令: %v\n", string(msg.Payload))
					//回复给平台已收到set命令
					var setCmd simulator.SetCommand
					if err := json.Unmarshal(msg.Payload, &setCmd); err != nil {
						panic(err)
					}
					resp := simulator.SetCommandResp{
						Id:   setCmd.Id,
						Code: 200,
						Msg:  "Success!",
					}
					respByte, err := json.Marshal(resp)
					if err != nil {
						panic(err)
					}

					// 发送回复
					if err := c.RespondToCommand(coap.Acknowledgement, msg.Token, int(msg.MessageID), respByte); err != nil {
						c.Errorf("回复set命令时发生错误:%v", err)
						continue
					}
					continue
				}
			}
		}
	}()

	// 持续发送保活请求
	go func() {
		for {
			if controller.KeepAlive {
				if err := c.SendHeartBeat(simulator.RandomMsgID()); err != nil {
					logger.Printf("%v", err)
				}
			}
			sleep(config.KeepAliveInterval)
		}
	}()

	// 持续上行数据
	go func() {
		for {
			if controller.UpwardData {
				bs, err := ioutil.ReadFile(workDir + "/data/property.json")
				if err != nil {
					panic(err)
				}

				ds := formatPayload(bs)

				logger.Debug("\n\n" + ds)
				if err := c.UpwardData(coap.Confirmable, token, simulator.RandomMsgID(), []byte(ds)); err != nil {
					logger.Println(err)
				}
			}
			sleep(config.DataUpwardInterval)
		}
	}()

	// 持续上行事件
	go func() {
		for {
			if controller.UpwardEvent {
				bs, err := ioutil.ReadFile(workDir + "/data/event.json")
				if err != nil {
					panic(err)
				}
				ds := formatPayload(bs)
				logger.Debug("\n\n" + ds)
				err = c.UpwardEvent(coap.Confirmable, token, simulator.RandomMsgID(), []byte(ds))
				if err != nil {
					logger.Println(err)
				}
			}
			sleep(config.EventUpwardInterval)
		}
	}()

	// 持续获取设备属性期望
	go func() {
		for {
			if controller.GetDesired {
				bs, err := ioutil.ReadFile(workDir + "/data/desired.json")
				if err != nil {
					panic(err)
				}
				ss := fmt.Sprintf(string(bs), simulator.RandomMsgID())
				logger.Debug("\n\n" + ss)
				err = c.GetDesired(coap.Confirmable, token, simulator.RandomMsgID(), []byte(ss))
				if err != nil {
					logger.Println(err)
				}
			}
			sleep(config.GetDesiredInterval)
		}
	}()

	var cmd string
	// 检测用户输入
	go func() {
		for {
			_, err := fmt.Scanln(&cmd)
			if err != nil {
				if err.Error() == "unexpected newline" {
					continue
				}
				fmt.Println(err)
			}
			// 输入c并按回车，可以发送清除期望值请求
			switch cmd {
			case "c":
				bs, err := ioutil.ReadFile(workDir + "/data/desired_clear.json")
				if err != nil {
					panic(err)
				}
				ss := fmt.Sprintf(string(bs), simulator.RandomMsgID())
				logger.Debug("\n\n" + ss)
				if err := c.ClearDesired(coap.Confirmable, token, simulator.RandomMsgID(), []byte(ss)); err != nil {
					panic(err)
				}
			case "g":
				bs, err := ioutil.ReadFile(workDir + "/data/desired.json")
				if err != nil {
					panic(err)
				}
				ss := fmt.Sprintf(string(bs), simulator.RandomMsgID())
				logger.Debug("\n\n" + ss)
				if err := c.GetDesired(coap.Confirmable, token, simulator.RandomMsgID(), []byte(ss)); err != nil {
					logger.Println(err)
				}
			}
		}
	}()

	// 根据terminal收到的不同信号做出不同动作
	for {
		select {
		case s := <-sigChan:
			switch s {
			case syscall.SIGINT: // 如果按CTRL + C, 则发送登出请求
				logout := c.AuthMsg
				logout.Lt = 0
				dr, err := simulator.AuthJson(logout, c.ProductID, c.DeviceName, c.DeviceKey)
				if err != nil {
					panic(err)
				}

				if err := c.Deregister(dr); err != nil {
					panic(err)
				}
			case syscall.SIGQUIT: // 如果按CTRL + \，则强制退出程序（可在登出请求失败，程序无法正常结束的情况下使用)
				c.Println("强制退出程序...")
				os.Exit(-1)
			}
		}
	}

}
func sleep(sec float64) {
	time.Sleep(time.Duration(sec) * time.Second)
}

func formatPayload(payload []byte) string {
	var result string
	if strings.Contains(string(payload), `"time"`) {
		if strings.Contains(string(payload), `%[2]v`) {
			result = fmt.Sprintf(string(payload), simulator.RandomMsgID(), time.Now().Unix()*1000)
		} else {
			result = fmt.Sprintf(string(payload), simulator.RandomMsgID())
		}
	} else {
		result = fmt.Sprintf(string(payload), simulator.RandomMsgID())
	}
	return result
}
