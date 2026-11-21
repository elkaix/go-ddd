// Package order illustrates a single-owner domain aggregate, not a payment or
// persistence system. Callers must serialize access to each Order instance.
package order

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
)

var (
	ErrInvalid   = errors.New("invalid domain value")
	ErrCurrency  = errors.New("currency mismatch")
	ErrOverflow  = errors.New("amount overflow")
	ErrState     = errors.New("invalid order transition")
	ErrDuplicate = errors.New("duplicate SKU")
	ErrLimit     = errors.New("order limit exceeded")
)

// These are illustrative business limits; derive real limits from the domain.
const MaxLines = 50
const MaxQuantity int64 = 100

// Money is a nonnegative price in minor units. This example supports only
// USD/EUR (cents) and JPY (yen); it does not perform conversion or rounding.
type Money struct {
	minor    int64
	currency string
}

func NewMoney(minor int64, currency string) (Money, error) {
	m := Money{minor: minor, currency: currency}
	if err := m.validate(); err != nil {
		return Money{}, err
	}
	return m, nil
}

func (m Money) Minor() int64 {
	return m.minor
}

func (m Money) Currency() string {
	return m.currency
}

func (m Money) validate() error {
	if m.minor < 0 {
		return fmt.Errorf("negative price: %w", ErrInvalid)
	}
	switch m.currency {
	case "USD", "EUR", "JPY":
		return nil
	default:
		return fmt.Errorf("unsupported currency: %w", ErrInvalid)
	}
}

func (m Money) Add(other Money) (Money, error) {
	if err := m.validate(); err != nil {
		return Money{}, err
	}
	if err := other.validate(); err != nil {
		return Money{}, err
	}
	if m.currency != other.currency {
		return Money{}, ErrCurrency
	}
	if other.minor > math.MaxInt64-m.minor {
		return Money{}, ErrOverflow
	}
	return Money{minor: m.minor + other.minor, currency: m.currency}, nil
}

type Line struct {
	sku       string
	quantity  int64
	unitPrice Money
}

func NewLine(sku string, quantity int64, unitPrice Money) (Line, error) {
	l := Line{sku: sku, quantity: quantity, unitPrice: unitPrice}
	if _, err := l.Total(); err != nil {
		return Line{}, err
	}
	return l, nil
}

func (l Line) SKU() string {
	return l.sku
}

func (l Line) Quantity() int64 {
	return l.quantity
}

func (l Line) UnitPrice() Money {
	return l.unitPrice
}

func (l Line) Total() (Money, error) {
	if strings.TrimSpace(l.sku) == "" || l.quantity < 1 || l.quantity > MaxQuantity {
		return Money{}, ErrInvalid
	}
	if err := l.unitPrice.validate(); err != nil {
		return Money{}, err
	}
	if l.unitPrice.minor > math.MaxInt64/l.quantity {
		return Money{}, ErrOverflow
	}
	return Money{minor: l.unitPrice.minor * l.quantity, currency: l.unitPrice.currency}, nil
}

type Status uint8

const (
	Draft Status = iota + 1 // The zero value is intentionally invalid.
	Placed
	Cancelled
)

// An Order owns its lines. A placed or cancelled order is immutable in this
// small model; refunds, fulfillment and provider state require separate models.
type Order struct {
	id     string
	status Status
	lines  []Line
	total  Money
}

func NewOrder(id, currency string, lines []Line) (*Order, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalid
	}
	total, err := NewMoney(0, currency)
	if err != nil {
		return nil, err
	}
	o := &Order{id: id, status: Draft, total: total}
	for _, line := range lines {
		if err := o.AddLine(line); err != nil {
			return nil, err
		}
	}
	return o, nil
}

func (o *Order) ID() string     { return o.id }
func (o *Order) Status() Status { return o.status }
func (o *Order) Total() Money   { return o.total }
func (o *Order) Lines() []Line  { return slices.Clone(o.lines) }

func (o *Order) requireDraft() error {
	if o == nil || o.id == "" || o.total.validate() != nil || o.status == 0 {
		return ErrInvalid
	}
	if o.status != Draft {
		return ErrState
	}
	return nil
}

func (o *Order) AddLine(line Line) error {
	if err := o.requireDraft(); err != nil {
		return err
	}
	if len(o.lines) >= MaxLines {
		return ErrLimit
	}
	for _, existing := range o.lines {
		if existing.sku == line.sku {
			return ErrDuplicate
		}
	}
	lineTotal, err := line.Total()
	if err != nil {
		return err
	}
	total, err := o.total.Add(lineTotal)
	if err != nil {
		return err
	}
	// Copy on mutation so a shallow Order copy cannot share writable lines.
	// Publish state only after every check has succeeded.
	o.lines = append(slices.Clone(o.lines), line)
	o.total = total
	return nil
}

func (o *Order) Place() error {
	if err := o.requireDraft(); err != nil {
		return err
	}
	if len(o.lines) == 0 {
		return ErrInvalid
	}
	o.status = Placed
	return nil
}

func (o *Order) Cancel() error {
	if err := o.requireDraft(); err != nil {
		return err
	}
	o.status = Cancelled
	return nil
}
