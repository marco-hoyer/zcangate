package dao

import (
	"github.com/stretchr/objx"
)

type StateDao struct {
	state objx.Map
}

func NewStateDao() StateDao {
	s, _ := objx.FromJSON(`{}`)
	return StateDao{state: s}
}

func (s *StateDao) GetBool(key string) bool {
	return s.state.Get(key).Bool()
}

func (s *StateDao) GetString(key string) string {
	return s.state.Get(key).Str()
}

func (s *StateDao) GetInt(key string) int {
	return s.state.Get(key).Int(0)
}

func (s *StateDao) Set(key string, value interface{}) {
	s.state.Set(key, value)
}
