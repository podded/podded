package queues

import (
	"context"
	"errors"
	"fmt"
	"github.com/podded/podded/ectoplasma"
	"github.com/podded/podded/types"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"strconv"
)

type Cleaner struct {
	goop *ectoplasma.PodGoo
}

const (
	changeoverUnix = 1602586800
)

func NewCleaner(goop *ectoplasma.PodGoo) *Cleaner {
	return &Cleaner{goop: goop}
}

func (c *Cleaner) StartCleaner() (err error) {

	if c == nil || c.goop == nil {
		return errors.New("need to init ESIKillmailIngest / goop struct")
	}

	ctx := context.Background()

	kmdb := c.goop.MongoClient.Database("podded").Collection("killmails")

	// Start cleaning
	for {
		// Check if any available IDS in the cleaning queue
		res, err := GetFromCleanQueueBlocking(c.goop.RedisClient)
		if err != nil {
			log.Println(err)
			continue
		}
		// res[0] will always contain the queue name, res[1] is the value we are getting out of redis

		killid, err := strconv.Atoi(res[1])
		if err != nil {
			// TODO Load this onto the error queue to be looked at later
			log.Printf("Invalid kill id popped from queue %s", res[1])
		}
		log.Printf("DEBUG ID:%v", killid)

		km := types.ESIKillmail{}


		err = kmdb.FindOne(ctx, bson.M{"_id": killid}).Decode(&km)
		if err != nil {
			// Welp!
			AddToErrorQueue(int32(killid), fmt.Sprintf("Failed to get hash for ID %d, err: %s", killid, err), cleaner, c.goop.RedisClient)
			continue
		}

		if km.ID == 0 {
			AddToErrorQueue(int32(killid), "got 0 value killmail from mongo", cleaner, c.goop.RedisClient)
			continue
		}

		kmc, err := c.CleanSystemAttributes(km)
		if err != nil {
			AddToErrorQueue(int32(killid), fmt.Sprintf("Failed to clean, err: %s", err), cleaner, c.goop.RedisClient)
			continue
		}

		filter := bson.M{"_id": km.ID}
		update := bson.M{"$set": bson.M{"esi_v1": kmc.Killmail}}
		_, _ = kmdb.UpdateOne(ctx, filter, update) // Not fussed about the error here // TODO Check the error

		// We have successfully scraped the killmail and updated the database
		CleanCompleted(km.ID, c.goop.RedisClient)

		continue


	}
}

func(c *Cleaner) CleanSystemAttributes(km types.ESIKillmail) (res types.ESIKillmail, err error)  {


	if km.Killmail.KillmailTime.Unix() <= changeoverUnix {
		// This is pre povchen
		var conID, regID int32
		err := c.goop.MariaClient.QueryRow("select constellationID, regionID from map_solar_systems_2020_10_08 where solarSystemID = ?", km.Killmail.SolarSystemID).Scan(&conID, &regID)
		if err != nil {
			return res, err
		}
		km.Killmail.ConstellationID = conID
		km.Killmail.RegionID = regID
	} else {
		// This is post povchen
		var conID, regID int32
		err := c.goop.MariaClient.QueryRow("select constellationID, regionID from map_solar_systems_2020_10_13 where solarSystemID = ?", km.Killmail.SolarSystemID).Scan(&conID, &regID)
		if err != nil {
			return res, err
		}
		km.Killmail.ConstellationID = conID
		km.Killmail.RegionID = regID

	}

	return km, nil
}