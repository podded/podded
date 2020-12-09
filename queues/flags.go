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

type Flags struct {
	goop *ectoplasma.PodGoo
}

func NewFlags(goop *ectoplasma.PodGoo) *Flags {
	return &Flags{goop: goop}
}

func (c *Flags) StartFlags() (err error) {

	if c == nil || c.goop == nil {
		return errors.New("Need to init ESIKillmailIngest / goop struct")
	}

	ctx := context.Background()

	kmdb := c.goop.MongoClient.Database("podded").Collection("killmails")

	// Start calcumelating
	for {
		// Check if any available IDS in the cleaning queue
		res, err := GetFromFlagsQueueBlocking(c.goop.RedisClient)
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

		km := types.ESIKillmailRaw{}

		err = kmdb.FindOne(ctx, bson.M{"_id": killid}).Decode(&km)
		if err != nil {
			// Welp!
			log.Printf("Failed to get hash for ID %d, err: %s", killid, err)
			continue
		}

		if km.ID == 0 {
			log.Fatal("0 Value KM")
		}

		flg := types.Flags{}

		flg.NPC = c.isNPC(km)

		if !flg.NPC {
			flg.Awox, err = c.isAwox(km)
			if err != nil {
				AddToErrorQueue(km.ID, fmt.Sprintf("failed to calc awox queues: %v", err), flags, c.goop.RedisClient)
				continue
			}
			flg.Solo, err = c.isSolo(km)
			if err != nil {
				AddToErrorQueue(km.ID, fmt.Sprintf("failed to calc solo queues: %v", err), flags, c.goop.RedisClient)
				continue
			}
		}


		filter := bson.M{"_id": km.ID}
		update := bson.M{"$set": bson.M{"flags": flg}}
		_, _ = kmdb.UpdateOne(ctx, filter, update) // Not fussed about the error here // TODO Check the error

		// We have successfully scraped the killmail and updated the database
		CleanCompleted(km.ID, c.goop.RedisClient)

		continue

	}
}

func (c *Flags) isNPC(km types.ESIKillmailRaw) (res bool) {

	if km.Killmail.Victim.CharacterID == 0 && km.Killmail.Victim.CorporationID > 1 && km.Killmail.Victim.CorporationID < 1999999 {
		return true
	}

	for _, a := range km.Killmail.Attackers {
		if a.CharacterID > 3999999{
			return false
		}

		if a.CorporationID > 1999999 {
			return false
		}
	}

	return true

}

func (c *Flags) isAwox(km types.ESIKillmailRaw) (res bool, err error) {

	var group int32
	err = c.goop.MariaClient.QueryRow("select groupID from invTypes where typeID = ?", km.Killmail.Victim.ShipTypeID).Scan(&group)
	if err != nil {
		return false, err
	}
	switch group {
	// No solo kills on Capsules, Corvettes
	case 29, 237:
		return false, nil
	}

	if km.Killmail.Victim.CorporationID == 0{
		return false, nil
	}

	for _, a := range km.Killmail.Attackers {
		if !a.FinalBlow {
			continue
		}
		if a.CorporationID <= 1999999 {
			continue
		}
		if a.CorporationID == km.Killmail.Victim.CorporationID{
			return true, nil
		}
	}

	return
}

func (c *Flags) isSolo(km types.ESIKillmailRaw) (res bool, err error) {

	var group int32
	err = c.goop.MariaClient.QueryRow("select groupID from invTypes where typeID = ?", km.Killmail.Victim.ShipTypeID).Scan(&group)
	if err != nil {
		return false, err
	}
	switch group {
	// No solo kills on Capsules, Shuttles or Corvettes
	case 29, 31, 237:
		return false, nil
	}

	// Make sure this is a ship
	var category int32
	err = c.goop.MariaClient.QueryRow("select categoryID from invGroups where groupID = ?", group).Scan(&category)
	if err != nil {
		return false, err
	}
	if category != 6 {
		return false, nil
	}

	numPlayers := 0
	for _, a := range km.Killmail.Attackers{
		if a.CharacterID > 3999999 {
			numPlayers++
		}
		if numPlayers > 1 {
			return false, nil
		}
		var group, category int32
		err = c.goop.MariaClient.QueryRow("select groupID from invTypes where typeID = ?", a.ShipTypeID).Scan(&group)
		if err != nil {
			return false, err
		}
		err = c.goop.MariaClient.QueryRow("select categoryID from invGroups where groupID = ?", group).Scan(&category)
		if err != nil {
			return false, err
		}
		// Citadels can help
		if category == 65 {
			return false, nil
		}
	}

	return numPlayers == 1, nil
}

func (c *Flags) calculateMeta(km types.ESIKillmailRaw) (mean, median, mode, min, max int32, err error) {

	for _, i := range km.Killmail.Victim.Items {
		if isFitted(i) {
			var _ int32
			//err =
		}
	}


	return
}

func isFitted(item types.VictimItem) (fitted bool) {

	return

}

