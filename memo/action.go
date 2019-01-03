package memo

import (
	"errors"

	"github.com/shopspring/decimal"
)

type Arena struct {
	E int64  // expire
	M string // max bet
}

func NewArena(memo string) (*Arena, error) {
	a := Arena{}
	if err := Unmarshal(memo, &a); err != nil {
		return nil, err
	}

	if a.E < 1 {
		return nil, errors.New("invalid expire option")
	}

	if max, _ := decimal.NewFromString(a.M); !max.IsPositive() {
		return nil, errors.New("max bet must be positive")
	}

	return &a, nil
}

type Bet struct {
	A string // arena id
	G string // gestures
}

func NewBet(memo string) (*Bet, error) {
	b := Bet{}
	if err := Unmarshal(memo, &b); err != nil {
		return nil, err
	}

	if len(b.A) == 0 {
		return nil, errors.New("empty arena id")
	}

	if len(b.G) == 0 {
		return nil, errors.New("no gesture")
	}

	return &b, nil
}

type Login struct {
	T int64 // timestamp is seconds
}

func NewLogin(memo string) (*Login, error) {
	l := Login{}
	if err := Unmarshal(memo, &l); err != nil {
		return nil, err
	}

	if l.T <= 0 {
		return nil, errors.New("timestamp is invalid")
	}

	return &l, nil
}
