package catalog

import "time"

type Product struct {
	ID                 int64     `json:"id"`
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	Category           string    `json:"category"`
	Price              float64   `json:"price"`
	DiscountPercentage float64   `json:"discountPercentage"`
	Rating             float64   `json:"rating"`
	Stock              int       `json:"stock"`
	Brand              string    `json:"brand"`
	Thumbnail          string    `json:"thumbnail"`
	Images             []string  `json:"images"`
	Source             string    `json:"source"`
	SourceID           string    `json:"sourceId"`
	SourceURL          string    `json:"sourceUrl"`
	ImportedAt         time.Time `json:"importedAt"`
	Variants           []Variant `json:"variants,omitempty"`
}

type Variant struct {
	SKU        string         `json:"sku"`
	Title      string         `json:"title"`
	Price      float64        `json:"price"`
	Stock      int            `json:"stock"`
	Attributes map[string]any `json:"attributes"`
	Source     string         `json:"source"`
	SourceID   string         `json:"sourceId"`
	IsActive   bool           `json:"isActive"`
}
