package cmd

import (
	"crypto/tls"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-redis/redis"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vinothzomato/go-foreman/foreman"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// LogLevel of the server
var LogLevel string

// Config file location
var cfgFile string

// Foreman client
var client *foreman.Client

// Foreman client
var redisClient *redis.Client

// Memcache client
var memcacheClient *memcache.Client

var (
	ip        string
	port      int
	baseurl   string
	username  string
	password  string
	zone      string
	cacheType string
	ttl       int
	logFile   string
)

// ServerCmd for server sub command to start the server
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "start the DNS server",
	Long:  `Starts the DNS server`,
	Run: func(cmd *cobra.Command, args []string) {
		if baseurl == "" {
			log.Panic("Forman url cannot be empty. Pleae run with --url or add url in the foremandns.yaml config")
			os.Exit(1)
		}
		if username == "" {
			log.Panic("Forman username cannot be empty. Pleae run with --username or add username in the foremandns.yaml config")
			os.Exit(1)
		}
		if password == "" {
			log.Panic("Forman password cannot be empty. Pleae run with --password or add url in the foremandns.yaml config")
			os.Exit(1)
		}

		log.Info(fmt.Sprintf("Starting the server on ip %s port %d \n", ip, port))

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		transport := &foreman.BasicAuthTransport{
			Username:  username,
			Password:  password,
			Transport: tr,
		}
		httpClient := &http.Client{Transport: transport}

		client = foreman.NewClient(httpClient)
		url, _ := url.Parse(baseurl)
		client.BaseURL = url

		switch cacheType {
		case "redis":
			redisServer := viper.GetString("redis.server")
			redisPassword := viper.GetString("redis.password")
			fmt.Printf("Redis server %v password %v\n", redisServer, redisPassword)
			redisClient = redis.NewClient(&redis.Options{
				Addr:     redisServer,
				Password: redisPassword,
				DB:       0,
			})

		case "memcache":
			memcacheServer := viper.GetString("memcache.server")
			fmt.Printf("Memcache Server %s\n", memcacheServer)
			memcacheClient = memcache.New(memcacheServer)

		}

		srv := &dns.Server{Addr: ip + ":" + strconv.Itoa(port), Net: "udp"}
		srv.Handler = &handler{}
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("Failed to set udp listener %s", err.Error())
		}
	},
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	cobra.OnInitialize(initConfig)

	ServerCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Config file path")

	ServerCmd.Flags().StringVarP(&ip, "ip", "i", "0.0.0.0", "Server listen ip address default is 0.0.0.0")
	ServerCmd.Flags().IntVarP(&port, "port", "t", 53, "Server listen port default is 53")
	ServerCmd.Flags().StringVarP(&LogLevel, "log-level", "l", "info", "Log level e.g. debug, info, warning & error")
	ServerCmd.Flags().StringVarP(&logFile, "log", "", "", "Log file path")
	ServerCmd.Flags().StringVarP(&baseurl, "url", "f", "", "Foreman base url e.g. https://foreman.example.com/")
	ServerCmd.Flags().StringVarP(&username, "username", "u", "", "Foreman username")
	ServerCmd.Flags().StringVarP(&password, "password", "p", "", "Foreman password")
	ServerCmd.Flags().StringVarP(&zone, "zone", "z", "", "Custom DNS zone for the hosts")
	ServerCmd.Flags().StringVarP(&cacheType, "cache-type", "", "", "Cache type e.g. redis, memory, memcached")
	ServerCmd.Flags().IntVarP(&ttl, "ttl", "", 1800, "Cache expiry time default 1800 seconds(30min)")

	viper.BindPFlag("ip", ServerCmd.Flags().Lookup("ip"))
	viper.BindPFlag("port", ServerCmd.Flags().Lookup("port"))
	viper.BindPFlag("url", ServerCmd.Flags().Lookup("url"))
	viper.BindPFlag("username", ServerCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", ServerCmd.Flags().Lookup("password"))
	viper.BindPFlag("zone", ServerCmd.Flags().Lookup("zone"))
	viper.BindPFlag("cache-type", ServerCmd.Flags().Lookup("cache-type"))
	viper.BindPFlag("log-level", ServerCmd.Flags().Lookup("log-level"))
	viper.BindPFlag("log", ServerCmd.Flags().Lookup("log"))
	viper.BindPFlag("ttl", ServerCmd.Flags().Lookup("ttl"))

	log.SetLevel(log.InfoLevel)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("foremandns")
		viper.AddConfigPath("/etc/foremandns/")
		viper.AddConfigPath("$HOME/.foremandns")
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		//log.Panic("Can't read config:", err)
		//os.Exit(1)
	}
	if ip == "0.0.0.0" {
		ip = viper.GetString("ip")
	}
	if port == 53 {
		port = viper.GetInt("port")
	}
	if baseurl == "" {
		baseurl = viper.GetString("url")
	}
	if username == "" {
		username = viper.GetString("username")
	}
	if password == "" {
		password = viper.GetString("password")
	}
	if zone == "" {
		zone = viper.GetString("zone")
	}
	if cacheType == "" {
		cacheType = viper.GetString("cache-type")
	}
	if LogLevel == "" {
		LogLevel = viper.GetString("log-level")
	}
	if logFile == "" {
		logFile = viper.GetString("log")
	}

	if ttl == 1800 {
		ttl = viper.GetInt("ttl")
	}

	switch LogLevel {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "erro":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			log.SetOutput(file)
		} else {
			log.Error("Failed to log to file, using default stderr")
		}
	} else {
		log.SetOutput(os.Stdout)
	}
}
