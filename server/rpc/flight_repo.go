package rpc

import (
	"github.com/isaiahwong/cz4013/store"
)

type FlightSearch func(q string, f *Flight) bool

type FlightRepo struct {
	db       *store.DB
	Relation string
}

func (r *FlightRepo) GetAll() ([]*Flight, error) {
	rel, err := r.db.GetRelation(r.Relation)
	if err != nil {
		return nil, err
	}

	fr, ok := rel.Tuples.([]*Flight)
	if !ok {
		return nil, ErrFailCast
	}

	return fr, nil
}

func (r *FlightRepo) FindByID(id int32) (*Flight, error) {
	rel, err := r.db.GetRelation(r.Relation)
	if err != nil {
		return nil, err
	}

	fr, ok := rel.Tuples.([]*Flight)
	if !ok {
		return nil, ErrFailCast
	}

	for _, f := range fr {
		if f.ID == id {
			return f, nil
		}
	}

	return nil, nil
}

func (r *FlightRepo) Update(f *Flight) error {
	predicate := func(old interface{}, new interface{}) bool {
		oldFlight, ok := old.(*Flight)
		if !ok {
			return false
		}

		newFlight, ok := new.(*Flight)
		if !ok {
			return false
		}

		return oldFlight.ID == newFlight.ID
	}

	return r.db.Update(r.Relation, f, predicate)
}

func NewFlightRepo(db *store.DB) *FlightRepo {
	return &FlightRepo{
		db:       db,
		Relation: "flights",
	}
}
