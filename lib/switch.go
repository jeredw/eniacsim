package lib

import (
	"fmt"
	"sync"
)

type Switchboard interface {
	FindSwitch(name string) (Switch, error)
}

type Switch interface {
	Get() string
	Set(value string) error
}

type BoolSwitchSetting struct {
	Key   string
	Value bool
}

type BoolSwitch struct {
	Owner    sync.Locker
	Name     string
	Data     *bool
	Settings []BoolSwitchSetting
}

func (s *BoolSwitch) Get() string {
	if s.Owner != nil {
		s.Owner.Lock()
		defer s.Owner.Unlock()
	}
	for i := range s.Settings {
		if *s.Data == s.Settings[i].Value {
			return s.Settings[i].Key
		}
	}
	return "?"
}

func (s *BoolSwitch) Set(value string) error {
	if s.Owner != nil {
		s.Owner.Lock()
		defer s.Owner.Unlock()
	}
	for i := range s.Settings {
		if value == s.Settings[i].Key {
			*s.Data = s.Settings[i].Value
			return nil
		}
	}
	return fmt.Errorf("invalid switch %s setting %s", s.Name, value)
}

type IntSwitchSetting struct {
	Key   string
	Value int
}

type IntSwitch struct {
	Owner    sync.Locker
	Name     string
	Data     *int
	Settings []IntSwitchSetting
}

func (s *IntSwitch) Get() string {
	if s.Owner != nil {
		s.Owner.Lock()
		defer s.Owner.Unlock()
	}
	for i := range s.Settings {
		if *s.Data == s.Settings[i].Value {
			return s.Settings[i].Key
		}
	}
	return "?"
}

func (s *IntSwitch) Set(value string) error {
	if s.Owner != nil {
		s.Owner.Lock()
		defer s.Owner.Unlock()
	}
	for i := range s.Settings {
		if value == s.Settings[i].Key {
			*s.Data = s.Settings[i].Value
			return nil
		}
	}
	return fmt.Errorf("invalid switch %s setting %s", s.Name, value)
}

type ByteSwitchSetting struct {
	Key   string
	Value byte
}

type ByteSwitch struct {
	Owner    sync.Locker
	Name     string
	Data     *byte
	Settings []ByteSwitchSetting
}

func (s *ByteSwitch) Get() string {
	if s.Owner != nil {
		s.Owner.Lock()
		defer s.Owner.Unlock()
	}
	for i := range s.Settings {
		if *s.Data == s.Settings[i].Value {
			return s.Settings[i].Key
		}
	}
	return "?"
}

func (s *ByteSwitch) Set(value string) error {
	if s.Owner != nil {
		s.Owner.Lock()
		defer s.Owner.Unlock()
	}
	for i := range s.Settings {
		if value == s.Settings[i].Key {
			*s.Data = s.Settings[i].Value
			return nil
		}
	}
	return fmt.Errorf("invalid switch %s setting %s", s.Name, value)
}
