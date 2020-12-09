package types

// TODO Save ETags

type (
	Region struct {
		Constellations []Constellation `json:"constellations" bson:"constellations"`
		Description    string          `json:"description" bson:"description"`
		Name           string          `json:"name" bson:"name"`
		RegionID       int             `json:"region_id" bson:"_id"`
	}

	Constellation struct {
		ConstellationID int      `json:"constellation_id" bson:"constellation_id"`
		Name            string   `json:"name" bson:"name"`
		Systems         []System `json:"systems" bson:"systems"`
		Position        Position `json:"position" bson:"position"`
	}

	System struct {
		SystemID       int        `json:"system_id" bson:"system_id"`
		Name           string     `json:"name" bson:"name"`
		Planets        []Planet   `json:"planets" bson:"planets"`
		Position       Position   `json:"position" bson:"position"`
		SecurityClass  string     `json:"security_class" bson:"security_class"`
		SecurityStatus float64    `json:"security_status" bson:"security_status"`
		Star           Star       `json:"star_id" bson:"star_id"`
		Stargates      []Stargate `json:"stargates" bson:"stargates"`
		Stations       []Station  `json:"stations" bson:"stations"`
	}

	Star struct {
		Age           int64   `json:"age" bson:"age"`
		Luminosity    float64 `json:"luminosity" bson:"luminosity"`
		Name          string  `json:"name" bson:"name"`
		Radius        int64   `json:"radius" bson:"radius"`
		SolarSystemID int     `json:"solar_system_id" bson:"solar_system_id"`
		SpectralClass string  `json:"spectral_class" bson:"spectral_class"`
		Temperature   int     `json:"temperature" bson:"temperature"`
		TypeID        int     `json:"type_id" bson:"type_id"`
		StarID        int     `json:"star_id,omitempty" bson:"star_id"`
	}

	Planet struct {
		Name          string         `json:"name" bson:"name"`
		PlanetID      int32          `json:"planet_id" bson:"planet_id"`
		Position      Position       `json:"position" bson:"position"`
		TypeID        int32          `json:"type_id" bson:"type_id"`
		Moons         []Moon         `json:"moons" bson:"moons,omitempty"`
		AsteroidBelts []AsteroidBelt `json:"asteroid_belts" bson:"asteroid_belts,omitempty"`
	}

	Moon struct {
		MoonID   int32    `json:"moon_id" bson:"moon_id"`
		Name     string   `json:"name" bson:"name"`
		Position Position `json:"position" bson:"position"`
	}

	AsteroidBelt struct {
		BeltID   int32    `json:"belt_id,omitempty" bson:"belt_id"`
		Name     string   `json:"name" bson:"name"`
		Position Position `json:"position" bson:"position"`
	}

	Stargate struct {
		StargateID  int32               `json:"stargate_id" bson:"stargate_id"`
		Destination StargateDestination `json:"destination" bson:"destination"`
		Name        string              `json:"name" bson:"name"`
		Position    Position            `json:"position" bson:"position"`
		TypeID      int32               `json:"type_id" bson:"type_id"`
	}

	StargateDestination struct {
		StargateID int32 `json:"stargate_id" bson:"stargate_id"`
		SystemID   int32 `json:"system_id" bson:"system_id"`
	}

	Station struct {
		MaxDockableShipVolume  float64  `json:"max_dockable_ship_volume" bson:"max_dockable_ship_volume"`
		Name                   string   `json:"name" bson:"name"`
		OfficeRentalCost       float64  `json:"office_rental_cost" bson:"office_rental_cost"`
		Owner                  int32    `json:"owner" bson:"owner"`
		Position               Position `json:"position" bson:"position"`
		RaceID                 int32    `json:"race_id" bson:"race_id"`
		ReprocessingEfficiency float32  `json:"reprocessing_efficiency" bson:"reprocessing_efficiency"`
		Services               []string `json:"services" bson:"services"`
		StationID              int32    `json:"station_id" bson:"station_id"`
		TypeID                 int32    `json:"type_id" bson:"type_id"`
	}

	Category struct {
		CategoryID int32   `json:"category_id" bson:"_id"`
		Groups     []Group `json:"groups" bson:"groups"`
		Name       string  `json:"name" bson:"name"`
		Published  bool    `json:"published" bson:"published"`
	}

	Group struct {
		GroupID   int32  `json:"group_id" bson:"group_id"`
		Name      string `json:"name" bson:"name"`
		Published bool   `json:"published" bson:"published"`
		Types     []Type `json:"types" bson:"types"`
	}

	Type struct {
		TypeID          int32                `json:"type_id" bson:"type_id"`
		Capacity        float64              `json:"capacity,omitempty" bson:"capacity,omitempty"`
		Description     string               `json:"description" bson:"description"`
		DogmaAttributes []TypeDogmaAttribute `json:"dogma_attributes,omitempty" bson:"dogma_attributes,omitempty"`
		DogmaEffects    []TypeDogmaEffect    `json:"dogma_effects,omitempty" bson:"dogma_effects,omitempty"`
		GraphicID       int32                `json:"graphic_id,omitempty" bson:"graphic_id,omitempty"`
		IconID          int32                `json:"icon_id,omitempty" bson:"icon_id,omitempty"`
		MarketGroupID   int32                `json:"market_group_id,omitempty" bson:"market_group_id,omitempty"`
		Mass            float64              `json:"mass,omitempty" bson:"mass,omitempty"`
		Name            string               `json:"name" bson:"name"`
		PackagedVolume  float64              `json:"packaged_volume,omitempty" bson:"packaged_volume,omitempty"`
		PortionSize     int32                `json:"portion_size,omitempty" bson:"portion_size,omitempty"`
		Published       bool                 `json:"published" bson:"published"`
		Radius          float64              `json:"radius,omitempty" bson:"radius,omitempty"`
		Volume          float64              `json:"volume,omitempty" bson:"volume,omitempty"`
	}

	TypeDogmaAttribute struct {
		AttributeID int32   `json:"attribute_id" bson:"attribute_id"`
		Value       float64 `json:"value" bson:"value"`
	}

	TypeDogmaEffect struct {
		EffectID  int32 `json:"effect_id" bson:"effect_id"`
		IsDefault bool  `json:"is_default" bson:"is_default"`
	}

	DogmaAttribute struct {
		AttributeID  int32   `json:"attribute_id" bson:"_id"`
		DefaultValue float64 `json:"default_value,omitempty" bson:"default_value"`
		Description  string  `json:"description,omitempty" bson:"description"`
		DisplayName  string  `json:"display_name,omitempty" bson:"display_name"`
		HighIsGood   bool    `json:"high_is_good,omitempty" bson:"high_is_good"`
		IconID       int32   `json:"icon_id,omitempty" bson:"icon_id"`
		Name         string  `json:"name,omitempty" bson:"name"`
		Published    bool    `json:"published,omitempty" bson:"published"`
		Stackable    bool    `json:"stackable,omitempty" bson:"stackable"`
		UnitID       int32   `json:"unit_id,omitempty" bson:"unit_id"`
	}

	DogmaEffect struct {
		Description              string                `json:"description,omitempty" bson:"description"`
		DisallowAutoRepeat       bool                  `json:"disallow_auto_repeat,omitempty" bson:"disallow_auto_repeat"`
		DischargeAttributeId     int32                 `json:"discharge_attribute_id,omitempty" bson:"discharge_attribute_id"`
		DisplayName              string                `json:"display_name,omitempty" bson:"display_name"`
		DurationAttributeId      int32                 `json:"duration_attribute_id,omitempty" bson:"duration_attribute_id"`
		EffectCategory           int32                 `json:"effect_category,omitempty" bson:"effect_category"`
		EffectId                 int32                 `json:"effect_id,omitempty" bson:"_id"`
		ElectronicChance         bool                  `json:"electronic_chance,omitempty" bson:"electronic_chance"`
		FalloffAttributeId       int32                 `json:"falloff_attribute_id,omitempty" bson:"falloff_attribute_id"`
		IconId                   int32                 `json:"icon_id,omitempty" bson:"icon_id"`
		IsAssistance             bool                  `json:"is_assistance,omitempty" bson:"is_assistance"`
		IsOffensive              bool                  `json:"is_offensive,omitempty" bson:"is_offensive"`
		IsWarpSafe               bool                  `json:"is_warp_safe,omitempty" bson:"is_warp_safe"`
		Modifiers                []DogmaEffectModifier `json:"modifiers,omitempty" bson:"modifiers"`
		Name                     string                `json:"name,omitempty" bson:"name"`
		PostExpression           int32                 `json:"post_expression,omitempty" bson:"post_expression"`
		PreExpression            int32                 `json:"pre_expression,omitempty" bson:"pre_expression"`
		Published                bool                  `json:"published,omitempty" bson:"published"`
		RangeAttributeId         int32                 `json:"range_attribute_id,omitempty" bson:"range_attribute_id"`
		RangeChance              bool                  `json:"range_chance,omitempty" bson:"range_chance"`
		TrackingSpeedAttributeId int32                 `json:"tracking_speed_attribute_id,omitempty" bson:"tracking_speed_attribute_id"`
	}

	DogmaEffectModifier struct {
		Domain               string `json:"domain,omitempty" bson:"domain"`
		EffectId             int32  `json:"effect_id,omitempty" bson:"effect_id"`
		Function             string `json:"func,omitempty" bson:"func"`
		ModifiedAttributeId  int32  `json:"modified_attribute_id,omitempty" bson:"modified_attribute_id"`
		ModifyingAttributeId int32  `json:"modifying_attribute_id,omitempty" bson:"modifying_attribute_id"`
		Operator             int32  `json:"operator,omitempty" bson:"operator"`
	}
)
