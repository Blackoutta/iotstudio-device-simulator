package simulator

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type StudioMqttClient struct {
	ServerHost string
	ServerPort int
	ProductID  string
	DeviceKey  string
	DeviceName string
	PemFile    string
	AuthMsg    AuthMessage
	Logger
	MQTT.Client
}

// TLS Option
func AddTLS(opts *MQTT.ClientOptions, pemFile string) {
	opts.SetTLSConfig(NewTLSConfig(pemFile))
}

func (sm *StudioMqttClient) NewMqttClient(keepAlive float64) error {
	opts := MQTT.NewClientOptions()
	if sm.PemFile != "" {
		sm.Println("检测到使用TLS...")
		AddTLS(opts, sm.PemFile)
	}
	opts.AddBroker(fmt.Sprintf("%v:%v", sm.ServerHost, sm.ServerPort))
	opts.SetClientID(sm.DeviceName)
	opts.SetUsername(sm.ProductID)
	password, err := GenerateSasToken(sm.AuthMsg, sm.ProductID, sm.DeviceName, sm.DeviceKey)
	if err != nil {
		return err
	}
	opts.SetPassword(string(password))
	opts.SetKeepAlive(time.Duration(keepAlive) * time.Second)
	opts.SetProtocolVersion(4)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(false)
	sm.Client = MQTT.NewClient(opts)
	fmt.Println("username:", sm.ProductID)
	fmt.Println("password:", string(password))
	return nil
}

func NewTLSConfig(pemFile string) *tls.Config {
	certPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile(pemFile)
	if err != nil {
		panic(err)
	}
	certPool.AppendCertsFromPEM(pem)
	return &tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: true,
	}
}
