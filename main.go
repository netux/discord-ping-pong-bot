package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	yaml "gopkg.in/yaml.v2"
)

type cfg struct {
	Token            string   `yaml:"token"`
	PingPrefix       string   `yaml:"ping-prefix"`
	PongPrefix       string   `yaml:"pong-prefix"`
	ChannelWhitelist []string `yaml:"channel-whitelist"`
}

// Cfg contains parsed configuration from config.yaml
var Cfg cfg

func conf() {
	// Recover from a panic, output as config parse error and exit
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("error parsing config, ", r)
		}
	}()

	// Read the whole file and store it's bytes into b.
	b, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			panic(errors.New("config.yaml doesn't exist"))
		} else {
			panic(err)
		}
	}

	// Parse bytes into YAML and assign values to Cfg
	err = yaml.Unmarshal(b, &Cfg)
	if err != nil {
		panic(err)
	}
}

func setupHandlers(s *discordgo.Session) {
	s.AddHandler(handleMessageCreate)
}

func run(s *discordgo.Session) {
	// Open a websocket connection to Discord and begin listening.
	err := s.Open()
	if err != nil {
		log.Fatal("error opening connection, ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	err = s.Close()
	if err != nil {
		log.Fatal("error closing connection, ", err)
	}
}

func main() {
	// Setup Cnf
	conf()
	if Cfg.Token == "" {
		log.Fatal("missing 'token' parameter in config.yaml")
	}

	SetRegexpPingpong(Cfg.PingPrefix, Cfg.PongPrefix)

	// Setup rand
	rand.Seed(time.Now().Unix())

	// Setup BOT
	s, err := discordgo.New("Bot " + Cfg.Token)
	if err != nil {
		log.Fatal(err)
	}

	setupHandlers(s)

	run(s)
}
