package dao

import (
	"github.com/stretchr/objx"
	"sync"
)

type StateDao struct {
	state objx.Map
	mutex sync.RWMutex
}

func NewStateDao() StateDao {
	s, _ := objx.FromJSON(`{}`)
	return StateDao{state: s}
}

func (s *StateDao) GetBool(key string) bool {
	s.mutex.RLock()
	result := s.state.Get(key).Bool()
	s.mutex.RUnlock()
	return result
}

func (s *StateDao) GetString(key string) string {
	s.mutex.RLock()
	result := s.state.Get(key).Str()
	s.mutex.RUnlock()
	return result
}

func (s *StateDao) GetInt(key string) int {
	s.mutex.RLock()
	result := s.state.Get(key).Int(0)
	s.state.Get(key).Float32()
	s.mutex.RUnlock()
	return result
}

func (s *StateDao) GetFloat64(key string) float64 {
	s.mutex.RLock()
	result := s.state.Get(key).Float64(0)

	s.mutex.RUnlock()
	return result
}

func (s *StateDao) Set(key string, value interface{}) {
	s.mutex.Lock()
	s.state.Set(key, value)

	s.mutex.Unlock()
}

func (s *StateDao) GetJson() string {
	s.mutex.Lock()
	result, _ := s.state.JSON()
	s.mutex.Unlock()
	return result
}
