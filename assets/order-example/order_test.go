package order_test

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	"example.com/go-ddd/order"
)

func price(t *testing.T, minor int64, currency string) order.Money {
	t.Helper()
	m, err := order.NewMoney(minor, currency)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func line(t *testing.T, sku string, quantity, minor int64, currency string) order.Line {
	t.Helper()
	l, err := order.NewLine(sku, quantity, price(t, minor, currency))
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func newOrder(t *testing.T, lines ...order.Line) *order.Order {
	t.Helper()
	o, err := order.NewOrder("order-1", "USD", lines)
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func wantError(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("error = %v; want %v", err, want)
	}
}

func TestMoneyValidationAndBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, currency string
		minor          int64
	}{
		{"negative", "USD", -1},
		{"empty currency", "", 0},
		{"unsupported", "ZZZ", 1},
		{"case sensitive", "usd", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := order.NewMoney(tc.minor, tc.currency)
			wantError(t, err, order.ErrInvalid)
		})
	}
	max := price(t, math.MaxInt64, "USD")
	zero := price(t, 0, "USD")
	got, err := max.Add(zero)
	if err != nil || got.Minor() != math.MaxInt64 {
		t.Fatalf("maximum + zero = %v, %v", got, err)
	}
	_, err = max.Add(price(t, 1, "USD"))
	wantError(t, err, order.ErrOverflow)
	_, err = max.Add(price(t, 1, "EUR"))
	wantError(t, err, order.ErrCurrency)
	_, err = (order.Money{}).Add(zero)
	wantError(t, err, order.ErrInvalid)
	_, err = zero.Add(order.Money{})
	wantError(t, err, order.ErrInvalid)
}

func TestLineValidationAndMultiplication(t *testing.T) {
	for _, tc := range []struct {
		name, sku string
		quantity  int64
		money     order.Money
		want      error
	}{
		{"blank SKU", " ", 1, price(t, 1, "USD"), order.ErrInvalid},
		{"zero quantity", "a", 0, price(t, 1, "USD"), order.ErrInvalid},
		{"negative quantity", "a", -1, price(t, 1, "USD"), order.ErrInvalid},
		{"quantity limit", "a", order.MaxQuantity + 1, price(t, 1, "USD"), order.ErrInvalid},
		{"zero money", "a", 1, order.Money{}, order.ErrInvalid},
		{"overflow", "a", 2, price(t, math.MaxInt64, "USD"), order.ErrOverflow},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := order.NewLine(tc.sku, tc.quantity, tc.money)
			wantError(t, err, tc.want)
		})
	}
	l := line(t, "a", 2, math.MaxInt64/2, "USD")
	total, err := l.Total()
	if err != nil || total.Minor() != math.MaxInt64-1 {
		t.Fatalf("near boundary total = %v, %v", total, err)
	}
	_, err = (order.Line{}).Total()
	wantError(t, err, order.ErrInvalid)
}

func TestOrderOwnsLines(t *testing.T) {
	original := line(t, "a", 2, 125, "USD")
	input := []order.Line{original}
	o, err := order.NewOrder("order-1", "USD", input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = line(t, "changed", 1, 999, "USD")
	output := o.Lines()
	output[0] = order.Line{}
	if o.Lines()[0] != original || o.Total().Minor() != 250 {
		t.Fatal("caller mutation changed aggregate state")
	}
}

func TestShallowOrderCopiesOwnSubsequentMutations(t *testing.T) {
	original := newOrder(t,
		line(t, "a", 1, 10, "USD"),
		line(t, "b", 1, 20, "USD"),
		line(t, "c", 1, 30, "USD"),
	)
	copied := *original
	if err := original.AddLine(line(t, "original", 1, 40, "USD")); err != nil {
		t.Fatal(err)
	}
	if err := copied.AddLine(line(t, "copied", 1, 50, "USD")); err != nil {
		t.Fatal(err)
	}
	if original.Lines()[3].SKU() != "original" || original.Total().Minor() != 100 {
		t.Fatal("copy mutation changed original aggregate")
	}
	if copied.Lines()[3].SKU() != "copied" || copied.Total().Minor() != 110 {
		t.Fatal("original mutation changed copied aggregate")
	}
}

func TestRejectedAddPreservesState(t *testing.T) {
	for _, tc := range []struct {
		name string
		line order.Line
		want error
	}{
		{"invalid", order.Line{}, order.ErrInvalid},
		{"duplicate", line(t, "a", 1, 10, "USD"), order.ErrDuplicate},
		{"currency", line(t, "b", 1, 10, "EUR"), order.ErrCurrency},
		{"overflow", line(t, "b", 1, math.MaxInt64, "USD"), order.ErrOverflow},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := newOrder(t, line(t, "a", 1, 10, "USD"))
			before, total, state := o.Lines(), o.Total(), o.Status()
			wantError(t, o.AddLine(tc.line), tc.want)
			if !reflect.DeepEqual(before, o.Lines()) || total != o.Total() || state != o.Status() {
				t.Fatal("rejected command changed state")
			}
		})
	}
}

func TestOrderTransitions(t *testing.T) {
	o := newOrder(t)
	wantError(t, o.Place(), order.ErrInvalid)
	if o.Status() != order.Draft {
		t.Fatal("empty order left draft state")
	}
	if err := o.AddLine(line(t, "a", 1, 100, "USD")); err != nil {
		t.Fatal(err)
	}
	if err := o.Place(); err != nil || o.Status() != order.Placed {
		t.Fatalf("place = %v; state = %v", err, o.Status())
	}
	wantError(t, o.Place(), order.ErrState)
	wantError(t, o.Cancel(), order.ErrState)
	wantError(t, o.AddLine(line(t, "b", 1, 100, "USD")), order.ErrState)
	if o.Status() != order.Placed || len(o.Lines()) != 1 || o.Total().Minor() != 100 {
		t.Fatal("terminal state changed")
	}
	cancelled := newOrder(t, line(t, "a", 1, 100, "USD"))
	if err := cancelled.Cancel(); err != nil || cancelled.Status() != order.Cancelled {
		t.Fatalf("cancel = %v; state = %v", err, cancelled.Status())
	}
	wantError(t, cancelled.Place(), order.ErrState)
	wantError(t, cancelled.Cancel(), order.ErrState)
	wantError(t, cancelled.AddLine(line(t, "b", 1, 100, "USD")), order.ErrState)
	if cancelled.Status() != order.Cancelled || len(cancelled.Lines()) != 1 || cancelled.Total().Minor() != 100 {
		t.Fatal("cancelled order changed")
	}
}

func TestOrderLineLimit(t *testing.T) {
	o := newOrder(t)
	for i := 0; i < order.MaxLines; i++ {
		if err := o.AddLine(line(t, fmt.Sprint(i), 1, 1, "USD")); err != nil {
			t.Fatal(err)
		}
	}
	wantError(t, o.AddLine(line(t, "extra", 1, 1, "USD")), order.ErrLimit)
	if len(o.Lines()) != order.MaxLines || o.Total().Minor() != int64(order.MaxLines) {
		t.Fatal("limit failure changed state")
	}
}

func TestInvalidAndZeroOrders(t *testing.T) {
	_, err := order.NewOrder(" ", "USD", nil)
	wantError(t, err, order.ErrInvalid)
	_, err = order.NewOrder("id", "", nil)
	wantError(t, err, order.ErrInvalid)
	got, err := order.NewOrder("id", "USD", []order.Line{order.Line{}})
	if got != nil {
		t.Fatal("invalid construction returned aggregate")
	}
	wantError(t, err, order.ErrInvalid)
	for _, o := range []*order.Order{nil, new(order.Order)} {
		wantError(t, o.Place(), order.ErrInvalid)
		wantError(t, o.Cancel(), order.ErrInvalid)
		wantError(t, o.AddLine(line(t, "a", 1, 10, "USD")), order.ErrInvalid)
	}
}
