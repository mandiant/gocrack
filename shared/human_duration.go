package shared

import "time"

// HumanDuration is YAML encodable duration for configuration objects
type HumanDuration struct {
	time.Duration
}

// MarshalYAML implements a YAML encoder to save the human duration within a YAML file
func (s HumanDuration) MarshalYAML() (interface{}, error) {
	return s.String(), nil
}

// UnmarshalYAML implements a YAML decoder to load the duration into a time.Duration
func (s *HumanDuration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var humanDuration string
	if err := unmarshal(&humanDuration); err != nil {
		return err
	}

	dur, err := time.ParseDuration(humanDuration)
	if err != nil {
		return err
	}

	s.Duration = dur
	return nil
}
