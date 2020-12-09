package main

import (
	"github.com/gobuffalo/envy"
	"github.com/pkg/profile"
	"github.com/podded/podded/api"
	"github.com/podded/podded/ectoplasma"
	"github.com/podded/podded/higgs"
	"github.com/podded/podded/maintainer"
	"github.com/podded/podded/queues"
	"github.com/podded/podded/server"
	"github.com/urfave/cli/v2"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "net/http/pprof"
)

func main() {

	defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()


	app := &cli.App{
		Name:                 "Podded",
		Description:          "podded - eve online killboard",
		EnableBashCompletion: true,
		Authors: []*cli.Author{
			{
				Name:  "Crypta Electrica",
				Email: "crypta@crypta.tech",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "run the podded api server",
				Subcommands: []*cli.Command{
					{
						Name:    "api",
						Aliases: []string{"a"},
						Usage:   "api server",
						Action:  apiServer,
					},
					{
						Name: "web",
						Aliases: []string{"w"},
						Usage: "web server",
						Action: serve,
					},
				},
			},
			{
				Name:    "maintain",
				Aliases: []string{"m"},
				Usage:   "Keep things healthy",
				Action:  maintain,
			},
			{
				Name:    "queue",
				Aliases: []string{"q"},
				Usage:   "run a queue worker",
				Subcommands: []*cli.Command{
					{
						Name:    "ingest",
						Aliases: []string{"i"},
						Usage:   "run the ingest queue",
						Action:  ingest,
					},
					{
						Name:    "flag",
						Aliases: []string{"f"},
						Usage:   "run the flag generation queue",
						Action:  flag,
					},
					{
						Name:    "clean",
						Aliases: []string{"c"},
						Usage:   "run the cleaner queue",
						Action:  clean,
					},
				},
			},
			{
				Name:    "higgs",
				Aliases: []string{"h"},
				Usage:   "update all static data",
				Action:  hgs,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getNewPod() (goop *ectoplasma.PodGoo, err error) {
	err = envy.Load()
	if err != nil {
		return
	}

	//Bouncer
	bouncerAddress := envy.Get("BOUNCER_ADDRESS", "http://localhost:13271")
	timeoutEnv := envy.Get("BOUNCER_TIMEOUT", "10")
	descriptor := envy.Get("BOUNCER_DESC", "ecto_api_server")

	timeout := 10
	i, err := strconv.Atoi(timeoutEnv)
	if err == nil {
		timeout = i
	}

	//Mongo
	mongoHost := envy.Get("MONGO_HOST", "localhost")
	mongoPortEnv := envy.Get("MONGO_PORT", "27017")

	mongoPort := 27017
	i, err = strconv.Atoi(mongoPortEnv)
	if err == nil {
		mongoPort = i
	}

	//Redis
	redisHost := envy.Get("REDIS_HOST", "localhost")
	redisPassword := envy.Get("REDIS_PASS", "")
	redisPortEnv := envy.Get("REDIS_PORT", "6379")
	redisDBEnv := envy.Get("REDIS_DB", "0")

	redisPort := 6379
	i, err = strconv.Atoi(redisPortEnv)
	if err == nil {
		redisPort = i
	}

	redisDB := 0
	i, err = strconv.Atoi(redisDBEnv)
	if err == nil {
		redisDB = i
	}
	//Maria
	mariaHost := envy.Get("MARIA_HOST", "localhost")
	mariaUser := envy.Get("MARIA_USER", "podded")
	mariaPass := envy.Get("MARIA_PASS", "podded")
	mariaDatabase := envy.Get("MARIA_DATABASE", "podded")

	pgc := ectoplasma.PodGooConfig{
		BouncerAddress:    bouncerAddress,
		BouncerTimeout:    time.Duration(timeout) * time.Second,
		BouncerDescriptor: descriptor,
		MongoHost:         mongoHost,
		MongoPort:         mongoPort,
		RedisHost:         redisHost,
		RedisPort:         redisPort,
		RedisPassword:     redisPassword,
		RedisDatabase:     redisDB,
		MariaHost:         mariaHost,
		MariaUser:         mariaUser,
		MariaPass:         mariaPass,
		MariaDatabase:     mariaDatabase,
	}

	goop = ectoplasma.NewPodGoo(pgc)
	return
}

func apiServer(c *cli.Context) error {

	goop, err := getNewPod()
	if err != nil {
		return err
	}

	// API Details
	apiHost := envy.Get("API_HOST", "0.0.0.0")
	apiPortEnv := envy.Get("API_PORT", "13270")

	apiPort := 13270
	i, err := strconv.Atoi(apiPortEnv)
	if err == nil {
		apiPort = i
	}

	if err = goop.ConnectBouncer(); err != nil {
		log.Fatal(err)
	}

	if err = goop.ConnectMongo(); err != nil {
		log.Fatal(err)
	}

	if err = goop.ConnectRedis(); err != nil {
		log.Fatal(err)
	}

	as := api.ApiServer{
		Goop: goop,
		Host: apiHost,
		Port: apiPort,
	}

	return as.ListenAndServe()
}

func maintain(c *cli.Context) error {
	goop, err := getNewPod()
	if err != nil {
		return err
	}

	if err = goop.ConnectMongo(); err != nil {
		log.Fatal(err)
	}

	if err = goop.ConnectRedis(); err != nil {
		log.Fatal(err)
	}

	mt := maintainer.NewMaintainer(goop)

	return mt.StartMaintainer()
}

func clean(c *cli.Context) error {
	goop, err := getNewPod()
	if err != nil {
		return err
	}

	if err = goop.ConnectMongo(); err != nil {
		log.Fatal(err)
	}

	if err = goop.ConnectRedis(); err != nil {
		log.Fatal(err)
	}

	if err = goop.ConnectMaria(); err != nil {
		log.Fatal(err)
	}

	mt := queues.NewCleaner(goop)

	return mt.StartCleaner()
}

func flag(c *cli.Context) error {
	goop, err := getNewPod()
	if err != nil {
		return err
	}

	if err = goop.ConnectMongo(); err != nil {
		log.Fatal(err)
	}

	if err = goop.ConnectRedis(); err != nil {
		log.Fatal(err)
	}

	if err = goop.ConnectMaria(); err != nil {
		log.Fatal(err)
	}

	mt := queues.NewFlags(goop)

	return mt.StartFlags()
}

func ingest(c *cli.Context) error {
	goop, err := getNewPod()
	if err != nil {
		return err
	}

	if err = goop.ConnectBouncer(); err != nil {
		log.Fatal(err)
	}

	if err = goop.ConnectMongo(); err != nil {
		log.Fatal(err)
	}

	if err = goop.ConnectRedis(); err != nil {
		log.Fatal(err)
	}

	esi := queues.NewESIKillmailIngest(goop)

	return esi.StartScraper()
}

func hgs(c *cli.Context) error {
	goop, err := getNewPod()
	if err != nil {
		return err
	}

	if err = goop.ConnectBouncer(); err != nil {
		log.Fatal(err)
	}

	if err = goop.ConnectMongo(); err != nil {
		log.Fatal(err)
	}

	if err = goop.ConnectMaria(); err != nil {
		log.Fatal(err)
	}

	mariaHost := envy.Get("MARIA_HOST", "localhost")
	mariaUser := envy.Get("MARIA_USER", "podded")
	mariaPass := envy.Get("MARIA_PASS", "podded")

	h := higgs.NewHiggs(goop, mariaUser, mariaPass, mariaHost)

	// PPROF
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	//if err = h.DeleteStaticData(); err != nil {
	//	return err
	//}
	//
	//return h.PopulateStaticData()

	return h.PopulateAllHistoricalSDE()
}

func serve(C *cli.Context) error {
	goop, err := getNewPod()
	if err != nil {
		return err
	}

	if err = goop.ConnectMongo(); err != nil {
		log.Fatal(err)
	}

	if err = goop.ConnectMaria(); err != nil {
		log.Fatal(err)
	}

	srv := server.NewServer(goop)
	return srv.RunServer()
}
