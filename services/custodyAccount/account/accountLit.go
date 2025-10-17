package account

import (
	"errors"
	"sync"
	"trade/btlLog"
	"trade/models"
	"trade/services/btldb"
)

type AccError error

var (
	CustodyAccountCreateErr AccError = errors.New("创建托管账户失败")
	CustodyAccountGetErr    AccError = errors.New("获取托管账户失败")
)

var CMutex sync.Mutex

func CreateAccount(user *models.User, accounttype models.AccountType) (*models.Account, error) {

	var accountModel models.Account
	accountModel.UserName = user.Username
	accountModel.UserId = user.ID
	accountModel.Type = accounttype
	accountModel.Status = models.AccountStatusEnable

	CMutex.Lock()
	defer CMutex.Unlock()
	err := btldb.CreateAccount(&accountModel)
	if err != nil {
		btlLog.CACC.Error(err.Error())
		return nil, err
	}

	return &accountModel, nil
}

func GetAccountByUserName(username string) (*models.Account, error) {
	return btldb.ReadAccountByName(username)
}
