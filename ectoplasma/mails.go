package ectoplasma

import (
	"context"
	"github.com/pkg/errors"
	"github.com/podded/podded/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	QueryFilter struct {
		// Entity Filters
		CharacterID   int32 `schema:"character_id"`
		CorporationID int32 `schema:"corporation_id"`
		AllianceID    int32 `schema:"alliance_id"`

		// Location Filter
		SolarSystem   int32 `schema:"solar_system"`
		Constellation int32 `schema:"constellation"`
		Region        int32 `schema:"region"`

		// Pagination occurs on all requests
		Page int `schema:"page"`
	}
)

func (sludge *PodGoo) getKillmailsRaw(ctx context.Context, filter bson.M, options *options.FindOptions) (err error, mails []types.ESIKillmail) {

	if err := sludge.BounceMongo(); err != nil {
		return err, nil
	}

	kms := []types.ESIKillmail{}
	col := sludge.MongoClient.Database("podded").Collection("killmails")

	cursor, err := col.Find(ctx, filter, options)
	if err != nil{
		return errors.Wrap(err, "failed to query mongo for killmails"), nil
	}

	for cursor.Next(ctx) {
		var km types.ESIKillmail
		err := cursor.Decode(&km)
		if err != nil {
			return errors.Wrap(err, "error decoding response from mongo"), nil
		}
		kms = append(kms, km)

	}

	err = cursor.Err()
	if err != nil {
		return errors.Wrap(err, "mongo cursor error"), nil
	}

	cursor.Close(ctx)

	return nil, kms
}

func (sludge *PodGoo) KillmailFetchIndividual(ctx context.Context, id int) (err error, mail types.ESIKillmail) {

	filter := bson.M{
		"_id": id,
	}

	opt := options.Find()

	err, mails := sludge.getKillmailsRaw(ctx, filter, opt)
	if err != nil {
		return errors.Wrap(err, "failed to query mongo"), mail
	}

	if len(mails) != 1 {
		return errors.New("killmail does not exist"), mail
	}

	return nil, mails[0]
}

func (sludge *PodGoo) KillmailFetchRecent(ctx context.Context, page int) (err error, mails []types.ESIKillmail) {
	// Pagination logic
	var skip int64
	lim := int64(50)

	if page > 0 {
		skip = int64(page) * lim
	}

	opt := options.Find().SetLimit(lim).SetSort(bson.M{"_id": -1}).SetSkip(skip)

	filter := bson.M{
		"esi_v1": bson.M{"$exists":true},
	}

	return sludge.getKillmailsRaw(ctx, filter, opt)
}



func (sludge *PodGoo) KillmailFetchCharacterRecent(ctx context.Context, characterID int, page int) (err error, mails []types.ESIKillmail) {
	// Pagination logic
	var skip int64
	lim := int64(50)

	if page > 0 {
		skip = int64(page) * lim
	}

	opt := options.Find().SetLimit(lim).SetSort(bson.M{"_id": -1}).SetSkip(skip)

	filter := bson.M{
		"esi_v1": bson.M{"$exists":true},
	}

	if characterID > 0 {
		filter["$or"] = []interface{}{
			bson.M{"esi_v1.attackers.character_id": characterID},
			bson.M{"esi_v1.victim.character_id": characterID},
		}
	}

	return sludge.getKillmailsRaw(ctx, filter, opt)
}