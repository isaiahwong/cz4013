package store

import (
	"errors"
	"reflect"
	"sync"
)

type Relation struct {
	T      reflect.Type
	Tuples interface{}
}

type DB struct {
	relationMux sync.Mutex
	Relation    map[string]*Relation
}

var (
	ErrRelationExists   = errors.New("Relation exists")
	ErrRelationNotFound = errors.New("Relation not found")
	ErrTupleRelation    = errors.New("Tuple is different from relation")
)

// CreateRelation creates a new relation in DB
func (db *DB) CreateRelation(relation string, t reflect.Type) error {
	db.relationMux.Lock()
	defer db.relationMux.Unlock()

	if _, ok := db.Relation[relation]; ok {
		return ErrRelationExists
	}
	r := new(Relation)
	r.T = t
	r.Tuples = reflect.MakeSlice(reflect.SliceOf(t), 0, 0)

	db.Relation[relation] = r
	return nil
}

func (db *DB) Insert(relation string, tuple interface{}) error {
	r, ok := db.Relation[relation]
	if !ok {
		return ErrRelationNotFound
	}
	if reflect.TypeOf(tuple) != r.T {
		return ErrTupleRelation
	}
	e := reflect.Indirect(reflect.ValueOf(tuple))

	reflect.ValueOf(r.Tuples).Elem().Set(
		reflect.Append(
			reflect.ValueOf(r.Tuples).Elem(), e.Elem(),
		),
	)
	return nil
}

func (db *DB) Update(relation string, tuple interface{}, predicate func(old interface{}, new interface{}) bool) error {
	db.relationMux.Lock()
	defer db.relationMux.Unlock()

	r, ok := db.Relation[relation]
	if !ok {
		return ErrRelationNotFound
	}
	if reflect.TypeOf(tuple) != r.T {
		return ErrTupleRelation
	}
	e := reflect.ValueOf(tuple)
	for i := 0; i < reflect.ValueOf(r.Tuples).Len(); i++ {
		if predicate(reflect.ValueOf(r.Tuples).Index(i).Interface(), tuple) {
			reflect.ValueOf(r.Tuples).Index(i).Set(e)
			break
		}
	}
	return nil
}

func (db *DB) BulkInsert(relation string, tuples interface{}) error {
	db.relationMux.Lock()
	defer db.relationMux.Unlock()

	r, ok := db.Relation[relation]
	if !ok {
		return ErrRelationNotFound
	}
	list := reflect.ValueOf(tuples)
	if list.Type().Elem() != r.T {
		return ErrTupleRelation
	}
	r.Tuples = tuples
	return nil
}

func (db *DB) GetRelation(relation string) (*Relation, error) {
	db.relationMux.Lock()
	defer db.relationMux.Unlock()

	r, ok := db.Relation[relation]
	if !ok {
		return nil, ErrRelationNotFound
	}
	return r, nil
}

func New() *DB {
	return &DB{
		Relation: make(map[string]*Relation),
	}
}
