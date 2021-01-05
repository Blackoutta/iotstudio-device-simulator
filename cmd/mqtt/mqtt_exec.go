package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"code.iot.chinamobile.com/huyangyi/onenetstudio-coap-client/simulator"
	"github.com/sirupsen/logrus"
)

var (
	serverHost = flag.String("host", "10.12.31.5", "OneNET Studio MQTT接入机地址")
	serverPort = flag.Int("port", 1883, "OneNET Studio MQTT接入机端口")
	productID  = flag.String("pid", "BV4BxD5f9K", "设备的ProductID，可在设备详情页找到")
	deviceKey  = flag.String("dk", "T+HfLbOB209zsf/Hn20xPbXRExGnGyWLNZl5bY9/MVs=", "设备密钥，可在设备详情页找到")
	deviceName = flag.String("dn", "mqtt_d_1", "设备名称，可在设备详情页找到")
	logLevel   = flag.String("l", "info", "日志等级，可选项：debug, info, error. debug能看到设备发送的json payload信息, 比较占空间，故默认设置为info")
	qos        = flag.Uint("qos", 1, "MQTT设备上行QOS，支持0, 1")
	retained   = flag.Bool("retained", false, "MQTT设备上行的retained属性, 接入机只支持0(false)")
	pem        = flag.String("pem", "", "使用TLS进行连接时需要的CA证书, 默认为空，为空时表示不使用TLS进行连接，不为空时表示使用TLS进行连接。注意，如要使用TLS进行连接，URI前缀必须为tls://, 示例：-host tls://10.12.31.5")
	connFile   = flag.String("cf", "", "除了使用flag来指定连接信息外，设备也可以使用配置文件的方式来进行连接。")
)

var (
	ll         int
	workDir    string
	controller simulator.Switch
	config     simulator.Config
	logger     *logrus.Logger
)

var (
	uplog    *logrus.Entry
	eventlog *logrus.Entry
	downlog  *logrus.Entry
	sublog   *logrus.Entry
	resplog  *logrus.Entry
	gdlog    *logrus.Entry
	cdlog    *logrus.Entry
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
		ReportCaller: false,
		Level:        logrus.Level(ll),
	}

	uplog = logger.WithField("动作", "上行属性")
	eventlog = logger.WithField("动作", "上行事件")
	downlog = logger.WithField("动作", "收到下行")
	sublog = logger.WithField("动作", "订阅")
	resplog = logger.WithField("动作", "回复set命令")
	gdlog = logger.WithField("动作", "get desired")
	cdlog = logger.WithField("动作", "clear desired")

	// 初始化设备配置
	simulator.ReadConfig(workDir+"/config.json", &config)
	fmt.Printf("%+v\n", config)

}

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGQUIT)

	// 初始化客户端配置
	c := &simulator.StudioMqttClient{
		ServerHost: *serverHost,
		ServerPort: *serverPort,
		ProductID:  *productID,
		DeviceKey:  *deviceKey,
		DeviceName: *deviceName,
		PemFile:    *pem,
		AuthMsg: simulator.AuthMessage{
			SasToken: simulator.SasToken{
				Version: "2018-10-31",
				Et:      time.Now().AddDate(0, 3, 0).Unix(),
				Method:  simulator.AUTH_SHA1,
			},
		},
		Logger: logger,
	}

	// 初始化客户端
	if err := c.NewMqttClient(config.KeepAlive); err != nil {
		panic(err)
	}
	c.Println("设备连接中...")

	// ============== 设备活动开始================//
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	c.Println("鉴权成功，设备成功上线!")

	// 持续上行属性
	go func() {
		pubTopic := fmt.Sprintf("$sys/%v/%v/thing/property/post", c.ProductID, c.DeviceName)
		subTopic := fmt.Sprintf("$sys/%v/%v/thing/property/post/reply", c.ProductID, c.DeviceName)

		// 订阅reply topic
		if token := c.Subscribe(subTopic, 2, upwardRespHandler); token.Wait() && token.Error() != nil {
			sublog.Errorf("订阅时发生错误: %v", token.Error())
			os.Exit(1)
		} else {
			checkSubResult(token, subTopic)
		}
		sublog.Printf("订阅主题成功: %v", subTopic)

		// 上行开始
		for {
			if controller.UpwardData {
				bs, err := ioutil.ReadFile(workDir + "/data/property.json")
				if err != nil {
					panic(err)
				}
				ds := formatPayload(bs)
				logger.Debug("\n\n" + ds)
				if token := c.Publish(pubTopic, byte(*qos), *retained, []byte(ds)); token.Wait() && token.Error() != nil {
					uplog.Errorf("上行属性时发生错误: %v", token.Error())
					sleep(config.DataUpwardInterval)
					continue
				}
				uplog.Printf("已上行属性, Topic: %v", pubTopic)
			}
			sleep(config.DataUpwardInterval)
		}
	}()
	// 持续上行事件
	go func() {
		pubTopic := fmt.Sprintf("$sys/%v/%v/thing/event/post", c.ProductID, c.DeviceName)
		subTopic := fmt.Sprintf("$sys/%v/%v/thing/event/post/reply", c.ProductID, c.DeviceName)

		// 订阅reply topic
		if token := c.Subscribe(subTopic, byte(*qos), upwardEventHandler); token.Wait() && token.Error() != nil {
			sublog.Errorf("订阅时发生错误: %v", token.Error())
			os.Exit(1)
		} else {
			checkSubResult(token, subTopic)
		}
		sublog.Printf("订阅主题成功: %v", subTopic)

		// 上行开始
		for {
			if controller.UpwardEvent {
				bs, err := ioutil.ReadFile(workDir + "/data/event.json")
				if err != nil {
					panic(err)
				}
				ds := formatPayload(bs)
				logger.Debug("\n\n" + ds)

				if token := c.Publish(pubTopic, byte(*qos), *retained, []byte(ds)); token.Wait() && token.Error() != nil {
					eventlog.Errorf("上行事件时发生错误: %v", token.Error())
					sleep(config.DataUpwardInterval)
					continue
				}
				eventlog.Printf("已上行事件, Topic: %v", pubTopic)
			}
			sleep(config.EventUpwardInterval)
		}
	}()

	// 监听设备属性设置
	go func() {
		subTopic := fmt.Sprintf("$sys/%v/%v/thing/property/set", c.ProductID, c.DeviceName)
		pubTopic := fmt.Sprintf("$sys/%v/%v/thing/property/set_reply", c.ProductID, c.DeviceName)
		// 订阅set topic
		if token := c.Subscribe(subTopic, byte(*qos), newPropertySetHandler(pubTopic)); token.Wait() && token.Error() != nil {
			// 订阅异常
			sublog.Errorf("订阅时发生错误: %v", token.Error())
			os.Exit(1)
		} else {
			checkSubResult(token, subTopic)
		}
		sublog.Printf("订阅主题成功: %v", subTopic)
	}()

	// 监听键盘输入，根据输入执行获取设备期望值或清除设备期望值
	go func() {
		subDelete := fmt.Sprintf("$sys/%v/%v/thing/property/desired/delete/reply", c.ProductID, c.DeviceName)
		subGet := fmt.Sprintf("$sys/%v/%v/thing/property/desired/get/reply", c.ProductID, c.DeviceName)

		// 订阅 clear desired response topic
		if token := c.Subscribe(subDelete, byte(*qos), clearDesiredHandler); token.Wait() && token.Error() != nil {
			// 订阅异常
			sublog.Errorf("订阅时发生错误: %v", token.Error())
			os.Exit(1)
		} else {
			checkSubResult(token, subDelete)
		}
		sublog.Printf("订阅主题成功: %v", subDelete)

		// 订阅 get desired response topic
		if token := c.Subscribe(subGet, byte(*qos), getDesiredHandler); token.Wait() && token.Error() != nil {
			// 订阅异常
			sublog.Errorf("订阅时发生错误: %v", token.Error())
			os.Exit(1)
		} else {
			checkSubResult(token, subGet)
		}
		sublog.Printf("订阅主题成功: %v", subGet)

		var cmd string
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

				// 发送清除期望值请求
				pubTopic := fmt.Sprintf("$sys/%v/%v/thing/property/desired/delete", c.ProductID, c.DeviceName)
				if token := c.Publish(pubTopic, byte(*qos), *retained, []byte(ss)); token.Wait() && token.Error() != nil {
					gdlog.Errorf("发送清除期望值请求时发生错误: %v", err)
					continue
				}
				gdlog.Printf("已发送清除期望值请求")
			case "g":
				bs, err := ioutil.ReadFile(workDir + "/data/desired.json")
				if err != nil {
					panic(err)
				}
				ss := fmt.Sprintf(string(bs), simulator.RandomMsgID())
				logger.Debug("\n\n" + ss)

				// 发送获取期望值请求
				pubTopic := fmt.Sprintf("$sys/%v/%v/thing/property/desired/get", c.ProductID, c.DeviceName)
				if token := c.Publish(pubTopic, byte(*qos), *retained, []byte(ss)); token.Wait() && token.Error() != nil {
					cdlog.Errorf("发送获取期望值请求时发生错误: %v", err)
					continue
				}
				cdlog.Printf("已发送获取期望值请求")
			}
		}
	}()

	// 监听登出和退出命令
	for s := range sigChan {
		switch s {
		case syscall.SIGINT: // 如果按CTRL + C, 则发送登出请求
			c.Println("发送登出请求...")
			c.Disconnect(30)
			c.Println("程序结束...")
			os.Exit(0)
		// 发送登出请求
		case syscall.SIGQUIT: // 如果按CTRL + \，则强制退出程序（可在登出请求失败，程序无法正常结束的情况下使用)
			c.Println("强制退出程序...")
			os.Exit(-1)
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
