package rpc

import (
	"github.com/isaiahwong/cz4013/store"
)

type FlightSearch func(q string, f *Flight) bool

// FlightRepo is a repository for flights
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

// ReservationRepo is a repository for reservations
type ReservationRepo struct {
	db       *store.DB
	Relation string
}

func (r *ReservationRepo) GetAll() ([]*Flight, error) {
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

func (r *ReservationRepo) Insert(rs *ReserveFlight) error {
	return r.db.Insert(r.Relation, rs)
}

func (r *ReservationRepo) FindByID(id string) (*ReserveFlight, error) {
	rel, err := r.db.GetRelation(r.Relation)
	if err != nil {
		return nil, err
	}

	fr, ok := rel.Tuples.([]*ReserveFlight)
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

func (r *ReservationRepo) Update(f *ReserveFlight) error {
	predicate := func(old interface{}, new interface{}) bool {
		or, ok := old.(*ReserveFlight)
		if !ok {
			return false
		}

		nr, ok := new.(*ReserveFlight)
		if !ok {
			return false
		}

		return or.ID == nr.ID
	}

	return r.db.Update(r.Relation, f, predicate)
}

func (r *ReservationRepo) Delete(f *ReserveFlight) error {
	predicate := func(current interface{}) bool {
		c, ok := current.(*ReserveFlight)
		if !ok {
			return false
		}

		return c.ID == f.ID
	}

	return r.db.Delete(r.Relation, predicate)
}

func NewReservationRepo(db *store.DB) *ReservationRepo {
	return &ReservationRepo{
		db:       db,
		Relation: "reservations",
	}
}
