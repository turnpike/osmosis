package keeper

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NOTE: never use new(Dec) or else we will panic unmarshalling into the
// nil embedded big.Int
type mutableDec struct {
	i *big.Int
}

var (
	precisionReuse = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	fivePrecision  = new(big.Int).Quo(precisionReuse, big.NewInt(2))
	oneInt         = big.NewInt(1)
)

func NewMutableDecFromDec(dec sdk.Dec) *mutableDec {
	return &mutableDec{
		dec.BigInt(),
	}
}

func NewMutableDecFromInt64(i int64) *mutableDec {
	return &mutableDec{
		sdk.NewDec(i).BigInt(),
	}
}

func (d *mutableDec) IsNil() bool               { return d.i == nil }           // is decimal nil
func (d *mutableDec) IsZero() bool              { return (d.i).Sign() == 0 }    // is equal to zero
func (d *mutableDec) IsNegative() bool          { return (d.i).Sign() == -1 }   // is negative
func (d *mutableDec) IsPositive() bool          { return (d.i).Sign() == 1 }    // is positive
func (d *mutableDec) Equal(d2 *mutableDec) bool { return (d.i).Cmp(d2.i) == 0 } // equal decimals
func (d *mutableDec) GT(d2 *mutableDec) bool    { return (d.i).Cmp(d2.i) > 0 }  // greater than
func (d *mutableDec) GTE(d2 *mutableDec) bool   { return (d.i).Cmp(d2.i) >= 0 } // greater than or equal
func (d *mutableDec) LT(d2 *mutableDec) bool    { return (d.i).Cmp(d2.i) < 0 }  // less than
func (d *mutableDec) LTE(d2 *mutableDec) bool   { return (d.i).Cmp(d2.i) <= 0 } // less than or equal
func (d *mutableDec) Neg() *mutableDec {
	d.i.Neg(d.i)
	return d
} // reverse the decimal sign
func (d *mutableDec) Abs() *mutableDec {
	d.i.Abs(d.i)
	return d
} // absolute value

// addition
func (d *mutableDec) Add(d2 *mutableDec) *mutableDec {
	d.i.Add(d.i, d2.i)

	if d.i.BitLen() > 255+sdk.DecimalPrecisionBits {
		panic("Int overflow")
	}
	return d
}

// subtraction
func (d *mutableDec) Sub(d2 *mutableDec) *mutableDec {
	d.i.Sub(d.i, d2.i)

	if d.i.BitLen() > 255+sdk.DecimalPrecisionBits {
		panic("Int overflow")
	}
	return d
}

// quotient
func (d1 *mutableDec) Quo(d2 *mutableDec) *mutableDec {
	// multiply precision twice
	d1.i.Mul(d1.i, precisionReuse)
	d1.i.Mul(d1.i, precisionReuse)

	d1.i.Quo(d1.i, d2.i)
	chopPrecisionAndRound(d1.i)

	return d1
}

// multiplication
func (d1 *mutableDec) Mul(d2 *mutableDec) *mutableDec {
	d1.i.Mul(d1.i, d2.i)
	chopPrecisionAndRound(d1.i)

	return d1
}

func (d1 *mutableDec) SetInt64(i int64) *mutableDec {
	d1.i.SetInt64(i).Mul(d1.i, precisionReuse)

	return d1
}

// MulInt64 - multiplication with int64
func (d1 *mutableDec) MulInt64(i int64) *mutableDec {
	d1.i.Mul(d1.i, big.NewInt(i))

	if d1.i.BitLen() > 255+sdk.DecimalPrecisionBits {
		panic("Int overflow")
	}
	return d1
}

func (d *mutableDec) Clone() *mutableDec {
	return &mutableDec{
		new(big.Int).Set(d.i),
	}
}

func (d *mutableDec) String() string {
	return d.Dec().String()
}

func (d *mutableDec) Dec() sdk.Dec {
	return sdk.NewDecFromBigIntWithPrec(d.i, sdk.Precision)
}

//     ____
//  __|    |__   "chop 'em
//       ` \     round!"
// ___||  ~  _     -bankers
// |         |      __
// |       | |   __|__|__
// |_____:  /   | $$$    |
//              |________|

// Remove a Precision amount of rightmost digits and perform bankers rounding
// on the remainder (gaussian rounding) on the digits which have been removed.
//
// Mutates the input. Use the non-mutative version if that is undesired
func chopPrecisionAndRound(d *big.Int) *big.Int {
	// remove the negative and add it back when returning
	if d.Sign() == -1 {
		// make d positive, compute chopped value, and then un-mutate d
		d = d.Neg(d)
		d = chopPrecisionAndRound(d)
		d = d.Neg(d)
		return d
	}

	// get the truncated quotient and remainder
	quo, rem := d, big.NewInt(0)
	quo, rem = quo.QuoRem(d, precisionReuse, rem)

	if rem.Sign() == 0 { // remainder is zero
		return quo
	}

	switch rem.Cmp(fivePrecision) {
	case -1:
		return quo
	case 1:
		return quo.Add(quo, oneInt)
	default: // bankers rounding must take place
		// always round to an even number
		if quo.Bit(0) == 0 {
			return quo
		}
		return quo.Add(quo, oneInt)
	}
}
