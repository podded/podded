package ectoplasma

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"database/sql"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/podded/bouncer/client"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	PodGoo struct {
		BoundHost string
		BoundPort int
		config    PodGooConfig

		MongoClient   *mongo.Client
		mongoBound    bool
		RedisClient   *redis.Client
		redisBound    bool
		BouncerClient *client.BouncerClient
		bouncerBound  bool
		MariaClient   *sql.DB
		mariadbBound  bool
	}

	PodGooConfig struct {
		BouncerAddress    string
		BouncerTimeout    time.Duration
		BouncerDescriptor string

		MongoHost string
		MongoPort int

		RedisHost     string
		RedisPort     int
		RedisPassword string
		RedisDatabase int

		MariaHost     string
		MariaUser     string
		MariaPass     string
		MariaDatabase string
	}
)

func (pgc PodGooConfig) Default() PodGooConfig {
	return PodGooConfig{
		BouncerAddress:    "localhost",
		BouncerTimeout:    30 * time.Second,
		BouncerDescriptor: "DEFAULT",
		MongoHost:         "localhost",
		MongoPort:         27017,
		RedisHost:         "localhost",
		RedisPort:         6379,
		RedisPassword:     "",
		RedisDatabase:     0,
		MariaHost:         "localhost",
		MariaUser:         "podded",
		MariaPass:         "podded",
		MariaDatabase:     "podded",
	}
}

func NewPodGoo(pgc PodGooConfig) (goop *PodGoo) {
	sludge := &PodGoo{config: pgc}

	return sludge
}

func (sludge *PodGoo) ConnectMongo() (err error) {

	clientOptions := options.Client().ApplyURI("mongodb://" + sludge.config.MongoHost + ":" + strconv.Itoa(sludge.config.MongoPort))
	cl, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		// TODO: Log this as a proper error
		return err
	}

	// Check the connection
	err = cl.Ping(context.TODO(), nil)
	if err != nil {
		return err
	}

	sludge.MongoClient = cl
	sludge.mongoBound = true
	return nil
}

func (sludge *PodGoo) BounceMongo() error {
	if !sludge.mongoBound {
		return errors.New("mongo not connected")
	}
	return nil
}

func (sludge *PodGoo) ConnectRedis() (err error) {
	rclient := redis.NewClient(&redis.Options{
		Addr:     sludge.config.RedisHost + ":" + strconv.Itoa(sludge.config.RedisPort),
		Password: sludge.config.RedisPassword,
		DB:       sludge.config.RedisDatabase,
	})

	pong, err := rclient.Ping().Result()
	if err != nil || pong != "PONG" {
		log.Fatalf("Failed to connect to redis: %s - %s\n", pong, err)
	}
	sludge.RedisClient = rclient
	sludge.redisBound = true
	return nil
}

func (sludge *PodGoo) BounceRedis() error {
	if !sludge.redisBound {
		return errors.New("redis not connected")
	}
	return nil
}


func (sludge *PodGoo) ConnectBouncer() (err error) {
	bc, version, err := client.NewBouncer(sludge.config.BouncerAddress, sludge.config.BouncerTimeout, sludge.config.BouncerDescriptor)
	if err != nil {
		return err
	}
	log.Printf("Connected to bouncer. version %s\n", version)

	sludge.BouncerClient = bc
	sludge.bouncerBound = true
	return nil
}

func (sludge *PodGoo) BounceBouncer() error {
	if !sludge.bouncerBound {
		return errors.New("bouncer not connected")
	}
	return nil
}

func (sludge *PodGoo) ConnectMaria() (err error) {
	const dbDriver = "mysql"
	db, err := sql.Open(dbDriver, sludge.config.MariaUser+":"+sludge.config.MariaPass+"@/"+sludge.config.MariaDatabase)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	sludge.MariaClient = db
	sludge.mariadbBound = true
	return nil
}

func (sludge *PodGoo) BounceMaria() error {
	if !sludge.mariadbBound {
		return errors.New("maria not connected")
	}
	return nil
}