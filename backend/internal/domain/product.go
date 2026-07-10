package domain

import (
	"errors"
	"strings"
	"time"
)

type Product struct {
	id          string
	name        string
	description string
	price       float64
	stock       int
	categoryID  *string
	imageURL    *string

	Timestamps
	SoftDelete
	Activatable
}

func NewProduct(name, description string, price float64, stock int, categoryID, imageURL *string) (*Product, error) {
	if err := validateProduct(name, price, stock); err != nil {
		return nil, err
	}

	product := &Product{}

	product.setData(
		name,
		description,
		price,
		stock,
		categoryID,
		imageURL,
	)

	return product, nil
}

func RestoreProduct(
	id, name, description string,
	price float64,
	stock int,
	categoryID, imageURL *string,
	active bool,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) (*Product, error) {
	if err := validateProduct(name, price, stock); err != nil {
		return nil, err
	}

	return &Product{
		id:          id,
		name:        name,
		description: description,
		price:       price,
		stock:       stock,
		categoryID:  categoryID,
		imageURL:    imageURL,
		Timestamps:  NewTimestampsFrom(createdAt, updatedAt),
		SoftDelete:  NewSoftDeleteFrom(deletedAt),
		Activatable: NewActivatableFrom(active),
	}, nil
}

func (p *Product) setData(
	name string,
	description string,
	price float64,
	stock int,
	categoryID *string,
	imageURL *string,
) {
	p.name = strings.TrimSpace(name)
	p.description = strings.TrimSpace(description)
	p.price = price
	p.stock = stock
	p.categoryID = categoryID
	p.imageURL = imageURL
}

func validateProduct(name string, price float64, stock int) error {
	var errs []error

	if strings.TrimSpace(name) == "" {
		errs = append(errs, ErrInvalidProductName)
	}
	if price <= 0 {
		errs = append(errs, ErrInvalidProductPrice)
	}
	if stock < 0 {
		errs = append(errs, ErrInvalidProductStock)
	}

	return errors.Join(errs...)
}

func (p *Product) Update(
	name string,
	description string,
	price float64,
	stock int,
	categoryID *string,
	imageURL *string,
) error {
	if err := validateProduct(name, price, stock); err != nil {
		return err
	}

	p.setData(
		name,
		description,
		price,
		stock,
		categoryID,
		imageURL,
	)

	return nil
}

func (p *Product) ID() string {
	return p.id
}

func (p *Product) Name() string {
	return p.name
}

func (p *Product) Description() string {
	return p.description
}

func (p *Product) Price() float64 {
	return p.price
}

func (p *Product) Stock() int {
	return p.stock
}

func (p *Product) CategoryID() *string {
	return p.categoryID
}

func (p *Product) ImageURL() *string {
	return p.imageURL
}

func (p *Product) HasStock(quantity int) bool {
	return quantity > 0 && p.stock >= quantity
}

func (p *Product) ReduceStock(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	if !p.HasStock(quantity) {
		return ErrInsufficientStock
	}

	p.stock -= quantity
	return nil
}

func (p *Product) RestoreStock(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	p.stock += quantity
	return nil
}
