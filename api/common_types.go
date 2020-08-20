package api

// DescriptedValue : a struct containing a value and its localized description.
type DescriptedValue struct {
	Value       int    `json:"value"`
	Description string `json:"description"`
}

// DescriptedString : a struct containing a string and its localized description.
type DescriptedString struct {
	Value       string `json:"value"`
	Description string `json:"description"`
}
