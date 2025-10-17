package services

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/lightninglabs/taproot-assets/taprpc"
	"github.com/lightninglabs/taproot-assets/taprpc/universerpc"
	"sort"
	"strconv"
	"strings"
	"time"
	"trade/api"
	"trade/btlLog"
	"trade/config"
	"trade/middleware"
	"trade/models"
	"trade/services/btldb"
	"trade/services/custodyAccount/defaultAccount/custodyFee"
	"trade/utils"
)

func CreateNftPresale(nftPresale *models.NftPresale) error {
	return btldb.CreateNftPresale(nftPresale)
}

func CreateNftPresales(nftPresales *[]models.NftPresale) error {
	return btldb.CreateNftPresales(nftPresales)
}

func ReadNftPresale(id uint) (*models.NftPresale, error) {
	return btldb.ReadNftPresale(id)
}

func ReadNftPresaleByAssetId(assetId string) (*models.NftPresale, error) {
	return btldb.ReadNftPresaleByAssetId(assetId)
}

func ReadNftPresaleByBatchGroupId(batchGroupId int) (*[]models.NftPresale, error) {
	return btldb.ReadNftPresaleByBatchGroupId(batchGroupId)
}

func ReadNftPresaleByGroupKeyPurchasable(groupKey string) (*[]models.NftPresale, error) {
	return btldb.ReadNftPresaleByGroupKeyPurchasable(groupKey)
}

func ReadNftPresaleByGroupKeyLikePurchasable(groupKeyPart string) (*[]models.NftPresale, error) {
	return btldb.ReadNftPresaleByGroupKeyLikePurchasable(groupKeyPart)
}

func ReadAllNftPresales() (*[]models.NftPresale, error) {
	return btldb.ReadAllNftPresales()
}

func ReadAllNftPresalesOnlyGroupKeyPurchasable() (*[]models.NftPresale, error) {
	return btldb.ReadAllNftPresalesOnlyGroupKeyPurchasable()
}

func ReadNftPresalesByNftPresaleState(nftPresaleState models.NftPresaleState) (*[]models.NftPresale, error) {
	return btldb.ReadNftPresalesByNftPresaleState(nftPresaleState)
}

func ReadNftPresalesBetweenNftPresaleState(stateStart models.NftPresaleState, stateEnd models.NftPresaleState) (*[]models.NftPresale, error) {
	return btldb.ReadNftPresalesBetweenNftPresaleState(stateStart, stateEnd)
}

func ReadNftPresalesByBuyerUserId(userId int) (*[]models.NftPresale, error) {
	return btldb.ReadNftPresalesByBuyerUserId(userId)
}

func UpdateNftPresale(nftPresale *models.NftPresale) error {
	return btldb.UpdateNftPresale(nftPresale)
}

func UpdateNftPresales(nftPresales *[]models.NftPresale) error {
	return btldb.UpdateNftPresales(nftPresales)
}

func CreateAndUpdateNftPresales(newNftPresales *[]models.NftPresale, nftPresales *[]models.NftPresale) error {
	return btldb.CreateAndUpdateNftPresales(newNftPresales, nftPresales)
}

func DeleteNftPresale(id uint) error {
	return btldb.DeleteNftPresale(id)
}

func ProcessNftPresale(nftPresaleSetRequest *models.NftPresaleSetRequest) *models.NftPresale {
	var assetId string
	assetId = nftPresaleSetRequest.AssetId
	if assetId == "" {
		btlLog.PreSale.Error("nftPresaleSetRequest.AssetId(" + assetId + ") is null")
		return nil
	}
	var name string
	var assetType string
	var groupKey string
	var amount int
	var meta string
	_asset, err := api.GetIncludeLeasedAssetById(assetId)

	if err != nil {

		btlLog.PreSale.Error("api GetIncludeLeasedAssetById err:%v", err)
	} else {
		name = _asset.AssetGenesis.Name
		assetType = _asset.AssetGenesis.AssetType.String()
		if _asset.AssetGroup != nil {
			groupKey = hex.EncodeToString(_asset.AssetGroup.TweakedGroupKey)
		}
		amount = int(_asset.Amount)

		assetMeta, err := api.FetchAssetMetaByAssetId(assetId)
		if err != nil {
			btlLog.PreSale.Error("api FetchAssetMetaByAssetId err:%v", err)
		}
		meta = assetMeta.Data
	}
	groupKeyByAssetId, err := api.GetGroupKeyByAssetId(assetId)
	if err != nil {
		btlLog.PreSale.Error("api GetGroupKeyByAssetId err:%v", err)
	} else {
		groupKey = groupKeyByAssetId
	}
	return &models.NftPresale{
		AssetId:    assetId,
		Name:       name,
		AssetType:  assetType,
		Meta:       meta,
		GroupKey:   groupKey,
		Amount:     amount,
		Price:      nftPresaleSetRequest.Price,
		LaunchTime: utils.GetTimestamp(),
		State:      models.NftPresaleStateLaunched,
	}
}

func ProcessNftPresales(nftPresaleSetRequests *[]models.NftPresaleSetRequest) *[]models.NftPresale {
	if nftPresaleSetRequests == nil {
		return nil
	}
	var nftPresales []models.NftPresale
	for _, nftPresaleSetRequest := range *nftPresaleSetRequests {
		nftPresales = append(nftPresales, *(ProcessNftPresale(&nftPresaleSetRequest)))
	}
	return &nftPresales
}

func GetNftPresaleByAssetId(assetId string) (*models.NftPresale, error) {
	return ReadNftPresaleByAssetId(assetId)
}

func GetNftPresaleByBatchGroupId(batchGroupId int) (*[]models.NftPresale, error) {
	return ReadNftPresaleByBatchGroupId(batchGroupId)
}

func GetNftPresaleByGroupKeyPurchasable(groupKey string) (*[]models.NftPresale, error) {
	var err error
	if len(groupKey) == 0 {
		err = errors.New("group_key is null string(" + groupKey + ")")
		return nil, err
	} else if len(groupKey) < 64 {
		return ReadNftPresaleByGroupKeyLikePurchasable(groupKey)
	} else if len(groupKey) == 64 || len(groupKey) == 66 {
		return ReadNftPresaleByGroupKeyPurchasable(groupKey)
	} else {
		return ReadNftPresaleByGroupKeyPurchasable(groupKey)
	}
}

func GetNftPresaleByGroupKeyLikePurchasable(groupKeyPart string) (*[]models.NftPresale, error) {
	return ReadNftPresaleByGroupKeyLikePurchasable(groupKeyPart)
}

func GetNftPresaleNoGroupKeyPurchasable() (*[]models.NftPresale, error) {
	return ReadNftPresaleByGroupKeyPurchasable("")
}

func GetNftPresalesByBuyerUserId(userId int) (*[]models.NftPresale, error) {
	return btldb.ReadNftPresalesByBuyerUserId(userId)
}

func GetNftPresalesByBuyerUsername(username string) (*[]models.NftPresale, error) {
	return btldb.ReadNftPresalesByBuyerUsername(username)
}

func GetAllNftPresalesOnlyGroupKeyPurchasable() (*[]models.NftPresale, error) {
	return btldb.ReadAllNftPresalesOnlyGroupKeyPurchasable()
}

type GroupKeyAndGroupName struct {
	GroupKey  string `json:"group_key"`
	GroupName string `json:"group_name"`
}

func GetAllNftPresaleGroupKeyPurchasable() (*[]GroupKeyAndGroupName, error) {
	presales, err := GetAllNftPresalesOnlyGroupKeyPurchasable()
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "GetAllNftPresalesOnlyGroupKeyPurchasable")
	}
	groupKeyMap := make(map[string]bool)
	for _, presale := range *presales {
		if presale.GroupKey != "" {
			groupKeyMap[presale.GroupKey] = true
		}
	}
	var groupKeys []string
	for key := range groupKeyMap {
		groupKeys = append(groupKeys, key)
	}
	var groupKeyAndGroupNames []GroupKeyAndGroupName
	var groupKeyMapName *map[string]string
	network, err := api.NetworkStringToNetwork(config.GetLoadConfig().NetWork)
	if err != nil {
		btlLog.PreSale.Error("api NetworkStringToNetwork err:%v", err)
	} else {
		groupKeyMapName, err = GetGroupNamesByGroupKeys(network, groupKeys)
		if err != nil {
			btlLog.PreSale.Error("GetGroupNamesByGroupKeys err:%v", err)
		}
	}
	for _, groupKey := range groupKeys {
		var groupName string
		if groupKeyMapName != nil {
			groupName = (*groupKeyMapName)[groupKey]
		}
		groupKeyAndGroupNames = append(groupKeyAndGroupNames, GroupKeyAndGroupName{
			GroupKey:  groupKey,
			GroupName: groupName,
		})
	}
	return &groupKeyAndGroupNames, nil
}

func GetLaunchedNftPresales() (*[]models.NftPresale, error) {
	return ReadNftPresalesByNftPresaleState(models.NftPresaleStateLaunched)
}

func IsNftPresalePurchasable(nftPresale *models.NftPresale) bool {
	if nftPresale == nil {
		return false
	}
	return nftPresale.State == models.NftPresaleStateLaunched
}

func IsNftPresaleAddrValid(nftPresale *models.NftPresale, addr *taprpc.Addr) (bool, error) {
	var err error
	if nftPresale == nil || addr == nil {
		err = errors.New("nftPresale or addr is nil")
		return false, err
	}
	addrAssetId := hex.EncodeToString(addr.AssetId)
	if addrAssetId != nftPresale.AssetId {
		err = errors.New("addrAssetId(" + addrAssetId + ") is not equal nftPresale.AssetId(" + nftPresale.AssetId + ")")
		return false, err
	}
	addrAssetType := addr.AssetType.String()
	if strings.ToLower(addrAssetType) != strings.ToLower(nftPresale.AssetType) {
		err = errors.New("addrAssetType(" + strings.ToLower(addrAssetType) + "[ToLower]) is not equal nftPresale.AssetType(" + strings.ToLower(nftPresale.AssetType) + "[ToLower])")
		return false, err
	}
	addrAmount := addr.Amount
	if addrAmount != uint64(nftPresale.Amount) {
		err = errors.New("addrAmount(" + strconv.FormatUint(addrAmount, 10) + ") is not equal nftPresale.Amount(" + strconv.Itoa(nftPresale.Amount) + ")")
		return false, err
	}
	addrGroupKey := hex.EncodeToString(addr.GroupKey)
	var isGroupKeyEqual bool

	if len(nftPresale.GroupKey) == 64 {
		isGroupKeyEqual = strings.Contains(addrGroupKey, nftPresale.GroupKey)

	} else if len(nftPresale.GroupKey) == 66 {
		isGroupKeyEqual = addrGroupKey == nftPresale.GroupKey
	} else {
		isGroupKeyEqual = addrGroupKey == nftPresale.GroupKey
	}
	if !isGroupKeyEqual {
		err = errors.New("addrGroupKey(" + addrGroupKey + ") is not equal or contains nftPresale.GroupKey(" + nftPresale.GroupKey + ")")
		return false, err
	}
	return true, nil
}

func UpdateNftPresaleByPurchaseInfo(userId int, username string, deviceId string, addr string, scriptKey string, internalKey string, nftPresale *models.NftPresale) error {
	var err error
	if nftPresale == nil {
		err = errors.New("nftPresale is nil")
		return err
	}
	nftPresale.BuyerUserId = userId
	nftPresale.BuyerUsername = username
	nftPresale.BuyerDeviceId = deviceId
	nftPresale.ReceiveAddr = addr
	nftPresale.AddrScriptKey = scriptKey
	nftPresale.AddrInternalKey = internalKey
	nftPresale.BoughtTime = utils.GetTimestamp()
	nftPresale.State = models.NftPresaleStateBoughtNotPay
	err = UpdateNftPresale(nftPresale)
	if err != nil {
		return utils.AppendErrorInfo(err, "UpdateNftPresale")
	}
	return nil
}

func IsDuringPurchasableTime(start int, end int) bool {
	now := int(time.Now().Unix())
	if end == 0 {
		return now >= start
	}
	return now >= start && now < end
}

func IsPurchasableTimeValid(nftPresale *models.NftPresale) (bool, error) {
	if nftPresale == nil {
		return false, errors.New("nftPresale is nil")
	}
	isValid := IsDuringPurchasableTime(nftPresale.StartTime, nftPresale.EndTime)
	if !isValid {
		return false, errors.New("nftPresale StartTime(" + strconv.Itoa(nftPresale.StartTime) + ") and EndTime(" + strconv.Itoa(nftPresale.EndTime) + ") is not valid")
	}
	return isValid, nil
}

func IsWhitelistPass(nftPresale *models.NftPresale, username string) (bool, error) {
	whitelists, err := GetNftPresaleWhitelistsByNftPresale(nftPresale)
	if err != nil {
		return false, utils.AppendErrorInfo(err, "GetNftPresaleWhitelistsByNftPresale")
	}
	if whitelists == nil || len(*whitelists) == 0 {

		return true, nil
	}
	for _, user := range *whitelists {
		if user == username {
			return true, nil
		}
	}
	return false, errors.New("username(" + username + ") not found in Whitelists")
}

func BuyNftPresale(userId int, username string, buyNftPresaleRequest models.BuyNftPresaleRequest) error {
	assetId := buyNftPresaleRequest.AssetId
	nftPresale, err := GetNftPresaleByAssetId(assetId)
	if err != nil {
		return utils.AppendErrorInfo(err, "GetNftPresaleByAssetId")
	}

	if !IsNftPresalePurchasable(nftPresale) {
		err = errors.New("nft(" + nftPresale.AssetId + ") is not purchasable, its state is " + nftPresale.State.String() + ", ")
		return utils.AppendErrorInfo(err, "IsNftPresalePurchasable")
	}

	_, err = IsPurchasableTimeValid(nftPresale)
	if err != nil {
		return utils.AppendErrorInfo(err, "IsPurchasableTimeValid")
	}

	_, err = IsWhitelistPass(nftPresale, username)
	if err != nil {
		return utils.AppendErrorInfo(err, "IsWhitelistPass")
	}

	price := nftPresale.Price
	{
		accountBalance, err := custodyFee.GetAccountBalance(uint(userId))
		if err != nil {
			return utils.AppendErrorInfo(err, "GetAccountBalance")
		}
		notPayAmount := CalculateAllNotPayAmount(username)
		isEnough := accountBalance >= int64(price)+notPayAmount
		if !isEnough {
			err = errors.New("user(" + strconv.Itoa(userId) + ")'s account balance(" + strconv.FormatInt(accountBalance, 10) + ") not enough to pay nft presale price" + "(" + strconv.Itoa(nftPresale.Price) + ") and not pay amount(" + strconv.FormatInt(notPayAmount, 10) + ")")
			return utils.AppendErrorInfo(err, "IsAccountBalanceEnough")
		}
	}

	var addr, scriptKey, internalKey string

	addr = buyNftPresaleRequest.ReceiveAddr

	if addr != "" {

		decodedAddrInfo, err := api.GetDecodedAddrInfo(addr)
		if err != nil {
			return utils.AppendErrorInfo(err, "GetDecodedAddrInfo")
		}

		isNftPresaleAddrValid, err := IsNftPresaleAddrValid(nftPresale, decodedAddrInfo)
		if err != nil || !isNftPresaleAddrValid {
			return utils.AppendErrorInfo(err, "IsNftPresaleAddrValid")
		}
		scriptKey = hex.EncodeToString(decodedAddrInfo.ScriptKey)
		internalKey = hex.EncodeToString(decodedAddrInfo.InternalKey)
	}

	deviceId := buyNftPresaleRequest.DeviceId

	err = UpdateNftPresaleByPurchaseInfo(userId, username, deviceId, addr, scriptKey, internalKey, nftPresale)
	if err != nil {
		return utils.AppendErrorInfo(err, "UpdateNftPresaleByPurchaseInfo")
	}
	return nil
}

func NftPresaleToNftPresaleSimplified(nftPresale *models.NftPresale, noMeta bool, noWhitelist bool) *models.NftPresaleSimplified {
	if nftPresale == nil {
		return nil
	}
	assetId := nftPresale.AssetId
	var assetMeta *models.AssetMeta
	var err error
	if !noMeta {
		assetMeta, err = GetAssetMetaByAssetId(assetId)
		if err != nil {
			btlLog.PreSale.Error("GetAssetMetaImageDataByAssetId err:%v", err)
			assetMeta = &models.AssetMeta{}
		}
	}
	if assetMeta == nil {
		assetMeta = &models.AssetMeta{}
	}
	var whitelists *[]string
	if !noWhitelist {
		whitelists, err = GetNftPresaleWhitelistsByNftPresale(nftPresale)
		if err != nil {
			btlLog.PreSale.Error("GetNftPresaleWhitelistsByNftPresale err:%v", err)
			whitelists = &[]string{}
		}
	}
	if whitelists == nil {
		whitelists = &[]string{}
	}
	return &models.NftPresaleSimplified{
		ID:              nftPresale.ID,
		UpdatedAt:       nftPresale.UpdatedAt,
		BatchGroupId:    nftPresale.BatchGroupId,
		AssetId:         nftPresale.AssetId,
		Name:            nftPresale.Name,
		AssetType:       nftPresale.AssetType,
		Meta:            nftPresale.Meta,
		GroupKey:        nftPresale.GroupKey,
		Amount:          nftPresale.Amount,
		Price:           nftPresale.Price,
		Info:            nftPresale.Info,
		BuyerUserId:     nftPresale.BuyerUserId,
		BuyerUsername:   nftPresale.BuyerUsername,
		BuyerDeviceId:   nftPresale.BuyerDeviceId,
		ReceiveAddr:     nftPresale.ReceiveAddr,
		AddrScriptKey:   nftPresale.AddrScriptKey,
		AddrInternalKey: nftPresale.AddrInternalKey,
		PayMethod:       nftPresale.PayMethod,
		LaunchTime:      nftPresale.LaunchTime,
		StartTime:       nftPresale.StartTime,
		EndTime:         nftPresale.EndTime,
		BoughtTime:      nftPresale.BoughtTime,
		PaidId:          nftPresale.PaidId,
		PaidSuccessTime: nftPresale.PaidSuccessTime,
		SentTime:        nftPresale.SentTime,
		SentTxid:        nftPresale.SentTxid,
		SentOutpoint:    nftPresale.SentOutpoint,
		SentAddress:     nftPresale.SentAddress,
		State:           nftPresale.State,
		ProcessNumber:   nftPresale.ProcessNumber,
		IsReLaunched:    nftPresale.IsReLaunched,
		MetaStr:         (*assetMeta).AssetMeta,
		Whitelist:       whitelists,
	}
}

func NftPresaleSliceToNftPresaleSimplifiedSlice(nftPresales *[]models.NftPresale, noMeta bool, noWhitelist bool) *[]models.NftPresaleSimplified {
	if nftPresales == nil {
		return nil
	}
	var nftPresaleSimplifiedSlice []models.NftPresaleSimplified
	nftPresaleSimplifiedSlice = make([]models.NftPresaleSimplified, 0)
	for _, nftPresale := range *nftPresales {
		nftPresaleSimplifiedSlice = append(nftPresaleSimplifiedSlice, *(NftPresaleToNftPresaleSimplified(&nftPresale, noMeta, noWhitelist)))
	}
	return &nftPresaleSimplifiedSlice
}

func GetAllNftPresaleStateBoughtNotPay() (*[]models.NftPresale, error) {
	return btldb.ReadNftPresalesByNftPresaleState(models.NftPresaleStateBoughtNotPay)
}

func GetAllNftPresaleStatePaidPending() (*[]models.NftPresale, error) {
	return btldb.ReadNftPresalesByNftPresaleState(models.NftPresaleStatePaidPending)
}

func GetAllNftPresaleStatePaidNotSend() (*[]models.NftPresale, error) {
	return btldb.ReadNftPresalesByNftPresaleStateAndReceiveAddrNotNull(models.NftPresaleStatePaidNotSend)
}

func GetAllNftPresaleStateSentPending() (*[]models.NftPresale, error) {
	return btldb.ReadNftPresalesByNftPresaleState(models.NftPresaleStateSentPending)
}

func IncreaseNftPresaleProcessNumber(nftPresale *models.NftPresale) (err error) {
	nftPresale.ProcessNumber += 1
	return UpdateNftPresale(nftPresale)
}

func StorePaidIdThenChangeStateAndClearProcessNumber(paidId int, nftPresale *models.NftPresale) error {
	nftPresale.PayMethod = models.FeePaymentMethodCustodyAccount
	nftPresale.PaidId = paidId
	nftPresale.State = models.NftPresaleStatePaidPending
	nftPresale.ProcessNumber = 0
	err := UpdateNftPresale(nftPresale)
	if err != nil {
		return utils.AppendErrorInfo(err, "UpdateNftPresale")
	}
	return nil
}

func SetNftPresaleFail(nftPresale *models.NftPresale) error {
	nftPresale.State = models.NftPresaleStateFailOrCanceled
	err := UpdateNftPresale(nftPresale)
	if err != nil {
		return utils.AppendErrorInfo(err, "UpdateNftPresale")
	}
	return nil
}

func ChangeNftPresaleStateAndUpdatePaidSuccessTimeThenClearProcessNumber(nftPresale *models.NftPresale) error {
	nftPresale.State = models.NftPresaleStatePaidNotSend
	nftPresale.PaidSuccessTime = utils.GetTimestamp()
	nftPresale.ProcessNumber = 0
	err := UpdateNftPresaleAndSelfAddBatchGroupSoldNumber(nftPresale)
	if err != nil {
		return utils.AppendErrorInfo(err, "UpdateNftPresaleAndSelfAddBatchGroupSoldNumber")
	}
	return nil
}

func ChangeNftPresaleStateAndClearProcessNumber(state models.NftPresaleState, nftPresale *models.NftPresale) error {
	nftPresale.State = state
	nftPresale.ProcessNumber = 0
	err := UpdateNftPresale(nftPresale)
	if err != nil {
		return utils.AppendErrorInfo(err, "UpdateNftPresale")
	}
	return nil
}

func UpdateNftPresaleAfterSent(outpoint string, txid string, address string, nftPresale *models.NftPresale) error {
	nftPresale.SentOutpoint = outpoint
	nftPresale.SentTxid = txid
	nftPresale.SentAddress = address
	nftPresale.State = models.NftPresaleStateSentPending
	nftPresale.ProcessNumber = 0
	nftPresale.SentTime = utils.GetTimestamp()
	err := UpdateNftPresale(nftPresale)
	if err != nil {
		return utils.AppendErrorInfo(err, "UpdateNftPresale")
	}
	return nil
}

func UpdateNftPresaleBySendAssetResponse(nftPresale *models.NftPresale, sendAssetResponse *taprpc.SendAssetResponse) error {
	var err error
	scriptKey := nftPresale.AddrScriptKey
	internalKey := nftPresale.AddrInternalKey
	var outpoint string
	var txid string
	var address string
	outpoint, err = SendAssetResponseScriptKeyAndInternalKeyToOutpoint(sendAssetResponse, scriptKey, internalKey)
	if err != nil {
		btlLog.PreSale.Error("SendAssetResponseScriptKeyAndInternalKeyToOutpoint:%v", err)
	} else {
		txid, _ = utils.GetTransactionAndIndexByOutpoint(outpoint)
		address, err = api.GetListChainTransactionsOutpointAddress(outpoint)
		if err != nil {
			btlLog.PreSale.Error("GetListChainTransactionsOutpointAddress:%v", err)
		}
	}
	err = UpdateNftPresaleAfterSent(outpoint, txid, address, nftPresale)
	if err != nil {
		btlLog.PreSale.Error("UpdateNftPresaleAfterSent:%v", err)
	}
	return nil
}

func UpdateNftPresaleAfterConfirmed(nftPresale *models.NftPresale) error {
	nftPresale.State = models.NftPresaleStateSent
	nftPresale.ProcessNumber = 0
	err := UpdateNftPresale(nftPresale)
	if err != nil {
		return utils.AppendErrorInfo(err, "UpdateNftPresale")
	}
	return nil
}

func UpdateNftPresaleAndSelfAddBatchGroupSoldNumber(nftPresale *models.NftPresale) error {
	var err error
	if nftPresale == nil {
		return errors.New("nftPresale is nil")
	}
	if nftPresale.BatchGroupId == 0 {
		return errors.New("nftPresale batch group id is 0")
	}
	tx := middleware.DB.Begin()

	var batchGroup models.NftPresaleBatchGroup
	err = tx.First(&batchGroup, nftPresale.BatchGroupId).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	batchGroup.SoldNumber += 1

	err = tx.Save(&batchGroup).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Save(nftPresale).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func ProcessNftPresaleStateBoughtNotPayService(nftPresale *models.NftPresale) error {

	paidId, err := PayGasFee(nftPresale.BuyerUserId, nftPresale.Price)
	if err != nil {
		return utils.AppendErrorInfo(err, "PayGasFee for nftPresale")
	}

	err = StorePaidIdThenChangeStateAndClearProcessNumber(paidId, nftPresale)
	if err != nil {
		return utils.AppendErrorInfo(err, "StorePaidIdThenChangeStateAndClearProcessNumber")
	}
	return nil
}

func ProcessNftPresaleStatePaidPendingService(nftPresale *models.NftPresale) error {

	var isFeePaid bool
	isFeePaid, err := IsFeePaid(nftPresale.PaidId)
	if err != nil {
		if errors.Is(err, models.CustodyAccountPayInsideMissionFaild) {
			err = SetNftPresaleFail(nftPresale)
			if err != nil {
				return utils.AppendErrorInfo(err, "SetNftPresaleFail")
			}
		}
	}

	if isFeePaid {

		err = ChangeNftPresaleStateAndUpdatePaidSuccessTimeThenClearProcessNumber(nftPresale)
		if err != nil {
			return utils.AppendErrorInfo(err, "ChangeNftPresaleStateAndUpdatePaidSuccessTimeThenClearProcessNumber")
		}
		return nil
	}
	return nil
}

func ProcessNftPresaleStatePaidNotSendService(nftPresale *models.NftPresale) error {
	var err error

	minBalance := 2000
	if !IsWalletBalanceEnough(minBalance) {
		err = errors.New("lnd wallet balance is not enough(less than " + strconv.Itoa(minBalance) + ")")
		return err
	}

	if !IsAssetBalanceEnough(nftPresale.AssetId, nftPresale.Amount) {
		err = errors.New("nft presale asset(" + strconv.Itoa(int(nftPresale.ID)) + ") balance is not enough")
		return utils.AppendErrorInfo(err, "IsAssetBalanceEnough")
	}

	if !IsAssetUtxoEnough(nftPresale.AssetId, nftPresale.Amount) {
		err = errors.New("nft presale asset(" + strconv.Itoa(int(nftPresale.ID)) + ") utxo is not enough")
		return utils.AppendErrorInfo(err, "IsAssetUtxoEnough")
	}
	if nftPresale.ReceiveAddr == "" {
		err = errors.New("nft presale asset(" + strconv.Itoa(int(nftPresale.ID)) + ")'s receive addr(" + nftPresale.ReceiveAddr + ") is null")
		return err
	}

	addrs := []string{nftPresale.ReceiveAddr}

	feeRate, err := UpdateAndGetFeeRateResponseTransformed()
	if err != nil {
		return utils.AppendErrorInfo(err, "UpdateAndGetFeeRateResponseTransformed("+strconv.Itoa(int(nftPresale.ID))+")")
	}
	feeRateSatPerKw := feeRate.SatPerKw.FastestFee

	response, err := api.SendAssetAddrSliceAndGetResponse(addrs, feeRateSatPerKw)
	if err != nil {
		return utils.AppendErrorInfo(err, "SendAssetAddrSliceAndGetResponse("+strconv.Itoa(int(nftPresale.ID))+")")
	}

	err = UpdateNftPresaleBySendAssetResponse(nftPresale, response)
	if err != nil {
		return utils.AppendErrorInfo(err, "UpdateNftPresaleBySendAssetResponse("+strconv.Itoa(int(nftPresale.ID))+")")
	}
	return nil
}

func ProcessNftPresaleStateSentPendingService(nftPresale *models.NftPresale) error {
	var err error
	if nftPresale.SentOutpoint == "" {
		err = errors.New("no outpoint generated, asset of presale(" + strconv.Itoa(int(nftPresale.ID)) + ") may has not been sent")
		return err
	}

	if IsTransactionConfirmed(nftPresale.SentTxid) {

		err = UpdateNftPresaleAfterConfirmed(nftPresale)
		if err != nil {
			return utils.AppendErrorInfo(err, "UpdateNftPresaleAfterConfirmed")
		}
		return nil
	} else {
		err = errors.New("nftPresale.SentTxid(" + nftPresale.SentTxid + ") is not confirmed")
		return err
	}
}

func ProcessAllNftPresaleStateBoughtNotPayService() (*[]ProcessionResult, error) {
	var processionResults []ProcessionResult
	nftPresales, err := GetAllNftPresaleStateBoughtNotPay()
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "GetAllNftPresaleStateBoughtNotPay")
	}
	for _, nftPresale := range *nftPresales {
		{
			err = IncreaseNftPresaleProcessNumber(&nftPresale)
			if err != nil {

			}
			err = ProcessNftPresaleStateBoughtNotPayService(&nftPresale)
			if err != nil {
				processionResults = append(processionResults, ProcessionResult{
					Id: int(nftPresale.ID),
					JsonResult: models.JsonResult{
						Success: false,
						Error:   err.Error(),
						Data:    nil,
					},
				})
				continue
			} else {
				processionResults = append(processionResults, ProcessionResult{
					Id: int(nftPresale.ID),
					JsonResult: models.JsonResult{
						Success: true,
						Error:   "",
						Data:    nil,
					},
				})
			}
		}
	}
	if processionResults == nil || len(processionResults) == 0 {
		err = errors.New("procession results null")
		return nil, err
	}
	return &processionResults, nil
}

func ProcessAllNftPresaleStatePaidPendingService() (*[]ProcessionResult, error) {
	var processionResults []ProcessionResult
	nftPresales, err := GetAllNftPresaleStatePaidPending()
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "GetAllNftPresaleStatePaidPending")
	}
	for _, nftPresale := range *nftPresales {
		{
			err = IncreaseNftPresaleProcessNumber(&nftPresale)
			if err != nil {

			}
			err = ProcessNftPresaleStatePaidPendingService(&nftPresale)
			if err != nil {
				processionResults = append(processionResults, ProcessionResult{
					Id: int(nftPresale.ID),
					JsonResult: models.JsonResult{
						Success: false,
						Error:   err.Error(),
						Data:    nil,
					},
				})
				continue
			} else {
				processionResults = append(processionResults, ProcessionResult{
					Id: int(nftPresale.ID),
					JsonResult: models.JsonResult{
						Success: true,
						Error:   "",
						Data:    nil,
					},
				})
			}
		}
	}
	if processionResults == nil || len(processionResults) == 0 {
		err = errors.New("procession results null")
		return nil, err
	}
	return &processionResults, nil
}

func ProcessAllNftPresaleStatePaidNotSendService() (*[]ProcessionResult, error) {
	var processionResults []ProcessionResult
	nftPresales, err := GetAllNftPresaleStatePaidNotSend()
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "GetAllNftPresaleStatePaidNotSend")
	}
	for _, nftPresale := range *nftPresales {
		{
			err = IncreaseNftPresaleProcessNumber(&nftPresale)
			if err != nil {

			}
			err = ProcessNftPresaleStatePaidNotSendService(&nftPresale)
			if err != nil {
				processionResults = append(processionResults, ProcessionResult{
					Id: int(nftPresale.ID),
					JsonResult: models.JsonResult{
						Success: false,
						Error:   err.Error(),
						Data:    nil,
					},
				})
				continue
			} else {
				processionResults = append(processionResults, ProcessionResult{
					Id: int(nftPresale.ID),
					JsonResult: models.JsonResult{
						Success: true,
						Error:   "",
						Data:    nil,
					},
				})
			}
		}

	}
	if processionResults == nil || len(processionResults) == 0 {
		err = errors.New("procession results null")
		return nil, err
	}
	return &processionResults, nil
}

func ProcessAllNftPresaleStateSentPendingService() (*[]ProcessionResult, error) {
	var processionResults []ProcessionResult
	nftPresales, err := GetAllNftPresaleStateSentPending()
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "GetAllNftPresaleStateSentPending")
	}
	for _, nftPresale := range *nftPresales {
		{
			err = IncreaseNftPresaleProcessNumber(&nftPresale)
			if err != nil {

			}
			err = ProcessNftPresaleStateSentPendingService(&nftPresale)
			if err != nil {
				processionResults = append(processionResults, ProcessionResult{
					Id: int(nftPresale.ID),
					JsonResult: models.JsonResult{
						Success: false,
						Error:   err.Error(),
						Data:    nil,
					},
				})
				continue
			} else {
				processionResults = append(processionResults, ProcessionResult{
					Id: int(nftPresale.ID),
					JsonResult: models.JsonResult{
						Success: true,
						Error:   "",
						Data:    nil,
					},
				})
			}
		}

	}
	if processionResults == nil || len(processionResults) == 0 {
		err = errors.New("procession results null")
		return nil, err
	}
	return &processionResults, nil
}

func ProcessNftPresaleBoughtNotPay() {
	processionResult, err := ProcessAllNftPresaleStateBoughtNotPayService()
	if err != nil {
		return
	}
	if processionResult == nil || len(*processionResult) == 0 {
		return
	}

	btlLog.PreSale.Error("[PRESALE.BNP]%v", "\n"+utils.ValueJsonString(processionResult))
}

func ProcessNftPresalePaidPending() {
	processionResult, err := ProcessAllNftPresaleStatePaidPendingService()
	if err != nil {
		return
	}
	if processionResult == nil || len(*processionResult) == 0 {
		return
	}

	btlLog.PreSale.Error("[PRESALE.PPD]%v", "\n"+utils.ValueJsonString(processionResult))
}

func ProcessNftPresalePaidNotSend() {
	processionResult, err := ProcessAllNftPresaleStatePaidNotSendService()
	if err != nil {
		return
	}
	if processionResult == nil || len(*processionResult) == 0 {
		return
	}

	btlLog.PreSale.Error("[PRESALE.PNS]%v", "\n"+utils.ValueJsonString(processionResult))
}

func ProcessNftPresaleSentPending() {
	processionResult, err := ProcessAllNftPresaleStateSentPendingService()
	if err != nil {
		return
	}
	if processionResult == nil || len(*processionResult) == 0 {
		return
	}

	btlLog.PreSale.Error("[PRESALE.SPD]%v", "\n"+utils.ValueJsonString(processionResult))
}

func GetProcessingNftPresale() (*[]models.NftPresale, error) {
	return ReadNftPresalesBetweenNftPresaleState(models.NftPresaleStateBoughtNotPay, models.NftPresaleStatePaidNotSend)
}

func CheckIsNftPresaleProcessing() error {
	processingNftPresale, err := GetProcessingNftPresale()
	if err != nil {
		return err
	}
	if processingNftPresale == nil || len(*processingNftPresale) == 0 {
		return nil
	}
	err = errors.New("processing nft presale exists")
	info := fmt.Sprintf("num: %d", len(*processingNftPresale))
	return utils.AppendErrorInfo(err, info)
}

func GetGroupNameByGroupKey(network models.Network, groupKey string) (string, error) {
	var groupName string

	assetKeys, err := api.AssetLeafKeys(true, groupKey, universerpc.ProofType_PROOF_TYPE_ISSUANCE)
	if err != nil {
		return "", utils.AppendErrorInfo(err, "AssetLeafKeys")
	}
	if len(*assetKeys) == 0 {
		err = errors.New("length of assetKeys(" + strconv.Itoa(len(*assetKeys)) + ") is zero, not fount AssetLeafKey")
		if err != nil {
			return "", utils.AppendErrorInfo(err, "AssetLeafKeys")
		}
	}
	var outpoints []string
	opMapScriptKey := make(map[string]string)
	for _, assetKey := range *assetKeys {
		outpoints = append(outpoints, assetKey.OpStr)
		opMapScriptKey[assetKey.OpStr] = assetKey.ScriptKeyBytes
	}

	outpointTime, err := api.GetTimesByOutpointSlice(network, outpoints)
	if err != nil {
		return "", utils.AppendErrorInfo(err, "GetTimesByOutpointSlice")
	}
	type timeAndAssetKey struct {
		Time           int    `json:"time"`
		OpStr          string `json:"op_str"`
		ScriptKeyBytes string `json:"script_key_bytes"`
	}
	var timeAndAssetKeys []timeAndAssetKey
	for op, _time := range outpointTime {
		timeAndAssetKeys = append(timeAndAssetKeys, timeAndAssetKey{
			Time:           _time,
			OpStr:          op,
			ScriptKeyBytes: opMapScriptKey[op],
		})
	}
	if len(timeAndAssetKeys) == 0 {
		err = errors.New("length of timeAndAssetKey(" + strconv.Itoa(len(timeAndAssetKeys)) + ") is zero")
		return "", utils.AppendErrorInfo(err, "")
	}
	{

	}

	func(tak []timeAndAssetKey) {
		sort.Slice(tak, func(i, j int) bool {
			return (tak)[i].Time < (tak)[j].Time
		})
	}(timeAndAssetKeys)

	firstAssetKey := timeAndAssetKeys[0]

	assetId, err := api.QueryProofToGetAssetId(groupKey, firstAssetKey.OpStr, firstAssetKey.ScriptKeyBytes)
	if err != nil {
		return "", utils.AppendErrorInfo(err, "QueryProofToGetAssetId")
	}

	assetMeta, err := api.FetchAssetMetaByAssetId(assetId)
	if err != nil {
		return "", utils.AppendErrorInfo(err, "FetchAssetMetaByAssetId")
	}

	var meta api.Meta
	meta.GetMetaFromStr(assetMeta.Data)
	groupName = meta.GroupName
	return groupName, nil
}

func GetGroupNamesByGroupKeys(network models.Network, groupKeys []string) (*map[string]string, error) {
	var totalOutpoints []string

	groupKeyMapName := make(map[string]string)

	groupKeyMapOps := make(map[string][]string)

	opMapScriptKey := make(map[string]string)
	for _, groupKey := range groupKeys {

		assetKeys, err := api.AssetLeafKeys(true, groupKey, universerpc.ProofType_PROOF_TYPE_ISSUANCE)
		if err != nil {
			btlLog.PreSale.Error("api AssetLeafKeys err:%v", err)
		}
		if len(*assetKeys) == 0 {
			err = errors.New("length of assetKeys(" + strconv.Itoa(len(*assetKeys)) + ") is zero, not fount AssetLeafKey")
			if err != nil {
				btlLog.PreSale.Error("%v", err)
			}
		}
		var outpoints []string
		for _, assetKey := range *assetKeys {
			outpoints = append(outpoints, assetKey.OpStr)
			opMapScriptKey[assetKey.OpStr] = assetKey.ScriptKeyBytes
		}
		totalOutpoints = append(totalOutpoints, outpoints...)
		groupKeyMapOps[groupKey] = outpoints
	}

	outpointTime, err := api.GetTimesByOutpointSlice(network, totalOutpoints)
	if err != nil {
		btlLog.PreSale.Error("api GetTimesByOutpointSlice err:%v", err)
	}
	type timeAndAssetKey struct {
		Time           int    `json:"time"`
		OpStr          string `json:"op_str"`
		ScriptKeyBytes string `json:"script_key_bytes"`
	}
	for _, groupKey := range groupKeys {
		var timeAndAssetKeys []timeAndAssetKey
		ops := groupKeyMapOps[groupKey]
		for _, op := range ops {
			timeAndAssetKeys = append(timeAndAssetKeys, timeAndAssetKey{
				Time:           outpointTime[op],
				OpStr:          op,
				ScriptKeyBytes: opMapScriptKey[op],
			})
		}
		if len(timeAndAssetKeys) == 0 {
			err = errors.New("length of timeAndAssetKey(" + strconv.Itoa(len(timeAndAssetKeys)) + ") is zero")
			btlLog.PreSale.Error("%v", err)
		}
		if len(totalOutpoints) != len(outpointTime) {
			err = errors.New("length of outpoints(" + strconv.Itoa(len(totalOutpoints)) + ") is not equal length of outpointTime(" + strconv.Itoa(len(outpointTime)) + ")")
			btlLog.PreSale.Error("%v", err)
		}

		func(tak []timeAndAssetKey) {
			sort.Slice(tak, func(i, j int) bool {
				return (tak)[i].Time < (tak)[j].Time
			})
		}(timeAndAssetKeys)

		firstAssetKey := timeAndAssetKeys[0]

		assetId, err := api.QueryProofToGetAssetId(groupKey, firstAssetKey.OpStr, firstAssetKey.ScriptKeyBytes)
		if err != nil {
			btlLog.PreSale.Error("api QueryProofToGetAssetId err:%v", err)
		}

		assetMeta, err := api.FetchAssetMetaByAssetId(assetId)
		if err != nil {
			btlLog.PreSale.Error("api FetchAssetMetaByAssetId err:%v", err)
		}

		var meta api.Meta
		meta.GetMetaFromStr(assetMeta.Data)
		groupKeyMapName[groupKey] = meta.GroupName
	}
	return &groupKeyMapName, nil
}

func GetFailOrCanceledNftPresale() (*[]models.NftPresale, error) {
	return btldb.ReadFailOrCanceledNftPresale()
}

func ProcessFailOrCanceledNftPresale(nftPresale models.NftPresale) models.NftPresale {
	return models.NftPresale{
		AssetId:    nftPresale.AssetId,
		Name:       nftPresale.Name,
		AssetType:  nftPresale.AssetType,
		Meta:       nftPresale.Meta,
		GroupKey:   nftPresale.GroupKey,
		Amount:     nftPresale.Amount,
		Price:      nftPresale.Price,
		Info:       nftPresale.Info,
		LaunchTime: nftPresale.LaunchTime,
		State:      models.NftPresaleStateLaunched,
	}
}

func ProcessFailOrCanceledNftPresales(nftPresales *[]models.NftPresale) *[]models.NftPresale {
	if *nftPresales == nil {
		return nil
	}
	var newNftPresales []models.NftPresale
	for _, nftPresale := range *nftPresales {
		newNftPresales = append(newNftPresales, ProcessFailOrCanceledNftPresale(nftPresale))
	}
	return &newNftPresales
}

func ReSetFailOrCanceledNftPresale() error {
	nftPresales, err := GetFailOrCanceledNftPresale()
	if err != nil {
		return utils.AppendErrorInfo(err, "ReadNftPresalesByNftPresaleState")
	}
	for i := range *nftPresales {
		(*nftPresales)[i].IsReLaunched = true
	}
	newNftPresales := ProcessFailOrCanceledNftPresales(nftPresales)
	return CreateAndUpdateNftPresales(newNftPresales, nftPresales)
}

func GetFirstAssetIdByBatchGroupId(batchGroupId int) (string, error) {
	var assetId string

	err := middleware.DB.Model(&models.NftPresale{}).
		Select("asset_id").
		Where("batch_group_id = ?", batchGroupId).
		Order("id ASC").
		First(&assetId).
		Error
	if err != nil {
		return "", err
	}
	return assetId, nil
}

func IsUserBoughtNftPresale(username string) (bool, error) {
	var count int64
	err := middleware.DB.Table("nft_presales").
		Where("buyer_username = ?", username).
		Count(&count).
		Error
	if err != nil {
		return false, utils.AppendErrorInfo(err, "select nft_presales count")
	}
	return count > 0, nil
}

func CalculateNftPresaleNotPayAmount(username string) (notPayAmount int64, err error) {
	err = middleware.DB.Table("nft_presales").
		Select("sum(price)").
		Where("buyer_username = ? and state = ?", username, models.NftPresaleStateBoughtNotPay).
		Scan(&notPayAmount).
		Error
	return notPayAmount, err
}

func CalculateAllNotPayAmount(username string) (notPayAmount int64) {
	nftPresaleNotPayAmount, _ := CalculateNftPresaleNotPayAmount(username)
	fairLaunchInfoNotPayAmount, _ := CalculateFairLaunchInfoNotPayAmount(username)
	fairLaunchMintedInfoNotPayAmount, _ := CalculateFairLaunchMintedInfoNotPayAmount(username)
	return nftPresaleNotPayAmount + fairLaunchInfoNotPayAmount + fairLaunchMintedInfoNotPayAmount
}
