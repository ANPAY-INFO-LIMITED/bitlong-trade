package account

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"sync"
	"time"
	"trade/btlLog"
	"trade/models"
	cModels "trade/models/custodyModels"
	"trade/services/btldb"
)

const interval = 20 * time.Second

var (
	ErrMaxUserPoolReached = fmt.Errorf("用户池已满，无法添加新用户")
	ErrUserNotFound       = fmt.Errorf("用户不存在")
	ErrUserLocked         = fmt.Errorf("用户已被冻结")
)

var pool *UserPool

type UserInfo struct {
	User          *models.User
	Account       *models.Account
	LockAccount   *cModels.LockAccount
	PaymentMux    sync.Mutex
	LastPayTime   time.Time
	RpcMux        sync.Mutex
	LastActiveMux sync.Mutex
	LastActive    time.Time
}

func (u *UserInfo) PayLock() bool {
	currentTime := time.Now()
	if currentTime.Sub(u.LastPayTime) < interval {
		return false
	}
	u.PaymentMux.Lock()
	return true
}

func (u *UserInfo) PayUnlock() {
	u.LastPayTime = time.Now()
	u.PaymentMux.Unlock()
}

type UserPool struct {
	users            map[string]*UserInfo
	mutex            sync.RWMutex
	maxCapacity      int
	inactiveDuration time.Duration
}

func init() {
	pool = NewUserPool(2000, 3*time.Minute)
	pool.StartCleanupScheduler(5 * time.Minute)
}

func NewUserPool(maxCapacity int, inactiveDuration time.Duration) *UserPool {
	if maxCapacity <= 0 {
		maxCapacity = 2000
	}
	return &UserPool{
		users:            make(map[string]*UserInfo),
		maxCapacity:      maxCapacity,
		inactiveDuration: inactiveDuration,
	}
}

func (pool *UserPool) RemoveUser(userName string) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if _, exists := pool.users[userName]; exists {
		delete(pool.users, userName)
		btlLog.CACC.Info("用户 %s 已被删除\n", userName)
	} else {
		btlLog.CACC.Warning("警告: 用户 %s 不存在，无法删除\n", userName)
	}
}

func (pool *UserPool) GetUser(userName string) (*UserInfo, bool) {
	pool.mutex.RLock()
	defer pool.mutex.RUnlock()

	user, exists := pool.users[userName]
	return user, exists
}

func (pool *UserPool) CreateUser(userName string) (*UserInfo, error) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if existingUser, exists := pool.users[userName]; exists {
		return existingUser, nil
	}

	if len(pool.users) >= pool.maxCapacity {
		return nil, ErrMaxUserPoolReached
	}

	u, a, l, err := GetUserInfoFromDb(userName)
	if err != nil {
		return nil, err
	}

	newUser := &UserInfo{
		User:        u,
		Account:     a,
		LockAccount: l,
		LastActive:  time.Now(),
	}

	pool.users[userName] = newUser
	btlLog.CACC.Info("用户 %s 已被添加到用户池\n,%d/%d", userName, len(pool.users), pool.maxCapacity)
	return newUser, nil
}

func GetUserNum() int {
	return len(pool.users)
}

func (pool *UserPool) ListUsers() []*UserInfo {
	pool.mutex.RLock()
	defer pool.mutex.RUnlock()

	users := make([]*UserInfo, 0, len(pool.users))
	for _, user := range pool.users {
		users = append(users, user)
	}
	return users
}

func (pool *UserPool) CleanupInactiveUsers() {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	now := time.Now()
	for userName, user := range pool.users {
		if userName == "admin" {
			continue
		}
		if now.Sub(user.LastActive) > pool.inactiveDuration {
			delete(pool.users, userName)
			btlLog.CACC.Info("用户 %s 因为不活跃被清理\n", userName)
		}
	}
}

func (pool *UserPool) StartCleanupScheduler(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			pool.CleanupInactiveUsers()
		}
	}()
}

func (pool *UserPool) UpdateUserActivity(userName string) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if user, exists := pool.users[userName]; exists {
		user.LastActive = time.Now()
		btlLog.CACC.Info("用户 %s 的活跃时间已更新为 %s\n", userName, user.LastActive)
	} else {
		btlLog.CACC.Warning("警告: 用户 %s 不存在，无法更新活动时间\n", userName)
	}
}

func GetUserInfoFromDb(username string) (*models.User, *models.Account, *cModels.LockAccount, error) {

	user, err := btldb.ReadUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			return nil, nil, nil, ErrUserNotFound
		}
		return nil, nil, nil, fmt.Errorf("%w: %w", models.ReadDbErr, err)
	}
	if user.Status != 0 {
		return nil, nil, nil, ErrUserLocked
	}

	account := &models.Account{}
	account, err = GetAccountByUserName(username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, nil, fmt.Errorf("%w: %w", models.ReadDbErr, err)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {

		account, err = CreateAccount(user, models.NormalAccount)
		if err != nil {
			return nil, nil, nil, CustodyAccountCreateErr
		}
	}

	lockAccount := &cModels.LockAccount{}
	lockAccount, err = GetLockAccountByUserName(username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, nil, fmt.Errorf("%w: %w", models.ReadDbErr, err)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {

		lockAccount, err = CreateLockAccount(user)
		if err != nil {
			return nil, nil, nil, CustodyAccountCreateErr
		}
	}

	return user, account, lockAccount, nil
}

func GetLockedUser(username string) (*UserInfo, error) {

	user, err := btldb.ReadUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("用户 %s 不存在", username)
		}
		return nil, fmt.Errorf("%w: %w", models.ReadDbErr, err)
	}
	if user.Status == 0 {
		return nil, fmt.Errorf("用户 %s 未被冻结", username)
	}
	userInfo := UserInfo{
		User: user,
	}

	account := &models.Account{}
	account, err = GetAccountByUserName(username)
	if err != nil {
		userInfo.Account = nil
	} else {
		userInfo.Account = account
	}

	lockAccount := &cModels.LockAccount{}
	lockAccount, err = GetLockAccountByUserName(username)
	if err != nil {
		userInfo.LockAccount = nil
	} else {
		userInfo.LockAccount = lockAccount
	}
	return &userInfo, nil
}
