package models

import "time"

type Product struct {
	ID          string    `json:"id" bson:"_id"`
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
	Price       float64   `json:"price" bson:"price"`
	Discount    float64   `json:"discount" bson:"discount"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required,min=1"`
	Description string  `json:"description" binding:"required,min=1"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Discount    float64 `json:"discount" binding:"min=0,max=100"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name" binding:"min=1"`
	Description string  `json:"description" binding:"min=1"`
	Price       float64 `json:"price" binding:"gt=0"`
	Discount    float64 `json:"discount" binding:"min=0,max=100"`
}
