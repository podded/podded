package maintainer

import (
	"context"
	"github.com/cheggaaa/pb/v3"
	"github.com/podded/podded/queues"
	"github.com/podded/podded/types"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"os"
	"strconv"
)

// orphanScrape will check for id hash pairs that are yet to be scraped and are not in the ingest queue
func (maint *Maintainer) orphanScrape() {

	ctx := context.Background()

	// Get the current ingest queue
	rq, err := maint.goop.RedisClient.LRange("PODDED:Q:INGEST", 0, -1).Result()
	if err != nil {
		log.Fatalln("Failed to read ingest queue")
	}

	// Now get the id hash pairs that have no esi data

	hashes := make(map[int32]bool)

	filter := bson.M{"esi_v1": bson.M{"$exists": false}}
	col := maint.goop.MongoClient.Database("podded").Collection("killmails")
	cursor, err := col.Find(ctx, filter, nil)
	for cursor.Next(ctx) {
		var idhp types.IDHashPair
		err := cursor.Decode(&idhp)
		if err != nil {
			log.Printf("ERROR: Failed to decode idhp: %s\n", err)
			continue
		}
		hashes[idhp.ID] = true
	}

	cursor.Close(ctx)

	for _, q := range rq {
		i, err := strconv.Atoi(q)
		if err != nil {
			// TODO: Implement Error Queue (non integer id in ingest queue)
			continue
		}

		hashes[int32(i)] = false
	}

	log.Println("Adding orphaned ingest hashes back to redis")

	bar := pb.Full.Start(len(hashes))
	bar.SetWriter(os.Stdout)

	for id, orphan := range hashes {
		if orphan {
			queues.AddToIngest(id, maint.goop.RedisClient)
		}
		bar.Increment()
	}

	bar.Finish()
}
