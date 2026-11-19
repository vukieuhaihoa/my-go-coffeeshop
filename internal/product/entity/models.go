// Package entity contains the data models for the product catalog.
package entity

// ItemType represents a type of item in the product catalog.
type ItemType struct {
	Name  string  `json:"name"`
	Type  int32   `json:"type"`
	Price float64 `json:"price"`
	Image string  `json:"image"`
}
