package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"ehang.io/nps/client"
	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/config"
	"ehang.io/nps/lib/crypt"
	"ehang.io/nps/lib/file"
	"ehang.io/nps/lib/install"
	"ehang.io/nps/lib/version"
	"github.com/astaxie/beego/logs"
	"github.com/ccding/go-stun/stun"
	"github.com/kardianos/service"
)

var (
	serverAddr     = flag.String("server", "", "Server addr (ip:port)")
	configPath     = flag.String("config", "", "Configuration file path")
	verifyKey      = flag.String("vkey", "", "Authentication key")
	logType        = flag.String("log", "stdout", "Log output mode（stdout|file）")
	connType       = flag.String("type", "tcp", "Connection type with the server（kcp|tcp）")
	proxyUrl       = flag.String("proxy", "", "proxy socks5 url(eg:socks5://111:222@127.0.0.1:9007)")
	logLevel       = flag.String("log_level", "7", "log level 0~7")
	registerTime   = flag.Int("time", 2, "register time long /h")
	localPort      = flag.Int("local_port", 2000, "p2p local port")
	password       = flag.String("password", "", "p2p password flag")
	target         = flag.String("target", "", "p2p target")
	localType      = flag.String("local_type", "p2p", "p2p target")
	logPath        = flag.String("log_path", "", "npc log path")
	debug          = flag.Bool("debug", true, "npc debug")
	pprofAddr      = flag.String("pprof", "", "PProf debug addr (ip:port)")
	stunAddr       = flag.String("stun_addr", "stun.stunprotocol.org:3478", "stun server address (eg:stun.stunprotocol.org:3478)")
	ver            = flag.Bool("version", false, "show current version")
	disconnectTime = flag.Int("disconnect_timeout", 60, "not receiving check packet times, until timeout will disconnect the client")
	allowedTargets = flag.String("allowed_targets", "", "local allowed targets, split by ','")
	autoAddClient  = flag.Bool("auto_add_client", true, "auto add vkey to server, if vkey is empty, will generate a random vkey")
	delClient      = flag.Bool("del_client", true, "delete client after exit")
	apiAddr        = flag.String("api", "", "web ui port, should be provided with -auto_add_client")
	tcpTunnel      = flag.String("tcp_tunnel", "", "format: server_port->target1:port1|server_port->target2:port2, eg: 8000->127.0.0.1:80")
	udpTunnel      = flag.String("udp_tunnel", "", "format: server_port->target1:port1|server_port->target2:port2, eg: 8000->127.0.0.1:80")
)

func main() {
	flag.Parse()
	logs.Reset()
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	if *ver {
		common.PrintVersion()
		return
	}
	if *logPath == "" {
		*logPath = common.GetNpcLogPath()
	}
	if common.IsWindows() {
		*logPath = strings.Replace(*logPath, "\\", "\\\\", -1)
	}
	if *debug {
		logs.SetLogger(logs.AdapterConsole, `{"level":`+*logLevel+`,"color":true}`)
	} else {
		logs.SetLogger(logs.AdapterFile, `{"level":`+*logLevel+`,"filename":"`+*logPath+`","daily":false,"maxlines":100000,"color":true}`)
	}

	// init service
	options := make(service.KeyValue)
	svcConfig := &service.Config{
		Name:        "Npc",
		DisplayName: "nps内网穿透客户端",
		Description: "一款轻量级、功能强大的内网穿透代理服务器。支持tcp、udp流量转发，支持内网http代理、内网socks5代理，同时支持snappy压缩、站点保护、加密传输、多路复用、header修改等。支持web图形化管理，集成多用户模式。",
		Option:      options,
	}
	if !common.IsWindows() {
		svcConfig.Dependencies = []string{
			"Requires=network.target",
			"After=network-online.target syslog.target"}
		svcConfig.Option["SystemdScript"] = install.SystemdScript
		svcConfig.Option["SysvScript"] = install.SysvScript
	}
	for _, v := range os.Args[1:] {
		switch v {
		case "install", "start", "stop", "uninstall", "restart":
			continue
		}
		if !strings.Contains(v, "-service=") && !strings.Contains(v, "-debug=") {
			svcConfig.Arguments = append(svcConfig.Arguments, v)
		}
	}
	svcConfig.Arguments = append(svcConfig.Arguments, "-debug=false")
	prg := &npc{
		exit: make(chan struct{}),
	}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		logs.Error(err, "service function disabled")
		run()
		// run without service
		wg := sync.WaitGroup{}
		wg.Add(1)
		wg.Wait()
		return
	}
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "status":
			if len(os.Args) > 2 {
				path := strings.Replace(os.Args[2], "-config=", "", -1)
				client.GetTaskStatus(path)
			}
		case "register":
			flag.CommandLine.Parse(os.Args[2:])
			client.RegisterLocalIp(*serverAddr, *verifyKey, *connType, *proxyUrl, *registerTime)
		case "update":
			install.UpdateNpc()
			return
		case "nat":
			c := stun.NewClient()
			c.SetServerAddr(*stunAddr)
			nat, host, err := c.Discover()
			if err != nil || host == nil {
				logs.Error("get nat type error", err)
				return
			}
			fmt.Printf("nat type: %s \npublic address: %s\n", nat.String(), host.String())
			os.Exit(0)
		case "start", "stop", "restart":
			// support busyBox and sysV, for openWrt
			if service.Platform() == "unix-systemv" {
				logs.Info("unix-systemv service")
				cmd := exec.Command("/etc/init.d/"+svcConfig.Name, os.Args[1])
				err := cmd.Run()
				if err != nil {
					logs.Error(err)
				}
				return
			}
			err := service.Control(s, os.Args[1])
			if err != nil {
				logs.Error("Valid actions: %q\n%s", service.ControlAction, err.Error())
			}
			return
		case "install":
			service.Control(s, "stop")
			service.Control(s, "uninstall")
			install.InstallNpc()
			err := service.Control(s, os.Args[1])
			if err != nil {
				logs.Error("Valid actions: %q\n%s", service.ControlAction, err.Error())
			}
			if service.Platform() == "unix-systemv" {
				logs.Info("unix-systemv service")
				confPath := "/etc/init.d/" + svcConfig.Name
				os.Symlink(confPath, "/etc/rc.d/S90"+svcConfig.Name)
				os.Symlink(confPath, "/etc/rc.d/K02"+svcConfig.Name)
			}
			return
		case "uninstall":
			err := service.Control(s, os.Args[1])
			if err != nil {
				logs.Error("Valid actions: %q\n%s", service.ControlAction, err.Error())
			}
			if service.Platform() == "unix-systemv" {
				logs.Info("unix-systemv service")
				os.Remove("/etc/rc.d/S90" + svcConfig.Name)
				os.Remove("/etc/rc.d/K02" + svcConfig.Name)
			}
			return
		}
	}
	s.Run()
}

type npc struct {
	exit chan struct{}
}

func (p *npc) Start(s service.Service) error {
	go p.run()
	return nil
}
func (p *npc) Stop(s service.Service) error {
	close(p.exit)
	if *apiAddr != "" && *delClient {
		id, err := getClientIdByVkey(*apiAddr, *verifyKey)
		if err != nil {
			logs.Error("get client id error", err)
		}
		if err := deleteClient(*apiAddr, id); err != nil {
			logs.Error("delete client error", err)
		}
		logs.Info("delete client <%s> success", *verifyKey)
	}
	if service.Interactive() {
		os.Exit(0)
	}
	return nil
}

func (p *npc) run() error {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			logs.Warning("npc: panic serving %v: %v\n%s", err, string(buf))
		}
	}()
	run()
	select {
	case <-p.exit:
		logs.Warning("stop...")
	}
	return nil
}

func run() {
	common.InitPProfFromArg(*pprofAddr)
	//p2p or secret command
	if *password != "" {
		commonConfig := new(config.CommonConfig)
		commonConfig.Server = *serverAddr
		commonConfig.VKey = *verifyKey
		commonConfig.Tp = *connType
		localServer := new(config.LocalServer)
		localServer.Type = *localType
		localServer.Password = *password
		localServer.Target = *target
		localServer.Port = *localPort
		commonConfig.Client = new(file.Client)
		commonConfig.Client.Cnf = new(file.Config)
		go client.StartLocalServer(localServer, commonConfig)
		return
	}
	env := common.GetEnvMap()
	if *serverAddr == "" {
		*serverAddr, _ = env["NPC_SERVER_ADDR"]
	}
	if *verifyKey == "" {
		*verifyKey, _ = env["NPC_SERVER_VKEY"]
	}
	logs.Info("the version of client is %s, the core version of client is %s", version.VERSION, version.GetVersion())
	if *verifyKey != "" && *serverAddr != "" && (*configPath == "" && *tcpTunnel == "" && *udpTunnel == "") {
		localAllowedTargets := make(map[string]struct{})
		if *allowedTargets != "" {
			for _, v := range strings.Split(*allowedTargets, ",") {
				localAllowedTargets[strings.TrimSpace(v)] = struct{}{}
			}
		}

		go func() {
			for {
				client.NewRPClient(*serverAddr, *verifyKey, *connType, *proxyUrl, nil, *disconnectTime, localAllowedTargets).Start()
				logs.Info("Client closed! It will be reconnected in five seconds")
				time.Sleep(time.Second * 5)
			}
		}()
	} else if *apiAddr != "" && (*tcpTunnel != "" || *udpTunnel != "") {
		logs.Info("start with cmd mode")
		// get bridge port, convert to serverAddr
		if bridgePort, err := getBridgePort(*apiAddr); err != nil {
			logs.Error("get bridge port error: %s", err.Error())
			return
		} else {
			var serverIP = common.GetIpByAddr(*apiAddr)
			*serverAddr = fmt.Sprintf("%s:%d", serverIP, bridgePort)
			logs.Info("get bridge port success: %d", bridgePort)
		}
		if *verifyKey == "" && !*autoAddClient {
			logs.Error("please input verifyKey")
			return
		}
		if *autoAddClient {
			if *verifyKey == "" {
				if hostname, err := os.Hostname(); err != nil {
					logs.Error("get hostname error: %s", err.Error())
					panic(err)
				} else {
					*verifyKey = crypt.Md5(hostname + *tcpTunnel + *udpTunnel)
					logs.Info("auto generate verifyKey: %s", *verifyKey)
				}
			}
			if err := addClient(*apiAddr, *verifyKey); err != nil {
				if !strings.Contains(err.Error(), "Vkey duplicate, please reset") {
					panic(err)
				}
				logs.Warn("add client failed: %s", err.Error())
			} else {
				logs.Info("add client success", *verifyKey)
			}
		}
		var tempConfigPath string = common.GetTmpPath() + "/npc_temp_config.conf"
		var tcpTunnelMap map[string]string = make(map[string]string)
		var udpTunnelMap map[string]string = make(map[string]string)
		var split = func(r rune) bool {
			return r == ',' || r == '，'
		}
		for _, v := range strings.FieldsFunc(*tcpTunnel, split) {
			tcpTunnelMap[strings.TrimSpace(strings.Split(v, "->")[0])] = strings.TrimSpace(strings.Split(v, "->")[1])
		}
		for _, v := range strings.FieldsFunc(*udpTunnel, split) {
			udpTunnelMap[strings.TrimSpace(strings.Split(v, "->")[0])] = strings.TrimSpace(strings.Split(v, "->")[1])
		}
		if err := config.GenerateConfig(*serverAddr, *verifyKey, *connType, tcpTunnelMap, udpTunnelMap, tempConfigPath); err != nil {
			logs.Error("generate config error", err)
			return
		}
		go client.StartFromFile(tempConfigPath)
	} else {
		if *tcpTunnel != "" || *udpTunnel != "" {
			logs.Error("you are using config file mode, -tcpTunnel and -udpTunnel are not allowed")
			return
		}
		if *configPath == "" {
			*configPath = common.GetConfigPath()
		}
		go client.StartFromFile(*configPath)
	}
}

func addClient(webUIAddr string, vkey string) error {
	url := fmt.Sprintf("http://%s/client/add", webUIAddr)
	payload := fmt.Sprintf("remark=undefined&u=&p=&vkey=%s&config_conn_allow=1&compress=1&crypt=0", vkey)
	body, err := postUrlEncoded(url, payload)

	if err != nil {
		return err
	}
	jsonBody := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &jsonBody); err != nil {
		return err
	}
	if jsonBody["status"].(float64) != 1 {
		return errors.New(jsonBody["msg"].(string))
	}
	return nil
}

func getClientIdByVkey(webUIAddr string, vkey string) (int, error) {
	url := fmt.Sprintf("http://%s/client/list", webUIAddr)
	payload := fmt.Sprintf("search=%s&order=asc&offset=0&limit=10", vkey)
	body, err := postUrlEncoded(url, payload)
	if err != nil {
		return 0, err
	}
	jsonBody := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &jsonBody); err != nil {
		return 0, err
	}
	rows := jsonBody["rows"].([]interface{})
	if len(rows) == 0 {
		return 0, errors.New("no client found")
	}
	for _, v := range rows {
		if v.(map[string]interface{})["VerifyKey"].(string) == vkey {
			return int(v.(map[string]interface{})["Id"].(float64)), nil
		}
	}
	return 0, errors.New("no client found")
}

func deleteClient(webUIAddr string, clientId int) error {
	url := fmt.Sprintf("http://%s/client/del", webUIAddr)
	payload := fmt.Sprintf("id=%d", clientId)
	body, err := postUrlEncoded(url, payload)
	if err != nil {
		return err
	}
	jsonBody := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &jsonBody); err != nil {
		return err
	}
	if jsonBody["status"].(float64) != 1 {
		return errors.New(jsonBody["msg"].(string))
	}
	return nil
}

func getBridgePort(webUIAddr string) (int, error) {
	url := fmt.Sprintf("http://%s/client/list", webUIAddr)
	payload := "search=&order=asc&offset=0&limit=10"
	body, err := postUrlEncoded(url, payload)
	if err != nil {
		return 0, err
	}
	jsonBody := make(map[string]interface{})
	if err := json.Unmarshal([]byte(body), &jsonBody); err != nil {
		return 0, err
	}
	return int(jsonBody["bridgePort"].(float64)), nil
}

func postUrlEncoded(path string, data string) (string, error) {
	url := path
	method := "POST"
	timestamp := time.Now().Unix()
	auth_key := crypt.Md5(strconv.FormatInt(timestamp, 10))
	payload := strings.NewReader(data + fmt.Sprintf("&auth_key=%s&timestamp=%d", auth_key, timestamp))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
