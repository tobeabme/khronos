package khronos

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/go-ini/ini"
	"github.com/voxelbrain/goptions"
)

// Config stores all configuration options for the khronos package.
type Configuration struct {
	Runmode  string
	NodeName string
	LogLevel string
	//for communicating between servers in cluster
	BindIP   string
	BindPort int
	RPCPort  int
	//storage e.g. etcd,etcdv3
	Backend         string
	BackendMachines []string
	Keyspace        string

	MailHost          string
	MailPort          int
	MailUsername      string
	MailPassword      string
	MailFrom          string
	MailPayload       string
	MailSubjectPrefix string
}

type SROptions struct {
	Env  string        `goptions:"-e, --env, description='The environment you want to run in (local, dev, sit, prod)'"`
	Help goptions.Help `goptions:"-h, --help, description='Show this help'"`
}

var (
	Options SROptions
	Config  *Configuration
)

// NewConfig creates a Config object and will set up the khronos configuration using
// the command line and any file configs.
func NewConfig() *Configuration {
	//Get system running parameters
	// Options.Env = "dev"
	goptions.ParseAndFail(&Options)
	if !StringInSlice(Options.Env, []string{"local", "dev", "sit", "prod"}) {
		err := errors.New("Wrong argument, environment should be in (local, dev, sit, prod)")
		if err != nil {
			fmt.Println(err.Error())
			log.Debug(err.Error())
		}
		panic(err)
	}
	Config = ReadConfig()
	return Config
}

// ReadConfig from file and create the actual config object.
func ReadConfig() *Configuration {
	env := Options.Env
	cfgFile := "./conf/config." + env + ".conf"
	cfg, err := ini.Load(cfgFile)
	if err != nil {
		panic(err)
	}

	return &Configuration{
		Runmode:         Options.Env,
		NodeName:        cfg.Section("").Key("node-name").String(),
		LogLevel:        cfg.Section("").Key("log-level").String(),
		BindIP:          cfg.Section("").Key("bind-ip").String(),
		BindPort:        cfg.Section("").Key("bind-port").MustInt(),
		RPCPort:         cfg.Section("").Key("rpc-port").MustInt(),
		Backend:         cfg.Section("").Key("backend").String(),
		BackendMachines: cfg.Section("").Key("backend-machines").Strings(","),
		Keyspace:        cfg.Section("").Key("keyspace").String(),

		MailHost:          cfg.Section("").Key("mail-host").String(),
		MailPort:          cfg.Section("").Key("mail-port").MustInt(),
		MailUsername:      cfg.Section("").Key("mail-username").String(),
		MailPassword:      cfg.Section("").Key("mail-password").String(),
		MailFrom:          cfg.Section("").Key("mail-from").String(),
		MailPayload:       cfg.Section("").Key("mail-payload").String(),
		MailSubjectPrefix: cfg.Section("").Key("mail-subject-prefix").String(),
	}
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
