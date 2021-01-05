package simulator

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type Connection struct {
	DeviceKey  string `json:"deviceKey"`
	DeviceName string `json:"deviceName"`
	Host       string `json:"host"`
	LogLevel   string `json:"logLevel"`
	ProductId  string `json:"productId"`
	Port       int    `json:"port"`
}

type Config struct {
	Lt                  float64 `json:"lt"`
	KeepAlive           float64 `json:"keepAlive"`
	DataUpwardInterval  float64 `json:"dataUpwardInterval"`
	EventUpwardInterval float64 `json:"eventUpwardInterval"`
	GetDesiredInterval  float64 `json:"getDesiredInterval"`
	KeepAliveInterval   float64 `json:"keepAliveInterval"`
}

type Switch struct {
	UpwardData  bool `json:"upwardData"`
	UpwardEvent bool `json:"upwardEvent"`
	GetDesired  bool `json:"getDesired"`
	KeepAlive   bool `json:"keepAlive"`
}

type Logger interface {
	Println(args ...interface{})
	Printf(format string, args ...interface{})
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	WithField(key string, value interface{}) *logrus.Entry
}

func ReadConfig(filePath string, dst interface{}) {
	bs, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(bs, dst); err != nil {
		panic(err)
	}
}

func GetWorkDir() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	wd := filepath.Dir(filepath.Dir(ex))
	return wd
}
