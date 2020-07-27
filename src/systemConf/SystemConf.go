package systemConf

import (
	"flag"
	"fmt"
	"github.com/liangdas/mqant/log"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"utils"
)

var SystemConfMgr *SystemConf

type SystemConf struct {
	ConfDir    string     //配置文件夹路径
	Conf       string     //配置文件路径
	Wd         string     //工作目录
	Pid        string     //线程id
	ThPID      string     //第三方服务器线程id
	LogDir     string     //日志文件夹
	BiDir      string     //日志文件夹
	Debug      bool       //是否调试环境
	appDir     string     //app路径
	RedisConf  *RedisConf //redis配置文件数据
	MysqlConf  *MysqlConf //redis配置文件数据
	ConsulConf *ConsulConf
	NatsConf   *NatsConf
}

//Redis配置文件数据
type RedisConf struct {
	Address     string
	Database    int
	Password    string
	MaxIdle     int
	MaxActive   int
	IdleTimeout time.Duration
}

//Redis配置文件数据
type MysqlConf struct {
	Address  string
	Database string
	Account  string
	Password string
}

type ConsulConf struct {
	Address          []string
	RegisterInterval int64
	RegisterTTL      int64
	KillWaitTTL      int64
}

type NatsConf struct {
	Address       string
	MaxReconnects int
}

func InitConfig() error {
	SystemConfMgr = &SystemConf{}
	SystemConfMgr.Debug = true
	var confDir, confFile, processID, applicationDir, thProcessID, logDir, biDir string
	flag.StringVar(&confFile, "conf", "", "Server configuration file path")
	flag.StringVar(&confDir, "confdir", "", "配置文件夹路径")
	flag.StringVar(&applicationDir, "wd", "", "Server work directory")
	flag.StringVar(&processID, "pid", "development", "Server ProcessID")
	flag.StringVar(&thProcessID, "thpid", "development", "其它包括数据库等第三方服务器的统一ProcessID")
	flag.StringVar(&logDir, "log", "", "Log file directory?")
	flag.StringVar(&biDir, "bi", "", "bi file directory?")
	flag.Parse() //解析输入的参数

	SystemConfMgr.Conf = confFile
	SystemConfMgr.Wd = applicationDir
	SystemConfMgr.Pid = processID
	SystemConfMgr.ThPID = thProcessID
	SystemConfMgr.LogDir = logDir
	SystemConfMgr.BiDir = biDir

	confData, err := SystemConfMgr.loadConfData()
	if err != nil{
		log.Error("not find server.json from %s", confFile)
		return err
	}

	if confDir == "" {
		confDir = fmt.Sprintf("%s/bin/conf", SystemConfMgr.appDir)
	}

	SystemConfMgr.ConfDir = confDir
	SystemConfMgr.setRedisConf(confData)
	SystemConfMgr.setMysqlConf(confData)
	SystemConfMgr.setConsulConf(confData)
	SystemConfMgr.setNatsConf(confData)
	log.Info("Load config data complete")
	return nil
}

//加载配置文件数据
func (this *SystemConf) loadConfData() ([]byte, error) {
	appDir := this.Wd
	if appDir != "" {
		_, err := os.Open(appDir)
		if err != nil {
			panic(err)
		}
		os.Chdir(appDir)
		appDir, err = os.Getwd()
	} else {
		var err error
		appDir, err = os.Getwd()
		if err != nil {
			file, _ := exec.LookPath(os.Args[0])
			ApplicationPath, _ := filepath.Abs(file)
			appDir, _ = filepath.Split(ApplicationPath)
		}
	}
	this.appDir = appDir
	confPath := this.Conf
	if confPath == "" {
		confPath = fmt.Sprintf("%s/bin/conf/server.json", appDir)
	}

	return ioutil.ReadFile(confPath)
}

//设置Redis配置文件数据
func (this *SystemConf) setRedisConf(confData []byte) {
	redisConfAry := gjson.GetBytes(confData, "Redis").Array()
	this.RedisConf = &RedisConf{}
	for _, result := range redisConfAry {
		redisConf := result.Map()
		pid := redisConf["ProcessID"].Str
		if pid != this.ThPID {
			continue
		}
		this.RedisConf.Address = redisConf["Address"].Str
		this.RedisConf.Database = int(redisConf["Database"].Int())
		this.RedisConf.Password = redisConf["Password"].Str
		this.RedisConf.MaxIdle = int(redisConf["MaxIdle"].Int())
		this.RedisConf.MaxActive = int(redisConf["MaxActive"].Int())
		this.RedisConf.IdleTimeout = time.Duration(redisConf["IdleTimeout"].Int())
	}
}

func (this *SystemConf) setMysqlConf(confData []byte) {
	redisConfAry := gjson.GetBytes(confData, "Mysql").Array()
	this.MysqlConf = &MysqlConf{}
	for _, result := range redisConfAry {
		redisConf := result.Map()
		pid := redisConf["ProcessID"].Str
		if pid != this.ThPID {
			continue
		}
		this.MysqlConf.Address = redisConf["Address"].Str
		this.MysqlConf.Database = redisConf["Database"].Str
		this.MysqlConf.Account = redisConf["Account"].Str
		this.MysqlConf.Password = redisConf["Password"].Str
	}
}

func (this *SystemConf) setConsulConf(confData []byte) {
	ary := gjson.GetBytes(confData, "Consul").Array()
	for _, result := range ary {
		consulConf := result.Map()
		pid := consulConf["ProcessID"].Str
		if pid != this.ThPID {
			continue
		}
		consulAddress := utils.StringToArray(consulConf["Address"].Raw, ",")
		RegisterInterval := consulConf["RegisterInterval"].Int()
		RegisterTTL := consulConf["RegisterTTL"].Int()
		KillWaitTTL := consulConf["KillWaitTTL"].Int()
		this.ConsulConf = &ConsulConf{
			Address:          consulAddress,
			RegisterInterval: RegisterInterval,
			RegisterTTL:      RegisterTTL,
			KillWaitTTL:      KillWaitTTL,
		}
	}
}

func (this *SystemConf) setNatsConf(confData []byte) {
	ary := gjson.GetBytes(confData, "Nats").Array()
	for _, result := range ary {
		natsConf := result.Map()
		pid := natsConf["ProcessID"].Str
		if pid != this.ThPID {
			continue
		}
		this.NatsConf = &NatsConf{
			Address:       "nats://" + natsConf["Address"].Str,
			MaxReconnects: int(natsConf["MaxReconnects"].Int()),
		}
	}
}


