// Package catalog provides a simple in-memory product catalog.
package catalog

import (
	"slices"

	"github.com/vukieuhaihoa/my-go-coffeeshop/internal/product/entity"
)

// Catalog represents the product catalog, containing a list of item types and a map for quick lookup by name.
type Catalog struct {
	items    []entity.ItemType
	itemsMap map[string]entity.ItemType
}

// New creates a new Catalog instance with predefined item types and initializes a map for quick lookup by name.
// Returns:
//   - A pointer to a Catalog instance.
func New() *Catalog {
	itemTypes := []entity.ItemType{
		{
			Name:  "CAPPUCCINO",
			Type:  0,
			Price: 4.5,
			Image: "img/CAPPUCCINO.png",
		},
		{
			Name:  "COFFEE_BLACK",
			Type:  1,
			Price: 3,
			Image: "img/COFFEE_BLACK.png",
		},
		{
			Name:  "COFFEE_WITH_ROOM",
			Type:  2,
			Price: 3,
			Image: "img/COFFEE_WITH_ROOM.png",
		},
		{
			Name:  "ESPRESSO",
			Type:  3,
			Price: 3.5,
			Image: "img/ESPRESSO.png",
		},
		{
			Name:  "ESPRESSO_DOUBLE",
			Type:  4,
			Price: 4.5,
			Image: "img/ESPRESSO_DOUBLE.png",
		},
		{
			Name:  "LATTE",
			Type:  5,
			Price: 4.5,
			Image: "img/LATTE.png",
		},
		{
			Name:  "CAKEPOP",
			Type:  6,
			Price: 2.5,
			Image: "img/CAKEPOP.png",
		},
		{
			Name:  "CROISSANT",
			Type:  7,
			Price: 3.25,
			Image: "img/CROISSANT.png",
		},
		{
			Name:  "MUFFIN",
			Type:  8,
			Price: 3,
			Image: "img/MUFFIN.png",
		},
		{
			Name:  "CROISSANT_CHOCOLATE",
			Type:  9,
			Price: 3.5,
			Image: "img/CROISSANT_CHOCOLATE.png",
		},
	}
	itemsMap := make(map[string]entity.ItemType)
	for _, item := range itemTypes {
		itemsMap[item.Name] = item
	}
	return &Catalog{
		items:    itemTypes,
		itemsMap: itemsMap,
	}
}

// All returns a copy of all item types in the catalog.
// Returns:
//   - A slice of ItemType containing all item types in the catalog.
func (c *Catalog) All() []entity.ItemType {
	return slices.Clone(c.items)
}

// GetByNames retrieves item types from the catalog based on their names.
// Parameters:
//   - names: A slice of strings representing the names of the item types to retrieve.
//
// Returns:
//   - A slice of ItemType containing the found item types.
//   - A slice of strings containing the names of item types that were not found in the catalog.
func (c *Catalog) GetByNames(names []string) (found []entity.ItemType, unknown []string) {
	for _, n := range names {
		if item, ok := c.itemsMap[n]; ok {
			found = append(found, item)
		} else {
			unknown = append(unknown, n)
		}
	}
	return found, unknown
}
