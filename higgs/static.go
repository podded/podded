package higgs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/podded/bouncer"
	"github.com/podded/podded/ectoplasma"
	"github.com/podded/podded/types"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"time"
	"unsafe"
)

const (
	Descriptor = "HIGGS"
)

type (
	Higgs struct {
		goop *ectoplasma.PodGoo
		client *http.Client

		sqlUser string
		sqlPass string
		sqlHost string
	}
)

func NewHiggs(goop *ectoplasma.PodGoo, user string, pass string, host string) *Higgs {
	return &Higgs{
		goop: goop,
		client: &http.Client{
			// Some people have crap internet. It doesnt hurt for this to be long running
			Timeout: 5 * time.Minute,
		},
		sqlUser: user,
		sqlPass: pass,
		sqlHost: host,
	}
}

func (h *Higgs) DeleteStaticData() error {
	ctx := context.TODO()
	log.Println("Deleting Universe")
	if err := h.deleteUniverse(ctx); err != nil {
		return err
	}
	log.Println("Deleting Types")
	if err := h.deleteTypes(ctx); err != nil {
		return err
	}
	log.Println("Deleting Effects")
	if err := h.deleteDogmaEffects(ctx); err != nil {
		return err
	}
	log.Println("Deleting Attributes")
	if err := h.deleteDogmaAttributes(ctx); err != nil {
		return err
	}
	log.Println("Deleting Inventory")
	if err := h.deleteInventory(ctx); err != nil {
		return err
	}
	return nil
}

func (h *Higgs) deleteUniverse(ctx context.Context) error {
	return h.goop.MongoClient.Database("podded").Collection("universe").Drop(ctx)
}

func (h *Higgs) deleteTypes(ctx context.Context) error {
	return h.goop.MongoClient.Database("podded").Collection("types").Drop(ctx)
}

func (h *Higgs) deleteDogmaEffects(ctx context.Context) error {
	return h.goop.MongoClient.Database("podded").Collection("dogma_effects").Drop(ctx)
}

func (h *Higgs) deleteDogmaAttributes(ctx context.Context) error {
	return h.goop.MongoClient.Database("podded").Collection("dogma_attributes").Drop(ctx)
}

func (h *Higgs) deleteInventory(ctx context.Context) error {
	return h.goop.MongoClient.Database("podded").Collection("inventory").Drop(ctx)
}

//TODO FIX ALL CONTEXT CHECKING AND PORPOGATION!!!!!
func (h *Higgs) PopulateStaticData() error {

	ctx := context.TODO()

	if err := h.populateDogmaAttributes(ctx); err != nil {
		return err
	}

	//if err := h.populateDogmaEffects(ctx); err != nil {
	//	return err
	//}

	if err := h.populateUniverse(ctx); err != nil {
		return err
	}

	//if err := h.populateInventory(ctx); err != nil {
	//	return err
	//}

	return nil
}

func (h *Higgs) populateDogmaAttributes(ctx context.Context) error {
	const urlAttList = "https://esi.evetech.net/v1/dogma/attributes/?datasource=tranquility"
	const urlAttSpec = "https://esi.evetech.net/v1/dogma/attributes/%v/?datasource=tranquility"

	req := bouncer.Request{
		URL:        urlAttList,
		Method:     "GET",
		Descriptor: Descriptor,
	}
	res, status, err := h.goop.BouncerClient.MakeRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed to get attribute list")
	}
	if status != 200 {
		return errors.New("non 200 getting attribute list")
	}

	var atts []int
	err = json.Unmarshal(res.Body, &atts)
	if err != nil {
		return errors.Wrap(err, "failed to decode attribute list")
	}

	const routines = 50

	var batches [][]int
	batchSize := (len(atts) / routines)

	for batchSize < len(atts) {
		atts, batches = atts[batchSize:], append(batches, atts[0:batchSize:batchSize])
	}
	batches = append(batches, atts)

	g, c := errgroup.WithContext(ctx)

	for _, b := range batches {
		batch := b
		g.Go(func() error {

			batts := make([]interface{}, len(batch))

			for i, r := range batch {
				if c.Err() != nil {
					return c.Err()
				}
				log.Printf("ATTRIB: %v\n", r)
				url := fmt.Sprintf(urlAttSpec, r)
				br := bouncer.Request{
					URL:        url,
					Method:     "GET",
					Descriptor: Descriptor,
				}
				res, status, err := h.goop.BouncerClient.MakeRequest(br)
				if err != nil {
					return errors.Wrapf(err, "failed to fetch attribute %v", r)
				}
				if status != 200 {
					return fmt.Errorf("failed to fetch attribute %v", r)
				}

				var att types.DogmaAttribute
				err = json.Unmarshal(res.Body, &att)
				if err != nil {
					return errors.Wrapf(err, "failed to unmarshal attribute %v", r)
				}
				batts[i] = att
			}

			_, err := h.goop.MongoClient.Database("podded").Collection("dogma_attributes").InsertMany(c, batts)
			return err
		})
	}

	return g.Wait()

}

func (h *Higgs) populateDogmaEffects(ctx context.Context) error {
	const urlEffList = "https://esi.evetech.net/v1/dogma/effects/?datasource=tranquility"
	const urlEffSpec = "https://esi.evetech.net/v1/dogma/effects/%v/?datasource=tranquility"

	req := bouncer.Request{
		URL:        urlEffList,
		Method:     "GET",
		Descriptor: Descriptor,
	}
	res, status, err := h.goop.BouncerClient.MakeRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed to get effects list")
	}
	if status != 200 {
		return errors.New("non 200 getting effects list")
	}

	var effs []int
	err = json.Unmarshal(res.Body, &effs)
	if err != nil {
		return errors.Wrap(err, "failed to decode effect list")
	}

	const routines = 50

	var batches [][]int
	batchSize := (len(effs) / routines)

	for batchSize < len(effs) {
		effs, batches = effs[batchSize:], append(batches, effs[0:batchSize:batchSize])
	}
	batches = append(batches, effs)

	g, c := errgroup.WithContext(ctx)

	for _, b := range batches {
		batch := b
		g.Go(func() error {

			beffs := make([]interface{}, len(batch))

			for i, r := range batch {
				if c.Err() != nil {
					return c.Err()
				}
				log.Printf("EFFECT: %v\n", r)
				url := fmt.Sprintf(urlEffSpec, r)
				br := bouncer.Request{
					URL:        url,
					Method:     "GET",
					Descriptor: Descriptor,
				}
				res, status, err := h.goop.BouncerClient.MakeRequest(br)
				if err != nil {
					return errors.Wrapf(err, "failed to fetch effect %v", r)
				}
				if status != 200 {
					return fmt.Errorf("failed to fetch effect %v", r)
				}

				var eff types.DogmaEffect
				err = json.Unmarshal(res.Body, &eff)
				if err != nil {
					return errors.Wrapf(err, "failed to unmarshal effect %v", r)
				}
				beffs[i] = eff
			}

			_, err := h.goop.MongoClient.Database("podded").Collection("dogma_effects").InsertMany(c, beffs)
			return err
		})
	}

	return g.Wait()

}

func (h *Higgs) populateUniverse(ctx context.Context) error {

	const urlRegList = "https://esi.evetech.net/v1/universe/regions/?datasource=tranquility"
	const urlRegSpec = "https://esi.evetech.net/v1/universe/regions/%v/?datasource=tranquility"

	req := bouncer.Request{
		URL:        urlRegList,
		Method:     "GET",
		Descriptor: Descriptor,
	}
	res, status, err := h.goop.BouncerClient.MakeRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed to get regions list")
	}
	if status != 200 {
		return errors.New("non 200 getting regions list")
	}

	var regs []int
	err = json.Unmarshal(res.Body, &regs)
	if err != nil {
		return errors.Wrap(err, "failed to decode regions list")
	}

	g, c := errgroup.WithContext(ctx)

	//bregs := make([]interface{}, len(regs))
	g.Go(func() error {
		for _, r1 := range regs {
			r := r1
			if c.Err() != nil {
				return c.Err()
			}
			log.Printf("REGION: %v\n", r)
			url := fmt.Sprintf(urlRegSpec, r)
			br := bouncer.Request{
				URL:        url,
				Method:     "GET",
				Descriptor: Descriptor,
			}
			res, status, err := h.goop.BouncerClient.MakeRequest(br)
			if err != nil {
				return errors.Wrapf(err, "failed to fetch region %v", r)
			}
			if status != 200 {
				return fmt.Errorf("failed to fetch region %v", r)
			}

			var reg ESIRegion
			err = json.Unmarshal(res.Body, &reg)
			if err != nil {
				return errors.Wrapf(err, "failed to unmarshal region %v", r)
			}

			// Now convert an ESIRegion to a types.Region
			// This includes feteching sub structured data
			region, err := h.populateEsiRegion(reg, c)

			log.Println(unsafe.Sizeof(region))

			_, err = h.goop.MongoClient.Database("podded").Collection("universe").InsertOne(c, region)
			if err != nil {
				return err
			}
		}


		return err
	})

	return g.Wait()
}

// populateEsiRegion will accept a default esi region with the list of constellation ids
// and return a region with the constellations array populated with data
func (h *Higgs) populateEsiRegion(reg ESIRegion, ctx context.Context) (region types.Region, err error) {

	const urlConSpec = "https://esi.evetech.net/v1/universe/constellations/%v/?datasource=tranquility"

	region = types.Region{
		Description: reg.Description,
		Name:        reg.Name,
		RegionID:    reg.RegionID,
	}

	const routines = 4

	cons := reg.Constellations

	constellations := make([]types.Constellation, len(cons))

	var batches [][]int
	batchSize := (len(cons) / routines)

	for batchSize < len(cons) {
		cons, batches = cons[batchSize:], append(batches, cons[0:batchSize:batchSize])
	}
	batches = append(batches, cons)

	g, c := errgroup.WithContext(ctx)

	for i1, b := range batches {
		batch := b
		i := i1
		g.Go(func() error {

			bcons := make([]interface{}, len(batch))

			for j, r := range batch {
				if c.Err() != nil {
					return c.Err()
				}
				log.Printf("CONSTE: %v\n", r)
				url := fmt.Sprintf(urlConSpec, r)
				br := bouncer.Request{
					URL:        url,
					Method:     "GET",
					Descriptor: Descriptor,
				}
				res, status, err := h.goop.BouncerClient.MakeRequest(br)
				if err != nil {
					return errors.Wrapf(err, "failed to fetch constellation %v", r)
				}
				if status != 200 {
					return fmt.Errorf("failed to fetch constellation %v", r)
				}

				var con ESIConstellation
				err = json.Unmarshal(res.Body, &con)
				if err != nil {
					return errors.Wrapf(err, "failed to unmarshal constellation %v", r)
				}

				// Now convert an ESIConstellation to a types.Constellation
				// This includes feteching sub structured data
				constellation, err := h.populateEsiConstellation(con, routines, c)
				if err != nil {
					return errors.Wrap(err, "failed to populate constellation")
				}
				constellations[(i*(batchSize-1))+j] = constellation

			}

			_, err := h.goop.MongoClient.Database("podded").Collection("universe").InsertMany(c, bcons)
			return err
		})
	}

	err = g.Wait()

	region.Constellations = constellations

	return region, err
}

// populateEsiConstellation will accept a default ESIConstellation with the list of system ids
// and return a Constellation with the systems array populated with data
func (h *Higgs) populateEsiConstellation(con ESIConstellation, routines int, ctx context.Context) (constellation types.Constellation, err error) {

	const urlSysSpec = "https://esi.evetech.net/v3/universe/systems/%v/?datasource=tranquility"

	constellation = types.Constellation{
		ConstellationID: con.ConstellationID,
		Name:            con.Name,
		Systems:         nil,
		Position: types.Position{
			X: con.Position.X,
			Y: con.Position.Y,
			Z: con.Position.Z,
		},
	}

	syss := con.Systems

	systems := make([]types.System, len(syss))

	g, c := errgroup.WithContext(ctx)
	for i1, r1 := range syss {
		i := i1
		r := r1
		g.Go(func() error {

			if c.Err() != nil {
				return c.Err()
			}
			log.Printf("SYSTEM: %v\n", r)
			url := fmt.Sprintf(urlSysSpec, r)
			br := bouncer.Request{
				URL:        url,
				Method:     "GET",
				Descriptor: Descriptor,
			}
			res, status, err := h.goop.BouncerClient.MakeRequest(br)
			if err != nil {
				return errors.Wrapf(err, "failed to fetch system %v", r)
			}
			if status != 200 {
				return fmt.Errorf("failed to fetch system %v", r)
			}

			var sys ESISystem
			err = json.Unmarshal(res.Body, &sys)
			if err != nil {
				return errors.Wrapf(err, "failed to unmarshal system %v", r)
			}

			// Now convert an ESIConstellation to a types.Constellation
			// This includes feteching sub structured data
			constellation, err := h.populateEsiSystem(sys, routines, c)
			if err != nil {
				return err
			}
			systems[i] = constellation
			return nil
		})
	}

	err = g.Wait()

	constellation.Systems = systems

	return constellation, err
}

func (h *Higgs) populateEsiSystem(sys ESISystem, routines int, ctx context.Context) (system types.System, err error) {

	system = types.System{
		SystemID: sys.SystemID,
		Name:     sys.Name,
		Planets:  nil,
		Position: types.Position{
			X: sys.Position.X,
			Y: sys.Position.Y,
			Z: sys.Position.Z,
		},
		SecurityClass:  sys.SecurityClass,
		SecurityStatus: sys.SecurityStatus,
		Star:           types.Star{},
		Stargates:      nil,
		Stations:       nil,
	}

	star, err := h.populateStar(sys.StarID, ctx)
	if err != nil && err != noStar {
		return system, err
	}
	if err == noStar {
		log.Printf("DEBUG: -1 StarID for System: %v", sys.SystemID)
	}
	system.Star = star

	// Context check
	if ctx.Err() != nil {
		return system, ctx.Err()
	}

	gates, err := h.populateStargates(sys.Stargates, ctx)
	if err != nil {
		return system, err
	}
	system.Stargates = gates

	// Context check
	if ctx.Err() != nil {
		return system, ctx.Err()
	}

	stations, err := h.populateStations(sys.Stations, ctx)
	if err != nil {
		return system, err
	}
	system.Stations = stations

	planets, err := h.populatePlanets(sys.Planets, routines, ctx)
	if err != nil {
		return system, err
	}
	system.Planets = planets

	return system, nil
}

var noStar = errors.New("No star with ID -1")

func (h *Higgs) populateStar(id int, ctx context.Context) (star types.Star, err error) {

	const urlStarSpec = "https://esi.evetech.net/v1/universe/stars/%v/?datasource=tranquility"

	if id == -1 {
		return types.Star{}, noStar
	}

	url := fmt.Sprintf(urlStarSpec, id)
	br := bouncer.Request{
		URL:        url,
		Method:     "GET",
		Descriptor: Descriptor,
	}
	res, status, err := h.goop.BouncerClient.MakeRequest(br)
	if err != nil {
		return star, errors.Wrapf(err, "failed to fetch star %v", id)
	}
	if status != 200 {
		return star, fmt.Errorf("failed to fetch star %v", id)
	}

	err = json.Unmarshal(res.Body, &star)
	if err != nil {
		return star, errors.Wrapf(err, "failed to unmarshal star %v", id)
	}

	return star, nil
}

func (h *Higgs) populateStargates(stargates []int, ctx context.Context) (gates []types.Stargate, err error) {

	const urlSysSpec = "https://esi.evetech.net/v1/universe/stargates/%v/?datasource=tranquility"

	gates = make([]types.Stargate, len(stargates))

	for i, gt := range stargates {

		// Context check
		if ctx.Err() != nil {
			return gates, ctx.Err()
		}

		url := fmt.Sprintf(urlSysSpec, gt)
		br := bouncer.Request{
			URL:        url,
			Method:     "GET",
			Descriptor: Descriptor,
		}
		res, status, err := h.goop.BouncerClient.MakeRequest(br)
		if err != nil {
			return gates, errors.Wrapf(err, "failed to fetch gate %v", gt)
		}
		if status != 200 {
			return gates, fmt.Errorf("failed to fetch gate %v", gt)
		}

		var gate types.Stargate
		err = json.Unmarshal(res.Body, &gate)
		if err != nil {
			return gates, errors.Wrapf(err, "failed to unmarshal gate %v", gt)
		}

		gates[i] = gate
	}

	return gates, nil
}

func (h *Higgs) populateStations(sta []int, ctx context.Context) (stations []types.Station, err error) {
	const urlSysSpec = "https://esi.evetech.net/v1/universe/stations/%v/?datasource=tranquility"

	stations = make([]types.Station, len(sta))

	for i, s := range sta {

		// Context check
		if ctx.Err() != nil {
			return stations, ctx.Err()
		}

		url := fmt.Sprintf(urlSysSpec, s)
		br := bouncer.Request{
			URL:        url,
			Method:     "GET",
			Descriptor: Descriptor,
		}
		res, status, err := h.goop.BouncerClient.MakeRequest(br)
		if err != nil {
			return stations, errors.Wrapf(err, "failed to fetch station %v", s)
		}
		if status != 200 {
			return stations, fmt.Errorf("failed to fetch station %v", s)
		}

		var station types.Station
		err = json.Unmarshal(res.Body, &station)
		if err != nil {
			return stations, errors.Wrapf(err, "failed to unmarshal station %v", s)
		}

		stations[i] = station
	}

	return stations, nil
}

func (h *Higgs) populatePlanets(plans []ESISystemPlanets, routines int, ctx context.Context) (planets []types.Planet, err error) {

	const urlPlanSpec = "https://esi.evetech.net/v1/universe/planets/%v/?datasource=tranquility"

	planets = make([]types.Planet, len(plans))

	g, c := errgroup.WithContext(ctx)

	for i1, p1 := range plans {
		i := i1
		p := p1
		g.Go(func() error {
			if c.Err() != nil {
				return c.Err()
			}

			url := fmt.Sprintf(urlPlanSpec, p.PlanetID)
			br := bouncer.Request{
				URL:        url,
				Method:     "GET",
				Descriptor: Descriptor,
			}
			res, status, err := h.goop.BouncerClient.MakeRequest(br)
			if err != nil {
				return errors.Wrapf(err, "failed to fetch planet %v", p.PlanetID)
			}
			if status != 200 {
				return fmt.Errorf("failed to fetch planet %v", p.PlanetID)
			}

			var plan ESIPlanet
			err = json.Unmarshal(res.Body, &plan)
			if err != nil {
				return errors.Wrapf(err, "failed to unmarshal planet %v", p.PlanetID)
			}

			// Now convert an ESIPlanet to a types.Planet
			// This includes fetching sub structured data
			planet := types.Planet{
				Name:          plan.Name,
				PlanetID:      plan.PlanetID,
				Position:      types.Position{},
				TypeID:        plan.TypeID,
				Moons:         nil,
				AsteroidBelts: nil,
			}

			moons, err := h.populateMoons(p.Moons, c)
			if err != nil {
				return err
			}
			planet.Moons = moons

			roids, err := h.populateBelts(p.AsteroidBelts, c)
			if err != nil {
				return err
			}
			planet.AsteroidBelts = roids

			planets[i] = planet

			return err
		})
	}

	err = g.Wait()

	return planets, err
}

func (h *Higgs) populateMoons(moonies []int, ctx context.Context) (moons []types.Moon, err error) {

	const urlMoonSpec = "https://esi.evetech.net/v1/universe/moons/%v/?datasource=tranquility"

	moons = make([]types.Moon, len(moonies))

	// TODO pass this context to bouncer when bouncer supports context
	g, c := errgroup.WithContext(ctx)

	for i1, p1 := range moonies {
		i := i1
		p := p1
		g.Go(func() error {
			if c.Err() != nil {
				return c.Err()
			}

			url := fmt.Sprintf(urlMoonSpec, p)
			br := bouncer.Request{
				URL:        url,
				Method:     "GET",
				Descriptor: Descriptor,
			}
			res, status, err := h.goop.BouncerClient.MakeRequest(br)
			if err != nil {
				return errors.Wrapf(err, "failed to fetch moon %v", p)
			}
			if status != 200 {
				return fmt.Errorf("failed to fetch moon %v", p)
			}

			var mn types.Moon
			err = json.Unmarshal(res.Body, &mn)
			if err != nil {
				return errors.Wrapf(err, "failed to unmarshal moon %v", p)
			}

			moons[i] = mn

			return err
		})
	}

	err = g.Wait()

	return moons, err
}

func (h *Higgs) populateBelts(roids []int, ctx context.Context) (belts []types.AsteroidBelt, err error) {
	const urlBeltSpec = "https://esi.evetech.net/v1/universe/asteroid_belts/%v/?datasource=tranquility"

	belts = make([]types.AsteroidBelt, len(roids))

	// TODO pass this context to bouncer when bouncer supports context
	g, c := errgroup.WithContext(ctx)

	for i1, p1 := range roids {
		i := i1
		p := p1
		g.Go(func() error {
			if c.Err() != nil {
				return c.Err()
			}
			url := fmt.Sprintf(urlBeltSpec, p)
			br := bouncer.Request{
				URL:        url,
				Method:     "GET",
				Descriptor: Descriptor,
			}
			res, status, err := h.goop.BouncerClient.MakeRequest(br)
			if err != nil {
				return errors.Wrapf(err, "failed to fetch moon %v", p)
			}
			if status != 200 {
				return fmt.Errorf("failed to fetch moon %v", p)
			}

			var bt types.AsteroidBelt
			err = json.Unmarshal(res.Body, &bt)
			if err != nil {
				return errors.Wrapf(err, "failed to unmarshal moon %v", p)
			}

			belts[i] = bt

			return err
		})
	}

	err = g.Wait()

	return belts, err
}

func (h *Higgs) populateInventory(ctx context.Context) error {

	const urlCatList = "https://esi.evetech.net/v1/universe/categories/?datasource=tranquility"
	const urlCatSpec = "https://esi.evetech.net/v1/universe/categories/%v/?datasource=tranquility"

	req := bouncer.Request{
		URL:        urlCatList,
		Method:     "GET",
		Descriptor: Descriptor,
	}
	res, status, err := h.goop.BouncerClient.MakeRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed to get regions list")
	}
	if status != 200 {
		return errors.New("non 200 getting regions list")
	}

	var cats []int
	err = json.Unmarshal(res.Body, &cats)
	if err != nil {
		return errors.Wrap(err, "failed to decode regions list")
	}

	g, c := errgroup.WithContext(ctx)

	bcats := make([]interface{}, len(cats))

	for i2, r := range cats {
		ct := r
		i := i2
		g.Go(func() error {

			if c.Err() != nil {
				return c.Err()
			}
			//log.Printf("CATEGO: %v\n", ct)
			url := fmt.Sprintf(urlCatSpec, ct)
			br := bouncer.Request{
				URL:        url,
				Method:     "GET",
				Descriptor: Descriptor,
			}
			res, status, err := h.goop.BouncerClient.MakeRequest(br)
			if err != nil {
				return errors.Wrapf(err, "failed to fetch region %v", ct)
			}
			if status != 200 {
				return fmt.Errorf("failed to fetch region %v", ct)
			}

			var cat ESICategory
			err = json.Unmarshal(res.Body, &cat)
			if err != nil {
				return errors.Wrapf(err, "failed to unmarshal region %v", ct)
			}
			// Now convert an ESICategory to a types.Category
			// This includes fetching sub structured data
			region, err := h.populateEsiCategory(cat, c)
			bcats[i] = region
			return nil
		})

	}

	err = g.Wait()
	if err != nil {
		return err
	}
	_, err = h.goop.MongoClient.Database("podded").Collection("inventory").InsertMany(ctx, bcats)
	return errors.Wrap(err, "failed to insert into inventory")
}

func (h *Higgs) populateEsiCategory(cat ESICategory, ctx context.Context) (category types.Category, err error) {
	const urlGrpSpec = "https://esi.evetech.net/v1/universe/groups/%v/?datasource=tranquility"

	category = types.Category{
		CategoryID: cat.CategoryID,
		Name:       cat.Name,
		Published:  cat.Published,
	}

	grps := cat.Groups

	groups := make([]types.Group, len(grps))

	g, c := errgroup.WithContext(ctx)

	for j1, r1 := range grps {
		j := j1
		r := r1
		g.Go(func() error {
			if c.Err() != nil {
				return c.Err()
			}
			//log.Printf("GROUP : %v\n", r)
			url := fmt.Sprintf(urlGrpSpec, r)
			br := bouncer.Request{
				URL:        url,
				Method:     "GET",
				Descriptor: Descriptor,
			}
			res, status, err := h.goop.BouncerClient.MakeRequest(br)
			if err != nil {
				return errors.Wrapf(err, "failed to fetch group %v", r)
			}
			if status != 200 {
				return fmt.Errorf("failed to fetch group %v", r)
			}

			var gp ESIGroup
			err = json.Unmarshal(res.Body, &gp)
			if err != nil {
				return errors.Wrapf(err, "failed to unmarshal group %v", r)
			}

			// Now convert an ESIConstellation to a types.Constellation
			// This includes feteching sub structured data
			group, err := h.populateEsiGroup(gp, c)
			if err != nil {
				return errors.Wrap(err, "failed to populate group")
			}
			groups[j] = group

			return nil
		})
	}
	err = g.Wait()
	category.Groups = groups
	return category, err
}

func (h *Higgs) populateEsiGroup(gp ESIGroup, ctx context.Context) (group types.Group, err error) {

	const urlTypSpec = "https://esi.evetech.net/v3/universe/types/%v/?datasource=tranquility"

	group = types.Group{
		GroupID:   gp.GroupID,
		Name:      gp.Name,
		Published: gp.Published,
	}

	tps := gp.Types

	typs := make([]types.Type, len(tps))

	g, c := errgroup.WithContext(ctx)

	for i1, r1 := range tps {
		r := r1
		i := i1
		g.Go(func() error {
			if c.Err() != nil {
				return c.Err()
			}
			//log.Printf("TYPE  : %v\n", r)
			url := fmt.Sprintf(urlTypSpec, r)
			br := bouncer.Request{
				URL:        url,
				Method:     "GET",
				Descriptor: Descriptor,
			}
			res, status, err := h.goop.BouncerClient.MakeRequest(br)
			if err != nil {
				return errors.Wrapf(err, "failed to fetch type %v", r)
			}
			if status != 200 {
				return fmt.Errorf("failed to fetch type %v", r)
			}

			var tp types.Type
			err = json.Unmarshal(res.Body, &tp)
			if err != nil {
				return errors.Wrapf(err, "failed to unmarshal type %v", r)
			}
			typs[i] = tp
			return nil
		})
	}

	err = g.Wait()

	group.Types = typs

	return group, err
}
