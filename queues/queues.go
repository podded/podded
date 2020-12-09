package queues

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/podded/podded/types"
	"log"
)

const (
	ingest       = "PODDED:Q:INGEST"
	cleaner      = "PODDED:Q:CLEAN"
	flags        = "PODDED:Q:FLAGS"
	interactions = "PODDED:Q:INTERACT"
	errored      = "PODDED:Q:ERROR"
)

// Queue Getters

func GetFromIngestQueueBlocking(client *redis.Client) ([]string, error) {
	return client.BLPop(0, ingest).Result()
}

func GetFromCleanQueueBlocking(client *redis.Client) ([]string, error) {
	return client.BLPop(0, cleaner).Result()
}

func GetFromFlagsQueueBlocking(client *redis.Client) ([]string, error) {
	return client.BLPop(0, flags).Result()
}

// Populate queues

func AddToIngest(killID int32, client *redis.Client) {
	client.RPush(ingest, killID)
}

// Queue Complete!

func IngestCompleted(killID int32, client *redis.Client) {
	client.RPush(cleaner, killID)
	client.RPush(flags, killID)
}

func CleanCompleted(killID int32, client *redis.Client) {
	client.RPush(interactions, killID)
}

// For when shit goes wrong

func AddToErrorQueue(killID int32, message string, source string, client *redis.Client) {
	msg := types.ErrorMessage{
		KillID:  killID,
		Message: message,
		SourceQueue: source,
	}

	bt, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(errors.Wrap(err, fmt.Sprintf("Failed to add to error queue: %d", killID)))
	}

	client.RPush(errored, string(bt))
}
