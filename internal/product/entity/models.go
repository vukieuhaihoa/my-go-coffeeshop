// Package entity contains the data models for the product catalog.
package entity

// ItemType represents a type of item in the product catalog.
type ItemType struct {
	Name  string  `json:"name"`
	Type  int     `json:"type"`
	Price float64 `json:"price"`
	Image string  `json:"image"`
}

// Item represents an individual item in the product catalog.
type Item struct {
	Type  int     `json:"type"`
	Price float64 `json:"price"`
}
