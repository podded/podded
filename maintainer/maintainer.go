package maintainer

import (
	"errors"
	"github.com/podded/podded/ectoplasma"
)

type Maintainer struct {
	goop *ectoplasma.PodGoo
}

func NewMaintainer(goop *ectoplasma.PodGoo) *Maintainer {
	return &Maintainer{goop: goop}
}

func (mt *Maintainer) StartMaintainer() (err error) {
	if mt == nil || mt.goop == nil {
		return errors.New("Need to init maintainer / goop struct")
	}

	//ctx := context.Background()

	mt.orphanScrape()

	return nil
}
