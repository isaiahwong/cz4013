package store

import (
	"errors"
	"reflect"
)

type Relation struct {
	Kind   reflect.Kind
	Tuples []interface{}
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
func (db *DB) CreateRelation(relation string, kind reflect.Kind) error {
	if _, ok := db.Relation[relation]; ok {
		return ErrRelationExists
	}
	r := new(Relation)
	r.Kind = kind
	r.Tuples = make([]interface{}, 0)

	db.Relation[relation] = r
	return nil
}

func (db *DB) Insert(relation string, tuple interface{}) error {
	r, ok := db.Relation[relation]
	if !ok {
		return ErrRelationNotFound
	}
	if reflect.TypeOf(tuple).Kind() != r.Kind {
		return ErrTupleRelation
	}
	r.Tuples = append(r.Tuples, tuple)
	return nil
}

func New() *DB {
	return &DB{
		Relation: make(map[string]*Relation),
	}
}
