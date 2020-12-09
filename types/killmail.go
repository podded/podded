package types

import "time"

type (
	IDHashPair struct {
		ID   int32  `json:"id" bson:"_id"`
		Hash string `json:"hash" bson:"hash"`
	}

	ESIKillmailRaw struct {
		// IDHashPair
		ID   int32  `json:"id" bson:"_id"`
		Hash string `json:"hash" bson:"hash"`
		//NEW
		Killmail RawMail `json:"esi_v1" bson:"esi_v1"`
	}

	ESIKillmail struct {
		// IDHashPair
		ID   int32  `json:"id" bson:"_id"`
		Hash string `json:"hash" bson:"hash"`
		//NEW
		Killmail CleanMail `json:"esi_v1" bson:"esi_v1"`
	}

	Flags struct {
		MetaMedian int32
		MetaMean   float32
		MetaMode   int32
		MetaMin    int32
		MetaMax    int32

		Awox bool
		NPC  bool
		Solo bool
	}

	RawMail struct {
		Victim        Victim     `json:"victim,omitempty" bson:"victim"`
		Attackers     []Attacker `json:"attackers,omitempty" bson:"attackers"`             /* attackers array */
		KillmailTime  time.Time  `json:"killmail_time,omitempty" bson:"killmail_time"`     /* Time that the victim was killed and the killmail generated  */
		KillmailID    int32      `json:"killmail_id,omitempty" bson:"killmail_id"`         /* ID of the killmail */
		MoonID        int32      `json:"moon_id,omitempty" bson:"moon_id"`                 /* Moon if the kill took place at one */
		SolarSystemID int32      `json:"solar_system_id,omitempty" bson:"solar_system_id"` /* Solar system that the kill took place in  */
		WarID         int32      `json:"war_id,omitempty" bson:"war_id"`                   /* War if the killmail is generated in relation to an official war  */
		ETag          string     `json:"etag,omitempty" bson:"etag"`                       /* The etag returned with the request */
	}

	CleanMail struct {
		Victim          Victim     `json:"victim,omitempty" bson:"victim"`
		Attackers       []Attacker `json:"attackers,omitempty" bson:"attackers"`               /* attackers array */
		KillmailTime    time.Time  `json:"killmail_time,omitempty" bson:"killmail_time"`       /* Time that the victim was killed and the killmail generated  */
		KillmailID      int32      `json:"killmail_id,omitempty" bson:"killmail_id"`           /* ID of the killmail */
		MoonID          int32      `json:"moon_id,omitempty" bson:"moon_id"`                   /* Moon if the kill took place at one */
		SolarSystemID   int32      `json:"solar_system_id,omitempty" bson:"solar_system_id"`   /* Solar system that the kill took place in  */
		ConstellationID int32      `json:"constellation_id,omitempty" bson:"constellation_id"` /* Solar system that the kill took place in  */
		RegionID        int32      `json:"region_id,omitempty" bson:"region_id"`               /* Solar system that the kill took place in  */
		WarID           int32      `json:"war_id,omitempty" bson:"war_id"`                     /* War if the killmail is generated in relation to an official war  */
		ETag            string     `json:"etag,omitempty" bson:"etag"`                         /* The etag returned with the request */
	}

	Attacker struct {
		AllianceID     int32   `json:"alliance_id,omitempty" bson:"alliance_id"`         /* alliance_id integer */
		CharacterID    int32   `json:"character_id,omitempty" bson:"character_id"`       /* character_id integer */
		CorporationID  int32   `json:"corporation_id,omitempty" bson:"corporation_id"`   /* corporation_id integer */
		DamageDone     int32   `json:"damage_done,omitempty" bson:"damage_done"`         /* damage_done integer */
		FactionID      int32   `json:"faction_id,omitempty" bson:"faction_id"`           /* faction_id integer */
		FinalBlow      bool    `json:"final_blow,omitempty" bson:"final_blow"`           /* Was the attacker the one to achieve the final blow  */
		SecurityStatus float32 `json:"security_status,omitempty" bson:"security_status"` /* Security status for the attacker  */
		ShipTypeID     int32   `json:"ship_type_id,omitempty" bson:"ship_type_id"`       /* What ship was the attacker flying  */
		WeaponTypeID   int32   `json:"weapon_type_id,omitempty" bson:"weapon_type_id"`   /* What weapon was used by the attacker for the kill  */
	}

	Victim struct {
		Items         []VictimItem `json:"items,omitempty" bson:"items"` /* items array */
		Position      Position     `json:"position,omitempty" bson:"position"`
		AllianceID    int32        `json:"alliance_id,omitempty" bson:"alliance_id"`       /* alliance_id integer */
		CharacterID   int32        `json:"character_id,omitempty" bson:"character_id"`     /* character_id integer */
		CorporationID int32        `json:"corporation_id,omitempty" bson:"corporation_id"` /* corporation_id integer */
		DamageTaken   int32        `json:"damage_taken,omitempty" bson:"damage_taken"`     /* How much total damage was taken by the victim  */
		FactionID     int32        `json:"faction_id,omitempty" bson:"faction_id"`         /* faction_id integer */
		ShipTypeID    int32        `json:"ship_type_id,omitempty" bson:"ship_type_id"`     /* The ship that the victim was piloting and was destroyed  */
	}

	VictimItem struct {
		Flag              int32          `json:"queues,omitempty" bson:"queues"`                             /* Flag for the location of the item  */
		ItemTypeID        int32          `json:"item_type_id,omitempty" bson:"item_type_id"`             /* item_type_id integer */
		Items             []ItemSubItems `json:"items,omitempty" bson:"items"`                           /* items array */
		QuantityDestroyed int64          `json:"quantity_destroyed,omitempty" bson:"quantity_destroyed"` /* How many of the item were destroyed if any  */
		QuantityDropped   int64          `json:"quantity_dropped,omitempty" bson:"quantity_dropped"`     /* How many of the item were dropped if any  */
		Singleton         int32          `json:"singleton,omitempty" bson:"singleton"`                   /* singleton integer */
	}

	ItemSubItems struct {
		Flag              int32 `json:"queues,omitempty" bson:"queues"`                             /* queues integer */
		ItemTypeID        int32 `json:"item_type_id,omitempty" bson:"item_type_id"`             /* item_type_id integer */
		QuantityDestroyed int64 `json:"quantity_destroyed,omitempty" bson:"quantity_destroyed"` /* quantity_destroyed integer */
		QuantityDropped   int64 `json:"quantity_dropped,omitempty" bson:"quantity_dropped"`     /* quantity_dropped integer */
		Singleton         int32 `json:"singleton,omitempty" bson:"singleton"`                   /* singleton integer */
	}

	Position struct {
		X float64 `json:"x,omitempty" bson:"x"` /* x number */
		Y float64 `json:"y,omitempty" bson:"y"` /* y number */
		Z float64 `json:"z,omitempty" bson:"z"` /* z number */
	}

	// Redis Error Queue
	ErrorMessage struct {
		KillID      int32  `json:"kill_id"`
		Message     string `json:"message"`
		SourceQueue string `json:"source_queue"`
	}
)
