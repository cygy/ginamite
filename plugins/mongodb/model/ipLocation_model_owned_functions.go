package model

// IsEmpty : returns true if this location is empty.
func (location *OwnIPLocation) IsEmpty() bool {
	return len(location.Region) == 0 && len(location.Country.Name) == 0 && len(location.Country.Name) == 0
}
