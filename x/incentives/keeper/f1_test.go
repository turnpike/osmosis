package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/epochs"
	"github.com/osmosis-labs/osmosis/x/epochs/types"
	incentivestypes "github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
)

var (
	DayDuration          = time.Hour * 24
	SevenDaysDuration    = time.Hour * 24 * 7
	FourteenDaysDuration = time.Hour * 24 * 14
	OwnerAddr            = "addr1---------------"
	User1Addr            = "user1---------------"
	User2Addr            = "user2---------------"
)

func (suite *KeeperTestSuite) TestF1Distribute() {
	now, height := suite.setupEpochAndLockableDurations()

	//next epoch
	suite.nextEpoch(&now, &height)

	//new gauge
	defaultGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: DayDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin("stake", 1000)},
	}
	gauges := suite.SetupGauges([]perpGaugeDesc{defaultGauge})
	initGauge := gauges[0]
	denom := initGauge.DistributeTo.Denom
	duration := initGauge.DistributeTo.Duration
	owner := sdk.AccAddress(OwnerAddr)

	suite.nextBlock(&now, &height)

	//1st lock
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, DayDuration)

	//next epoch
	suite.nextEpoch(&now, &height)

	currentReward, err := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, duration)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt64Coin("lptoken", 10), currentReward.Coin)
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 1000)}, currentReward.Rewards)
	suite.T().Logf("current_reward=%v", currentReward)

	//2nd lock
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 40)}, DayDuration)

	//next epoch
	suite.nextEpoch(&now, &height)

	currentReward, err = suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, duration)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt64Coin("lptoken", 50), currentReward.Coin)
	suite.Require().Equal(sdk.Coins(nil), currentReward.Rewards)
	suite.T().Logf("current_reward=%v", currentReward)

	prevHistoricalReward, err := suite.app.IncentivesKeeper.GetHistoricalReward(suite.ctx, denom, duration, currentReward.Period-1)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewInt64DecCoin("stake", 100), prevHistoricalReward.CummulativeRewardRatio[0])
	suite.T().Logf("historical_reward=%v", prevHistoricalReward)
}

func (suite *KeeperTestSuite) setupEpochAndLockableDurations() (time.Time, int64) {
	epochInfos := suite.app.EpochsKeeper.AllEpochInfos(suite.ctx)
	for _, epochInfo := range epochInfos {
		suite.app.EpochsKeeper.DeleteEpochInfo(suite.ctx, epochInfo.Identifier)
	}

	//now, _ := time.Parse("2006-01-02", "2021-10-01")
	now := time.Now()
	height := int64(1)
	suite.ctx = suite.ctx.WithBlockHeight(height).WithBlockTime(now)

	epochs.InitGenesis(suite.ctx, suite.app.EpochsKeeper, types.GenesisState{
		Epochs: []types.EpochInfo{
			{
				Identifier:            "day",
				StartTime:             now,
				Duration:              DayDuration,
				CurrentEpoch:          0,
				CurrentEpochStartTime: time.Time{},
				EpochCountingStarted:  false,
			},
		},
	})

	// add additional lockable durations
	suite.app.IncentivesKeeper.SetParams(suite.ctx, incentivestypes.Params{DistrEpochIdentifier: "day"})
	suite.app.IncentivesKeeper.SetLockableDurations(suite.ctx,
		[]time.Duration{
			time.Hour * 24,
			time.Hour * 24 * 7,
			time.Hour * 24 * 14})

	return now, height
}

func (suite *KeeperTestSuite) nextEpoch(now *time.Time, height *int64) {
	epochInfo := suite.app.EpochsKeeper.AllEpochInfos(suite.ctx)[0]
	if epochInfo.StartTime.After(epochInfo.CurrentEpochStartTime) {
		*now = (*now).Add(DayDuration + time.Second)
	} else {
		*now = epochInfo.CurrentEpochStartTime.Add(DayDuration + time.Second)
	}
	*height = *height + 1
	suite.ctx = suite.ctx.WithBlockHeight(*height).WithBlockTime(*now)
	suite.app.LockupKeeper.WithdrawAllMaturedLocks(suite.ctx)
	epochs.BeginBlocker(suite.ctx, suite.app.EpochsKeeper)
}

func (suite *KeeperTestSuite) nextBlock(now *time.Time, height *int64) {
	*now = (*now).Add(time.Second)
	*height = *height + 1
	suite.ctx = suite.ctx.WithBlockHeight(*height).WithBlockTime(*now)
	suite.app.LockupKeeper.WithdrawAllMaturedLocks(suite.ctx)
	epochs.BeginBlocker(suite.ctx, suite.app.EpochsKeeper)
}

func (suite *KeeperTestSuite) setupNonPerpetualGauge(owner sdk.AccAddress, now time.Time, duration time.Duration) (string, time.Duration) {
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 1400)}
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      duration,
	}
	numEpochsPaidOver := uint64(14)
	_, initGauge := suite.CreateGauge(false, owner, coins, distrTo, now, numEpochsPaidOver)
	return initGauge.DistributeTo.Denom, initGauge.DistributeTo.Duration
}

func (suite *KeeperTestSuite) TestLockFor1Day() {
	now, height := suite.setupEpochAndLockableDurations()
	suite.nextEpoch(&now, &height)

	//new non-perpetual gauge
	owner := sdk.AccAddress(OwnerAddr)
	denom, duration := suite.setupNonPerpetualGauge(owner, now, DayDuration)

	suite.nextBlock(&now, &height)

	//lock
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, DayDuration)

	for i := 2; i <= 15; i++ {
		suite.nextEpoch(&now, &height)
		currentReward, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, duration)
		suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(100*(i-1)))}, currentReward.Rewards)
		suite.T().Logf("period=%v, current_reward=%v", i, currentReward.Rewards)
	}
}

func (suite *KeeperTestSuite) TestLockFor14Days() {
	now, height := suite.setupEpochAndLockableDurations()
	suite.nextEpoch(&now, &height)

	//new non-perpetual gauge
	owner := sdk.AccAddress(OwnerAddr)
	denom, _ := suite.setupNonPerpetualGauge(owner, now, DayDuration)
	suite.setupNonPerpetualGauge(owner, now, SevenDaysDuration)
	suite.setupNonPerpetualGauge(owner, now, FourteenDaysDuration)

	suite.nextBlock(&now, &height)

	//lock
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, FourteenDaysDuration)

	for i := 2; i <= 15; i++ {
		suite.nextEpoch(&now, &height)
		currentReward1Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, DayDuration)
		currentReward7Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, SevenDaysDuration)
		currentReward14Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, FourteenDaysDuration)
		//suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(100*i))}, currentReward.Rewards)
		suite.T().Logf("period=%v, current_reward={1d=%v, 7d=%v, 14d=%v}", i, currentReward1Day.Rewards, currentReward7Day.Rewards, currentReward14Day.Rewards)
	}
}

func (suite *KeeperTestSuite) TestLockAndUnlockFor14Days() {
	now, height := suite.setupEpochAndLockableDurations()
	suite.nextEpoch(&now, &height)

	//new non-perpetual gauge
	owner := sdk.AccAddress(OwnerAddr)
	denom, _ := suite.setupNonPerpetualGauge(owner, now, DayDuration)
	suite.setupNonPerpetualGauge(owner, now, SevenDaysDuration)
	suite.setupNonPerpetualGauge(owner, now, FourteenDaysDuration)

	suite.nextBlock(&now, &height)

	stakeBalance := suite.app.BankKeeper.GetBalance(suite.ctx, owner, "stake")

	//lock
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, FourteenDaysDuration)

	//unlock
	suite.app.LockupKeeper.BeginUnlockPeriodLockByID(suite.ctx, 1)

	for i := 2; i <= 15; i++ {
		suite.nextEpoch(&now, &height)
		currentReward1Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, DayDuration)
		currentReward7Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, SevenDaysDuration)
		currentReward14Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, FourteenDaysDuration)
		//suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(100*i))}, currentReward.Rewards)
		suite.T().Logf("period=%v, current_reward={1d=%v, 7d=%v, 14d=%v}", i, currentReward1Day, currentReward7Day, currentReward14Day)

		lock, _ := suite.app.LockupKeeper.GetLockByID(suite.ctx, 1)
		estLockReward, err := suite.app.IncentivesKeeper.EstimateLockReward(suite.ctx, *lock)
		suite.Require().NoError(err)
		suite.T().Logf("period=%v, EstimatedLockReward=%v", i, estLockReward.Rewards)
	}

	suite.nextEpoch(&now, &height)
	suite.nextEpoch(&now, &height)
	suite.nextEpoch(&now, &height)
	suite.nextEpoch(&now, &height)

	epochInfo := suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "day")
	lock, _ := suite.app.LockupKeeper.GetLockByID(suite.ctx, 1)
	currentReward1Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, DayDuration)
	currentReward7Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, SevenDaysDuration)
	currentReward14Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, FourteenDaysDuration)
	suite.T().Logf(">>> END <<< ")
	suite.T().Logf(" - epoch=%v", epochInfo.CurrentEpoch)
	suite.T().Logf(" - lock=%v", lock)
	suite.T().Logf(" - current_reward={1d=%v, 7d=%v, 14d=%v}", currentReward1Day, currentReward7Day, currentReward14Day)

	newStakeBalance := suite.app.BankKeeper.GetBalance(suite.ctx, owner, "stake")
	suite.T().Logf(" - stakeBalance %v->%v", stakeBalance, newStakeBalance)
}

func (suite *KeeperTestSuite) TestMultipleTokenLockOnSingleLock() {
	now, height := suite.setupEpochAndLockableDurations()
	suite.nextEpoch(&now, &height)

	//new non-perpetual gauge
	owner := sdk.AccAddress(OwnerAddr)
	denom, duration := suite.setupNonPerpetualGauge(owner, now, DayDuration)

	suite.nextBlock(&now, &height)

	//lock
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, DayDuration)
	periodLockReward, err := suite.app.IncentivesKeeper.GetPeriodLockReward(suite.ctx, 1)
	suite.Require().NoError(err)
	suite.T().Logf("periodLockReward=%v", periodLockReward)

	suite.nextEpoch(&now, &height)
	suite.nextEpoch(&now, &height)
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 1000)}, DayDuration)
	periodLockReward, _ = suite.app.IncentivesKeeper.GetPeriodLockReward(suite.ctx, 1)
	suite.T().Logf("periodLockReward=%v", periodLockReward)

	suite.nextEpoch(&now, &height)
	suite.nextEpoch(&now, &height)
	suite.LockTokens(owner, sdk.Coins{sdk.NewInt64Coin("lptoken", 1000000)}, DayDuration)
	periodLockReward, _ = suite.app.IncentivesKeeper.GetPeriodLockReward(suite.ctx, 1)
	suite.T().Logf("periodLockReward=%v", periodLockReward)

	for i := 6; i <= 15; i++ {
		suite.nextEpoch(&now, &height)
		currentReward, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, duration)
		suite.T().Logf("period=%v, current_reward=%v", i, currentReward.Rewards)
		periodLockReward, _ = suite.app.IncentivesKeeper.GetPeriodLockReward(suite.ctx, 1)
		suite.T().Logf("periodLockReward=%v", periodLockReward)
	}
	lock, _ := suite.app.LockupKeeper.GetLockByID(suite.ctx, 1)
	periodLockReward, err = suite.app.IncentivesKeeper.GetPeriodLockReward(suite.ctx, 1)
	suite.Require().NoError(err)

	epochInfo := suite.app.EpochsKeeper.AllEpochInfos(suite.ctx)[0]
	durations := []time.Duration{duration}
	err = suite.app.IncentivesKeeper.UpdateHistoricalRewardFromCurrentReward(suite.ctx, lock.Coins, duration, epochInfo, durations)
	suite.Require().NoError(err)
	err = suite.app.IncentivesKeeper.UpdateRewardForLock(suite.ctx, *lock, periodLockReward, epochInfo, durations)
	suite.Require().NoError(err)
	periodLockReward, _ = suite.app.IncentivesKeeper.GetPeriodLockReward(suite.ctx, 1)
	suite.T().Logf("periodLockReward=%v", periodLockReward)
	suite.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin("stake", 1400)), periodLockReward.Rewards)
}

func (suite *KeeperTestSuite) TestMultipleLocks() {
	now, height := suite.setupEpochAndLockableDurations()
	suite.nextEpoch(&now, &height)

	//new non-perpetual gauge
	owner := sdk.AccAddress(OwnerAddr)
	denom, _ := suite.setupNonPerpetualGauge(owner, now, DayDuration)
	suite.setupNonPerpetualGauge(owner, now, FourteenDaysDuration)

	suite.nextBlock(&now, &height)

	//lock
	suite.T().Logf(">>> LOCK <<< : user1=10lptoken,1day; user2=10lptoken,14days ")
	user1 := sdk.AccAddress(User1Addr)
	user2 := sdk.AccAddress(User2Addr)
	suite.LockTokens(user1, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, DayDuration)
	suite.LockTokens(user2, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, FourteenDaysDuration)

	for i := 0; i < 3; i++ {
		epochInfo := suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "day")
		suite.nextEpoch(&now, &height)
		currentReward1Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, DayDuration)
		currentReward7Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, SevenDaysDuration)
		currentReward14Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, FourteenDaysDuration)
		//suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(100*i))}, currentReward.Rewards)
		suite.T().Logf("epoch(%v) : current_reward={1d=%v, 7d=%v, 14d=%v}", epochInfo.CurrentEpoch, currentReward1Day, currentReward7Day, currentReward14Day)

		user1Lock, _ := suite.app.LockupKeeper.GetLockByID(suite.ctx, 1)
		user1Reward, _ := suite.app.IncentivesKeeper.EstimateLockReward(suite.ctx, *user1Lock)
		user2Lock, _ := suite.app.LockupKeeper.GetLockByID(suite.ctx, 2)
		user2Reward, _ := suite.app.IncentivesKeeper.EstimateLockReward(suite.ctx, *user2Lock)
		suite.T().Logf("           estimated_reward={user1=%v, user2=%v}", user1Reward.Rewards, user2Reward.Rewards)
	}

	suite.T().Logf(">>> UNLOCK <<< : user1=10lptoken,1day ")
	suite.app.LockupKeeper.BeginUnlockPeriodLockByID(suite.ctx, 1)

	for i := 0; i < 3; i++ {
		epochInfo := suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "day")
		suite.nextEpoch(&now, &height)
		currentReward1Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, DayDuration)
		currentReward7Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, SevenDaysDuration)
		currentReward14Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, FourteenDaysDuration)
		//suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(100*i))}, currentReward.Rewards)
		suite.T().Logf("epoch(%v) : current_reward={1d=%v, 7d=%v, 14d=%v}", epochInfo.CurrentEpoch, currentReward1Day, currentReward7Day, currentReward14Day)

		user2Lock, _ := suite.app.LockupKeeper.GetLockByID(suite.ctx, 2)
		user2Reward, _ := suite.app.IncentivesKeeper.EstimateLockReward(suite.ctx, *user2Lock)
		suite.T().Logf("           estimated_reward={user2=%v}", user2Reward.Rewards)
	}

	suite.T().Logf(">>> UNLOCK <<< : user2=10lptoken,14days ")
	suite.app.LockupKeeper.BeginUnlockPeriodLockByID(suite.ctx, 2)

	for i := 0; i < 13; i++ {
		epochInfo := suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "day")
		suite.nextEpoch(&now, &height)
		currentReward1Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, DayDuration)
		currentReward7Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, SevenDaysDuration)
		currentReward14Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, FourteenDaysDuration)
		//suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(100*i))}, currentReward.Rewards)
		suite.T().Logf("epoch(%v) : current_reward={1d=%v, 7d=%v, 14d=%v}", epochInfo.CurrentEpoch, currentReward1Day, currentReward7Day, currentReward14Day)

		user2Lock, _ := suite.app.LockupKeeper.GetLockByID(suite.ctx, 2)
		user2Reward, _ := suite.app.IncentivesKeeper.EstimateLockReward(suite.ctx, *user2Lock)
		suite.T().Logf("           estimated_reward={user2=%v}", user2Reward.Rewards)
	}

	epochInfo := suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "day")
	currentReward1Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, DayDuration)
	currentReward7Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, SevenDaysDuration)
	currentReward14Day, _ := suite.app.IncentivesKeeper.GetCurrentReward(suite.ctx, denom, FourteenDaysDuration)
	suite.T().Logf(">>> END <<< ")
	suite.T().Logf("epoch(%v) : current_reward={1d=%v, 7d=%v, 14d=%v}", epochInfo.CurrentEpoch, currentReward1Day, currentReward7Day, currentReward14Day)

	user2Lock, _ := suite.app.LockupKeeper.GetLockByID(suite.ctx, 2)
	user2Reward, _ := suite.app.IncentivesKeeper.EstimateLockReward(suite.ctx, *user2Lock)
	suite.T().Logf("           estimated_reward={user2=%v}", user2Reward.Rewards)
	claimedReward, _ := suite.app.IncentivesKeeper.ClaimLockReward(suite.ctx, *user2Lock, owner)
	suite.T().Logf("           claimed_reward={user2=%v}", claimedReward)
}
