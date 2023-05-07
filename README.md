# IOT Studio 设备模拟器说明文档

## 目录
- [IOT Studio 设备模拟器说明文档](#iot-studio-设备模拟器说明文档)
	- [目录](#目录)
	- [如何运行Coap设备模拟器](#如何运行coap设备模拟器)
			- [示例1: 使用flag运行](#示例1-使用flag运行)
			- [示例2：使用配置文件运行](#示例2使用配置文件运行)
	- [如何运行MQTT设备模拟器](#如何运行mqtt设备模拟器)
		- [以非TLS方式运行](#以非tls方式运行)
			- [示例1：flag运行](#示例1flag运行)
			- [示例2：使用配置文件运行](#示例2使用配置文件运行-1)
		- [以TLS方式运行](#以tls方式运行)
			- [示例1：flag运行](#示例1flag运行-1)
			- [示例2：使用配置文件运行](#示例2使用配置文件运行-2)
	- [通用配置](#通用配置)
		- [控制设备运行时行为(修改后不可实时生效, 需重启程序)](#控制设备运行时行为修改后不可实时生效-需重启程序)
		- [控制设备上行数据(修改后可实时生效，无需重启程序)](#控制设备上行数据修改后可实时生效无需重启程序)
		- [实时控制设备上行开关(修改后立即生效，无需重启程序)](#实时控制设备上行开关修改后立即生效无需重启程序)
		- [手动获取期望值](#手动获取期望值)
		- [手动清除期望值](#手动清除期望值)
		- [设备下线](#设备下线)
		- [强制结束程序](#强制结束程序)


## 如何运行Coap设备模拟器
Windows在命令行运行bin目录下的```studio-coap-windows.exe```

Linux在命令行运行bin目录下的```studio-coap-linux```


首先查看需要的运行参数
```
./bin/studio-coap-linux -help

Usage of ./bin/studio-coap-linux:
  -cf string
        除了使用flag来指定连接信息外，设备也可以使用配置文件的方式来进行连接。(配置文件见本仓库中的coap_conn_1.json, 使用-cf后，配置文件中的配置信息将覆盖其他flag提供的配置信息)
  -dk string
        设备密钥，可在设备详情页找到 
  -dn string
        设备名称，可在设备详情页找到 
  -host string
         Studio Coap接入机地址 
  -l string
        日志等级，可选项：debug, info, error. debug能看到设备发送的json payload信息, 比较占空间，故默认设置为info 
  -pid string
        设备的ProductID，可在设备详情页找到 
  -port int
         Studio Coap接入机端口 
```

然后输入参数运行程序(不输入会使用默认值运行程序，默认值使用的是测试环境的某一个设备)


#### 示例1: 使用flag运行
```
./bin/studio-coap-linux \
-host {host ip} \
-port {port number} \
-pid {product id} \
-dk {device key} \
-dn {device name} 
```

#### 示例2：使用配置文件运行
```
./bin/studio-coap-linux -cf conn/coap_conn.json
```


## 如何运行MQTT设备模拟器
Windows在命令行运行bin目录下的```studio-mqtt-windows.exe```

Linux在命令行运行bin目录下的```studio-mqtt-linux```

首先查看需要的运行参数
```
./bin/studio-mqtt-linux -help

Usage of ./bin/studio-mqtt-linux:
  -cf string
        除了使用flag来指定连接信息外，设备也可以使用配置文件的方式来进行连接。
  -dk string
        设备密钥，可在设备详情页找到 
  -dn string
        设备名称，可在设备详情页找到 
  -host string
         Studio MQTT接入机地址 
  -l string
        日志等级，可选项：debug, info, error. debug能看到设备发送的json payload信息, 比较占空间，故默认设置为info 
  -pem string
        使用TLS进行连接时需要的CA证书, 默认为空，为空时表示不使用TLS进行连接，不为空时表示使用TLS进行连接。注意，如要使用TLS进行连接，URI前缀必须为tls://, 示例：-host tls://{host ip}
  -pid string
        设备的ProductID，可在设备详情页找到 
  -port int
         Studio MQTT接入机端口 
  -qos uint
        MQTT设备上行QOS，支持0, 1 
  -retained
        MQTT设备上行的retained属性, 接入机只支持0
```

然后输入参数运行程序(不输入会使用默认值运行程序，默认值使用的是测试环境的某一个设备)


### 以非TLS方式运行
#### 示例1：flag运行
```
./bin/studio-mqtt-linux \
-host {host ip} \
-port {port number} \
-pid {product id} \
-dk {device key} \
-dn {device name}
```

#### 示例2：使用配置文件运行
```
./bin/studio-mqtt-linux -cf conn/mqtt_conn_normal.json
```


### 以TLS方式运行
要以TLS方式运行模拟器，需要：
1. 加入-pem选项, 证书文件请向研发索要。
2. 在-host的参数中加入tls://前缀


#### 示例1：flag运行
```
./bin/studio-mqtt \
-host '{host ip}' \
-port {port number} \
-pid {product id} \
-dk {device key} \
-dn {device name}
-pem {path to .perm file} 
```

#### 示例2：使用配置文件运行
```
./bin/studio-mqtt-linux -cf conn/mqtt_conn_tls.json -pem serverCert.pem

```


## 通用配置
### 控制设备运行时行为(修改后不可实时生效, 需重启程序)
可通过在运行模拟器前，修改```config.json```文件中的参数来控制设备的行为。在程序运行后修改该配置，必须重新运行程序后方能生效。
```
{
    "lt": 50,                       // Coap设备的LifeTime参数，单位秒，代表设备存活时长, 支持10s - 7天
    "keepAlive: 180,                // MQTT设备的KeepAlive参数, 单位秒，代表设备存活时长，还吃10s - 1800s
    "dataUpwardInterval": 6,        // 设备上行数据间隔(秒), 10次/s，超过后，Non报文作丢弃处理，Con报文回复403
    "eventUpwardInterval": 20,      // 设备上行事件间隔(秒)  10次/s，超过后，Non报文作丢弃处理，Con报文回复403
    "getDesiredInterval": 30,       // 设备从平台获取期望值间隔(秒)
    "keepAliveInterval": 40         // 设备上行保活请求间隔(秒), 5s内不超过10次，超过后踢设备下线，并回复rst，拉黑设备5分钟。
}
```


### 控制设备上行数据(修改后可实时生效，无需重启程序)
可通过修改data目录下的```property.json, event.json, desired.json, desired_clear.json```来控制设备上行的数据

可在设备运行时，通过修改上述三个文件来实时上报不同的数据，无需重启程序。

json文件中的id和time会由程序自动生成, 请勿修改。

``` 除了id处的"%s"字符，请勿在payload中使用任何其他 "%"字符，否则会导致ID格式化的错误，如需测试上行特殊字符操作，请使用其他特殊字符等价类代替，如"$" ```


```property.json```    控制设备上行的属性
```
{
    "id": "%s",                     // 由程序自动生成，勿动
    "version": "1.0",               // 物模型版本号，可以不填
    "params": {                     // 可在params中添加多个属性的定义
        "temp": {                   // 示例中的temp是属性标识符，可根据自身设备的物模型来设定
            "value": 55,            // 属性的值，可随意修改
            "time": %[2]v           // 如果写 `%[2]v`, 时间戳由程序自动生成；如果自己写时间戳，则程序使用自己写的值进行上报；如果不传该字段，则由平台自动生成。
        }
    }
}
```

```event.json```       控制设备上行的事件
```
{
    "id": "%s",                      // 由程序自动生成，勿动 
    "version": "1.0",                // 物模型版本号，可以不填
    "params": {                      // 可在params中添加多个事件的定义
        "someEvent": {               // 示例中的someEvent是事件的标识符，可根据自身设备的物模型来设定
            "value": {
                "someOutput": 50     // 事件的值，可随意修改
            },
            "time": %[2]v            // 如果写 `%[2]v`, 时间戳由程序自动生成；如果自己写时间戳，则程序使用自己写的值进行上报；如果不传该字段，则由平台自动生成。
        }
    }
}
```

```desired.json```     控制设备上行的获取期望值请求
```
{
    "id": "%s",                       // 由程序自动生成，勿动
    "version": "1.0",                 // 物模型版本号，可以不填
    "params": [                       // 可在params中添加个多个期望获取的属性的标识符
        "temp"                        // 属性的标识符
    ]
}
```

```desired_clear.json``` 控制设备上行的清除期望值请求
```
{
    "id": "%v",             // 由程序自动生成，勿动
    "version": "1.0",       // 物模型版本号，可以不填
    "params": {             // params中可输入多个需要清除的期望值
        "someInt64": {
            "version": 1    // 要删除期望属性的版本号, 选填。具体作用见需求文档。
        },
        "someInt32": {}     // 不填version时需保留{} 
    }
}
```


### 实时控制设备上行开关(修改后立即生效，无需重启程序)
在data目录下的```switch.json```文件中可以通过修改参数来实时开启或停止上行数据，修改后立即生效，无需重启程序。
```
{
    "upwardData": true,     // 上行数据开关
    "upwardEvent": true,    // 上行事件开关
    "getDesired": true ,    // 获取设备期望值开关
    "keepAlive": true       // 上行保活开关
}
```


### 手动获取期望值
在命令行输入``` G + 回车 ```后，程序会读取data/desired.json中的数据作为payload并发送获取期望值请求
```
g
DEBU[2020-06-22 12:39:38] 
{
    "id": "37052",
    "version": "1.0",
    "params": [
        "someInt32",
        "someInt64",
        "someString",
        "someFloat32",
        "someFloat64",
        "someStruct",
        "someBool",
        "someEnum",
        "someBitmap"
    ]
} 
INFO[2020-06-22 12:39:38] <<<<< 已发送消息: Type: Confirmable, Code: POST, MsgID: 10936, URIPATH: $sys/H445PYs3Gm/coap_d_1/thing/property/desired/get  上行类型="获取期望值"
INFO[2020-06-22 12:39:38] >>>>> 已收到平台消息: Type: Acknowledgement, Code: Content, MsgID: 10936  下行类型="ACK Content"
DEBU[2020-06-22 12:39:38] 收到期望值Payload: {"msg":{"someBool":{"version":15,"value":true},"someBitmap":{"version":15,"value":3},"someEnum":{"version":15,"value":0}},"id":"37052","code":200} 
DEBU[2020-06-22 12:39:42] 
```


### 手动清除期望值
在命令行输入``` C + 回车 ```后，程序会读取data/desired_clear.json中的数据作为payload并发送获清除取期望值请求
```
c
DEBU[2020-06-18 09:53:43] 

{
    "id": "39671",
    "version": "1.0",
    "params": {
        "someInt64": {
        }
    }
} 
INFO[2020-06-18 09:53:43] <<<<< 已发送消息: Type: Confirmable, Code: POST, MsgID: 14732, URIPATH: $sys/H445PYs3Gm/coap_d_1/thing/property/desired/delete  上行类型="清除期望值"
INFO[2020-06-18 09:53:43] >>>>> 已收到平台消息: Type: Acknowledgement, Code: Content, MsgID: 14732  下行类型="ACK Content"
```


### 设备下线
按``` CTRL + C ```时，程序会发起登出请求，如果请求成功，程序自动结束。(MQTT设备无论请求成功与否均会直接结束程序)
```
^CDEBU[2020-06-18 09:54:29] 登出请求Payload:
INFO[2020-06-18 09:54:29] <<<<< 已发送消息: Type: Confirmable, Code: DELETE, MsgID: 3232, URIPATH: $sys/H445PYs3Gm/coap_d_1/log_out  上行类型="登出"
INFO[2020-06-18 09:54:29] >>>>> 已收到平台消息: Type: Acknowledgement, Code: Deleted, MsgID: 3232  下行类型="ACK DELETED"
INFO[2020-06-18 09:54:29] 设备已正常下线, 程序结束...          
```


### 强制结束程序
如果登出请求失败，程序不会自动结束，此时可使用``` CTRL + \``` 让程序强制结束
```
^\INFO[2020-06-18 09:55:12] 强制退出程序...                                    
exit status 255
```
[返回顶部](#目录)