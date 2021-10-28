package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
)

func (suite *KeeperTestSuite) TestBeginUnlocking() { // test for all unlockable coins
	suite.SetupTest()

	// initial check
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].EndTime, time.Time{})
	suite.Require().Equal(locks[0].IsUnlocking(), false)

	// begin unlock
	locks, unlockCoins, err := suite.app.LockupKeeper.BeginUnlockAllNotUnlockings(suite.ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(unlockCoins, coins)
	suite.Require().Equal(locks[0].ID, uint64(1))

	// check locks
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().NotEqual(locks[0].EndTime, time.Time{})
	suite.Require().NotEqual(locks[0].IsUnlocking(), false)
}

func (suite *KeeperTestSuite) TestBeginUnlockPeriodLock() {
	suite.SetupTest()

	// initial check
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].EndTime, time.Time{})
	suite.Require().Equal(locks[0].IsUnlocking(), false)

	// begin unlock
	lock1, err := suite.app.LockupKeeper.BeginUnlockPeriodLockByID(suite.ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(lock1.ID, uint64(1))

	// check locks
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().NotEqual(locks[0].EndTime, time.Time{})
	suite.Require().NotEqual(locks[0].IsUnlocking(), false)
}

func (suite *KeeperTestSuite) TestBeginPartialUnlockPeriodLock() {
	suite.SetupTest()

	// initial check
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].EndTime, time.Time{})
	suite.Require().Equal(locks[0].IsUnlocking(), false)

	// 1st begin unlock
	lock1, lock2, err := suite.app.LockupKeeper.BeginPartialUnlockPeriodLockByID(suite.ctx, 1, sdk.NewCoins(sdk.NewInt64Coin("stake", 1)))
	suite.Require().NoError(err)
	suite.Require().Equal(lock1.ID, uint64(1))
	suite.Require().Equal(lock2.ID, uint64(2))

	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)

	// check locks
	suite.Require().Len(locks, 2)
	suite.Require().Equal(locks[0].EndTime, time.Time{})
	suite.Require().Equal(locks[0].IsUnlocking(), false)
	suite.Require().NotEqual(locks[1].EndTime, time.Time{})
	suite.Require().Equal(locks[1].IsUnlocking(), true)

	// 2nd begin unlock
	lock1, lock2, err = suite.app.LockupKeeper.BeginPartialUnlockPeriodLockByID(suite.ctx, 1, sdk.NewCoins(sdk.NewInt64Coin("stake", 9)))
	suite.Require().NoError(err)
	suite.Require().Equal(lock1, (*types.PeriodLock)(nil))
	suite.Require().Equal(lock2.ID, uint64(1))

	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)

	// check locks
	suite.Require().Len(locks, 2)
	suite.Require().NotEqual(locks[0].EndTime, time.Time{})
	suite.Require().Equal(locks[0].IsUnlocking(), true)
	suite.Require().Equal(locks[0].Coins.AmountOf("stake"), sdk.NewInt(9))
	suite.Require().NotEqual(locks[1].EndTime, time.Time{})
	suite.Require().Equal(locks[1].IsUnlocking(), true)
	suite.Require().Equal(locks[1].Coins.AmountOf("stake"), sdk.NewInt(1))
}

func (suite *KeeperTestSuite) TestGetPeriodLocks() {
	suite.SetupTest()

	// initial check
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
}

func (suite *KeeperTestSuite) TestUnlockPeriodLockByID() {
	suite.SetupTest()
	now := suite.ctx.BlockTime()

	// initial check
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// unlock lock just now
	lock1, err := suite.app.LockupKeeper.UnlockPeriodLockByID(suite.ctx, 1)
	suite.Require().Error(err)
	suite.Require().Equal(lock1.ID, uint64(1))

	// unlock lock after 1s before starting unlock
	lock2, err := suite.app.LockupKeeper.UnlockPeriodLockByID(suite.ctx.WithBlockTime(now.Add(time.Second)), 1)
	suite.Require().Error(err)
	suite.Require().Equal(lock2.ID, uint64(1))

	// begin unlock
	lock3, err := suite.app.LockupKeeper.BeginUnlockPeriodLockByID(suite.ctx.WithBlockTime(now.Add(time.Second)), 1)
	suite.Require().NoError(err)
	suite.Require().Equal(lock3.ID, uint64(1))

	// unlock 1s after begin unlock
	lock4, err := suite.app.LockupKeeper.UnlockPeriodLockByID(suite.ctx.WithBlockTime(now.Add(time.Second*2)), 1)
	suite.Require().NoError(err)
	suite.Require().Equal(lock4.ID, uint64(1))

	// check locks
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)
}

func (suite *KeeperTestSuite) TestLock() {
	// test for coin locking
	suite.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	lock := types.NewPeriodLock(1, addr1, time.Second, suite.ctx.BlockTime().Add(time.Second), coins)

	// try lock without balance
	err := suite.app.LockupKeeper.Lock(suite.ctx, lock)
	suite.Require().Error(err)

	// lock with balance
	err = suite.app.BankKeeper.SetBalances(suite.ctx, addr1, coins)
	suite.Require().NoError(err)
	err = suite.app.LockupKeeper.Lock(suite.ctx, lock)
	suite.Require().NoError(err)

	// lock with balance with same id
	err = suite.app.BankKeeper.SetBalances(suite.ctx, addr1, coins)
	suite.Require().NoError(err)
	err = suite.app.LockupKeeper.Lock(suite.ctx, lock)
	suite.Require().Error(err)

	// lock with balance with different id
	lock = types.NewPeriodLock(2, addr1, time.Second, suite.ctx.BlockTime().Add(time.Second), coins)
	err = suite.app.BankKeeper.SetBalances(suite.ctx, addr1, coins)
	suite.Require().NoError(err)
	err = suite.app.LockupKeeper.Lock(suite.ctx, lock)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestUnlock() {
	// test for coin unlocking
	suite.SetupTest()
	now := suite.ctx.BlockTime()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	lock := types.NewPeriodLock(1, addr1, time.Second, now.Add(time.Second), coins)

	// lock with balance
	err := suite.app.BankKeeper.SetBalances(suite.ctx, addr1, coins)
	suite.Require().NoError(err)
	err = suite.app.LockupKeeper.Lock(suite.ctx, lock)
	suite.Require().NoError(err)

	// begin unlock with lock object
	err = suite.app.LockupKeeper.BeginUnlock(suite.ctx, lock)
	suite.Require().NoError(err)

	// unlock with lock object
	err = suite.app.LockupKeeper.Unlock(suite.ctx.WithBlockTime(now.Add(time.Second)), lock)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestModuleLockedCoins() {
	suite.SetupTest()

	// initial check
	lockedCoins := suite.app.LockupKeeper.GetModuleLockedCoins(suite.ctx)
	suite.Require().Equal(lockedCoins, sdk.Coins(nil))

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// final check
	lockedCoins = suite.app.LockupKeeper.GetModuleLockedCoins(suite.ctx)
	suite.Require().Equal(lockedCoins, coins)
}

func (suite *KeeperTestSuite) TestLocksPastTimeDenom() {
	suite.SetupTest()

	now := time.Now()
	suite.ctx = suite.ctx.WithBlockTime(now)

	// initial check
	locks := suite.app.LockupKeeper.GetLocksPastTimeDenom(suite.ctx, "stake", now)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	coins = sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Minute)

	// check locks
	locks = suite.app.LockupKeeper.GetLocksPastTimeDenom(suite.ctx, "stake", now)
	suite.Require().Len(locks, 2)

	// unlock 1 sec lock
	for _, lock := range locks {
		if lock.Duration == time.Second {
			suite.app.LockupKeeper.BeginUnlock(suite.ctx, lock)
			break
		}
	}

	// final check
	locks = suite.app.LockupKeeper.GetLocksPastTimeDenom(suite.ctx, "stake", now.Add(time.Second))
	suite.Require().Len(locks, 1)
}

func (suite *KeeperTestSuite) TestLocksLongerThanDurationDenom() {
	suite.SetupTest()

	// initial check
	duration := time.Second
	locks := suite.app.LockupKeeper.GetLocksLongerThanDurationDenom(suite.ctx, "stake", duration)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// final check
	locks = suite.app.LockupKeeper.GetLocksLongerThanDurationDenom(suite.ctx, "stake", duration)
	suite.Require().Len(locks, 1)
}

func (suite *KeeperTestSuite) TestLockTokensAlot() {
	suite.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	startAveragingAt := 1000
	totalNumLocks := 10000
	for i := 1; i < startAveragingAt; i++ {
		suite.LockTokens(addr1, coins, time.Second)
	}
	runningTotal := uint64(0)
	maxGas := uint64(0)
	for i := startAveragingAt; i < totalNumLocks; i++ {
		if i%1000 == 0 {
			fmt.Printf("entering %dth lock now\n", i)
		}

		alreadySpent := suite.ctx.GasMeter().GasConsumed()
		suite.LockTokens(addr1, coins, time.Second)
		newSpent := suite.ctx.GasMeter().GasConsumed()
		spentNow := newSpent - alreadySpent
		runningTotal += spentNow
		if spentNow > maxGas {
			maxGas = spentNow
		}
	}
	fmt.Printf("test deets: total locks created %d, begin average at %d\n", totalNumLocks, startAveragingAt)
	fmt.Println("average gas / lock:", runningTotal/(uint64(totalNumLocks-startAveragingAt)))
	fmt.Println("max gas / lock:", maxGas)

	// panic(1)
}

func (suite *KeeperTestSuite) TestAddTokensToLock() {
	suite.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].Coins, coins)
	// check accumulation store is correctly updated
	accum := suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")

	// add more tokens to lock
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	err = suite.app.BankKeeper.SetBalances(suite.ctx, addr1, addCoins)
	suite.Require().NoError(err)
	_, err = suite.app.LockupKeeper.AddTokensToLockByID(suite.ctx, addr1, locks[0].ID, addCoins)
	suite.Require().NoError(err)

	// check locks after adding tokens to lock
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].Coins, coins.Add(addCoins...))

	// check accumulation store is correctly updated
	accum = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "20")

	// try to add tokens to unavailable lock
	cacheCtx, _ := suite.ctx.CacheContext()
	err = suite.app.BankKeeper.SetBalances(cacheCtx, addr1, addCoins)
	suite.Require().NoError(err)
	_, err = suite.app.LockupKeeper.AddTokensToLockByID(cacheCtx, addr1, 1111, addCoins)
	suite.Require().Error(err)

	// try to add tokens with lack balance
	cacheCtx, _ = suite.ctx.CacheContext()
	_, err = suite.app.LockupKeeper.AddTokensToLockByID(cacheCtx, addr1, locks[0].ID, addCoins)
	suite.Require().Error(err)

	// try to add tokens to lock that is owned by others
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	err = suite.app.BankKeeper.SetBalances(cacheCtx, addr2, addCoins)
	suite.Require().NoError(err)
	_, err = suite.app.LockupKeeper.AddTokensToLockByID(cacheCtx, addr2, locks[0].ID, addCoins)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestEndblockerWithdrawAllMaturedLockups() {
	suite.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	totalCoins := coins.Add(coins...).Add(coins...)

	// lock coins for 5 second, 1 seconds, and 3 seconds in that order
	times := []time.Duration{time.Second * 5, time.Second, time.Second * 3}
	sortedTimes := []time.Duration{time.Second, time.Second * 3, time.Second * 5}
	unbondBlockTimes := make([]time.Time, len(times))

	// setup locks for 5 second, 1 second, and 3 seconds, and begin unbonding them.
	setupInitLocks := func() {
		for i := 0; i < len(times); i++ {
			unbondBlockTimes[i] = suite.ctx.BlockTime().Add(sortedTimes[i])
		}

		for i := 0; i < len(times); i++ {
			suite.LockTokens(addr1, coins, times[i])
		}

		// consistency check locks
		locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
		suite.Require().NoError(err)
		suite.Require().Len(locks, 3)
		for i := 0; i < len(times); i++ {
			suite.Require().Equal(locks[i].EndTime, time.Time{})
			suite.Require().Equal(locks[i].IsUnlocking(), false)
		}

		// begin unlock
		locks, unlockCoins, err := suite.app.LockupKeeper.BeginUnlockAllNotUnlockings(suite.ctx, addr1)
		suite.Require().NoError(err)
		suite.Require().Len(locks, len(times))
		suite.Require().Equal(unlockCoins, totalCoins)
		for i := 0; i < len(times); i++ {
			suite.Require().Equal(locks[i].ID, uint64(i+1))
		}

		// check locks, these should now be sorted by unbonding completion time
		locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
		suite.Require().NoError(err)
		suite.Require().Len(locks, 3)
		for i := 0; i < 3; i++ {
			suite.Require().NotEqual(locks[i].EndTime, time.Time{})
			suite.Require().Equal(locks[i].EndTime, unbondBlockTimes[i])
			suite.Require().Equal(locks[i].IsUnlocking(), true)
		}
	}
	setupInitLocks()

	// try withdrawing before mature
	suite.app.LockupKeeper.WithdrawAllMaturedLocks(suite.ctx)
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 3)

	// withdraw at 1 sec, 3 sec, and 5 sec intervals, check automatically withdrawn
	for i := 0; i < len(times); i++ {
		suite.app.LockupKeeper.WithdrawAllMaturedLocks(suite.ctx.WithBlockTime(unbondBlockTimes[i]))
		locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
		suite.Require().NoError(err)
		suite.Require().Len(locks, len(times)-i-1)
	}
	suite.Require().Equal(suite.app.BankKeeper.GetAccountsBalances(suite.ctx)[1].Address, addr1.String())
	suite.Require().Equal(suite.app.BankKeeper.GetAccountsBalances(suite.ctx)[1].Coins, totalCoins)

	suite.SetupTest()
	setupInitLocks()
	// now withdraw all locks and ensure all got withdrawn
	suite.app.LockupKeeper.WithdrawAllMaturedLocks(suite.ctx.WithBlockTime(unbondBlockTimes[len(times)-1]))
	suite.Require().Len(locks, 0)
}

func (suite *KeeperTestSuite) TestLockAccumulationStore() {
	suite.SetupTest()

	// initial check
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr := sdk.AccAddress([]byte("addr1---------------"))

	// 1 * time.Second: 10 + 20
	// 2 * time.Second: 20 + 30
	// 3 * time.Second: 30
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr, coins, time.Second)
	coins = sdk.Coins{sdk.NewInt64Coin("stake", 20)}
	suite.LockTokens(addr, coins, time.Second)
	suite.LockTokens(addr, coins, time.Second*2)
	coins = sdk.Coins{sdk.NewInt64Coin("stake", 30)}
	suite.LockTokens(addr, coins, time.Second*2)
	suite.LockTokens(addr, coins, time.Second*3)

	// check accumulations
	acc := suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: 0,
	})
	suite.Require().Equal(int64(110), acc.Int64())
	acc = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 1,
	})
	suite.Require().Equal(int64(110), acc.Int64())
	acc = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 2,
	})
	suite.Require().Equal(int64(80), acc.Int64())
	acc = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 3,
	})
	suite.Require().Equal(int64(30), acc.Int64())
	acc = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		Denom:    "stake",
		Duration: time.Second * 4,
	})
	suite.Require().Equal(int64(0), acc.Int64())
}

func (suite *KeeperTestSuite) TestGetUnlockingsBetweenTimeDenom() {
	suite.SetupTest()

	// initial check
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	addr := sdk.AccAddress([]byte("addr1---------------"))

	// lock 1sec duration coins
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr, coins, time.Second)

	// lock 3sec duration coins
	coins = sdk.Coins{sdk.NewInt64Coin("stake", 20)}
	suite.LockTokens(addr, coins, time.Second*3)

	// Get 1~4sec endTime unlocking locks
	beginTime := suite.ctx.BlockTime().Add(time.Second)
	endTime := suite.ctx.BlockTime().Add(time.Second * 4)
	locks = suite.app.LockupKeeper.GetUnlockingsBetweenTimeDenom(suite.ctx, "stake", beginTime, endTime)
	suite.Require().Len(locks, 0)

	// unlock 1sec duration coins
	lock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, 1)
	suite.app.LockupKeeper.BeginUnlock(suite.ctx, *lock)

	locks = suite.app.LockupKeeper.GetUnlockingsBetweenTimeDenom(suite.ctx, "stake", beginTime, endTime)
	suite.Require().Len(locks, 1)

	// unlock 3sec duration coins
	lock, err = suite.app.LockupKeeper.GetLockByID(suite.ctx, 2)
	suite.app.LockupKeeper.BeginUnlock(suite.ctx, *lock)

	locks = suite.app.LockupKeeper.GetUnlockingsBetweenTimeDenom(suite.ctx, "stake", beginTime, endTime)
	suite.Require().Len(locks, 2)

	// Get 2~4sec endTime unlocking locks
	beginTime = suite.ctx.BlockTime().Add(time.Second * 2)
	endTime = suite.ctx.BlockTime().Add(time.Second * 4)
	locks = suite.app.LockupKeeper.GetUnlockingsBetweenTimeDenom(suite.ctx, "stake", beginTime, endTime)
	suite.Require().Len(locks, 1)
}
