package store

import (
	"errors"
	"reflect"
)

type Relation struct {
	T      reflect.Type
	Tuples interface{}
}

type DB struct {
	Relation map[string]*Relation
}

var (
	ErrRelationExists   = errors.New("Relation exists")
	ErrRelationNotFound = errors.New("Relation not found")
	ErrTupleRelation    = errors.New("Tuple is different from relation")
)

// CreateRelation creates a new relation in DB
func (db *DB) CreateRelation(relation string, t reflect.Type) error {
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

func (db *DB) BulkInsert(relation string, tuples interface{}) error {
	r, ok := db.Relation[relation]
	if !ok {
		return ErrRelationNotFound
	}
	list := reflect.Indirect(reflect.ValueOf(tuples))
	if list.Type().Elem() != r.T {
		return ErrTupleRelation
	}
	r.Tuples = tuples

	return nil
}

func New() *DB {
	return &DB{
		Relation: make(map[string]*Relation),
	}
}
