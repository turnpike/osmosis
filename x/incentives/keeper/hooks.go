package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	params := k.GetParams(ctx)
	if epochIdentifier == params.DistrEpochIdentifier {
		// begin distribution if it's start time
		gauges := k.GetUpcomingGauges(ctx)
		for _, gauge := range gauges {
			if !ctx.BlockTime().Before(gauge.StartTime) {
				if err := k.BeginDistribution(ctx, gauge); err != nil {
					panic(err)
				}
			}
		}

		// distribute due to epoch event
		gauges = k.GetActiveGauges(ctx)
		for _, gauge := range gauges {
			//_, err := k.Distribute(ctx, gauge)
			err := k.F1Distribute(ctx, &gauge)
			if err != nil {
				panic(err)
			}
			if !gauge.IsPerpetual && gauge.NumEpochsPaidOver <= gauge.FilledEpochs {
				if err := k.FinishDistribution(ctx, gauge); err != nil {
					panic(err)
				}
			}
		}
	}
}

// ___________________________________________________________________________________________________

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}

//////////////////////////// START //////////////////////////////////

var _ lockuptypes.LockupHooks = Hooks{}

func (h Hooks) OnTokenLocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	lockableDurations := h.k.GetLockableDurations(ctx)
	ctx.Logger().Info(fmt.Sprintf("STH::: token locked [%d]", lockID))
	for _, lockableDuration := range lockableDurations {
		if lockDuration < lockableDuration {
			continue
		}
		for _, coin := range amount {
			denom := coin.Denom
			currentReward, err := h.k.GetCurrentReward(ctx, denom, lockableDuration)
			if err != nil {
				panic(err)
			}
			ctx.Logger().Info(fmt.Sprintf("STH::: new historical reward[%d]", currentReward.Period))
			epochInfo := h.k.GetEpochInfo(ctx)
			h.k.CalculateHistoricalRewards(ctx, &currentReward, denom, lockableDuration, epochInfo)
			h.k.setCurrentReward(ctx, currentReward, denom, lockableDuration)

		}
		h.k.UpdateRewardForLock(ctx, lockID, lockableDuration)
	}
}

func (h Hooks) OnTokenUnlocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	lockableDurations := h.k.GetLockableDurations(ctx)
	for _, lockableDuration := range lockableDurations {
		if lockDuration < lockableDuration {
			continue
		}
		// for _, coin := range amount {
		// 	h.k.GetCurrentReward(coin.GetDenom(), lockableDuration).IsNewEpoch = true
		// }
		h.k.UpdateRewardForLock(ctx, lockID, lockableDuration)
		h.k.ClaimRewardForLock(ctx, lockID, lockableDuration)
	}
	h.k.clearPeriodLockReward(ctx, lockID)
}

////////////////////////////  END //////////////////////////////////
