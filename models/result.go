package models

import (
	"encoding/json"
	"errors"
)

type JsonResult struct {
	Success bool    `json:"success"`
	Error   string  `json:"error"`
	Code    ErrCode `json:"code"`
	Data    any     `json:"data"`
}

type Result2 struct {
	Errno  ErrCode `json:"errno"`
	ErrMsg string  `json:"errmsg"`
	Data   any     `json:"data"`
}

type ErrCode int

var (
	SuccessErr = SUCCESS.Error()
)

const (
	SUCCESS    ErrCode = 200
	DefaultErr ErrCode = -1
	ReadDbErr  ErrCode = 4001
	SUCCESS_2  ErrCode = 0
)

func (e ErrCode) Code() int {
	return int(e)
}

func (e ErrCode) ToCode() Code {
	return Code(e)
}

const (
	NameToIdErr ErrCode = iota + 500
	IdAtoiErr
	ShouldBindJsonErr
	SyncAssetIssuanceErr
	GetAssetInfoErr
	IsIdoParticipateTimeRightErr
	IsNotRightTime
	IdoIsNotPublished
	GetAllIdoPublishInfosErr
	GetOwnIdoPublishInfoErr
	GetOwnIdoParticipateInfoErr
	GetIdoParticipateInfoErr
	GetIdoParticipateInfosByAssetIdErr
	GetIdoPublishedInfosErr
	ProcessIdoPublishInfoErr
	ProcessIdoParticipateInfoErr
	SetIdoPublishInfoErr
	SetIdoParticipateInfoErr
	GetBtcBalanceByUsernameErr
	CreateOrUpdateBtcBalanceErr
	ProcessAssetTransferErr
	CreateAssetTransferErr
	GetAssetTransfersByUserIdErr
	GetAddressByOutpointErr
	GetAddressesByOutpointSliceErr
	DecodeRawTransactionSliceErr
	DecodeRawTransactionErr
	GetRawTransactionsByTxidsErr
	GenerateBlocksErr
	FaucetTransferBtcErr
	CreateAssetTransferProcessedErr
	GetAssetTransferProcessedSliceByUserIdErr
	GetAssetTransferCombinedSliceByUserIdErr
	CreateOrUpdateAssetTransferProcessedInputSliceErr
	CreateOrUpdateAssetTransferProcessedOutputSliceErr
	GetAddrReceiveEventsByUserIdErr
	CreateAddrReceiveEventsErr
	GetAddrReceiveEventsProcessedOriginByUserIdErr
	CreateOrUpdateBatchTransferErr
	GetBatchTransfersByUserIdErr
	CreateOrUpdateBatchTransfersErr
	GetAssetAddrsByUserIdErr
	CreateOrUpdateAssetAddrErr
	GetAssetLocksByUserIdErr
	CreateOrUpdateAssetLockErr
	GetAssetBalancesByUserIdErr
	CreateOrUpdateAssetBalanceErr
	CreateOrUpdateAssetBalancesErr
	GetAssetTransferCombinedSliceByAssetIdErr
	GetAssetAddrsByScriptKeyErr
	GetAssetBalancesByUserIdNonZeroErr
	GetAssetHolderNumberAssetBalanceErr
	GetAssetIdAndBalancesByAssetIdErr
	GetTimeByOutpointErr
	GetTimesByOutpointSliceErr
	ValidateAndGetProofFilePathErr
	IsLimitAndOffsetValidErr
	GetAssetBalanceByAssetIdNonZeroLengthErr
	GetAllUsernameAssetBalancesErr
	GetAssetAddrsByEncodedErr
	GetAssetBurnsByUserIdErr
	CreateAssetBurnErr
	UpdateUsernameByUserIdAllErr
	GetAssetBurnTotalErr
	GetAllUsernameAssetBalanceSimplifiedErr
	GetAllAssetAddrsErr
	GetAllAssetAddrSimplifiedErr
	GetAllAssetIdAndBalanceSimplifiedErr
	GetAllAssetIdAndBatchTransfersErr
	GetAllAddrReceiveSimplifiedErr
	GetAllAssetIdAndAddrReceiveSimplifiedErr
	GetAllAssetTransferCombinedSliceErr
	GetAllAssetTransferSimplifiedErr
	GetAllAssetIdAndAssetTransferCombinedSliceSimplifiedErr
	GetAssetLocalMintsByUserIdErr
	GetAssetLocalMintByAssetIdErr
	SetAssetLocalMintErr
	SetAssetLocalMintsErr
	GetAllAssetLocalMintSimplifiedErr
	UpdateUserIpByClientIpErr
	GetAllUserSimplifiedErr
	GetAllAssetBurnSimplifiedErr
	GetAssetRecommendsByUserIdErr
	GetAssetRecommendByAssetIdErr
	SetAssetRecommendErr
	GetAllAssetRecommendSimplifiedErr
	GetAssetRecommendByUserIdAndAssetIdErr
	GetAllFairLaunchInfosErr
	FairLaunchInfoIdInvalidErr
	GetFairLaunchInfoErr
	FairLaunchMintedInfoIdInvalidErr
	GetFairLaunchMintedInfosByFairLaunchIdErr
	ProcessFairLaunchInfoErr
	SetFairLaunchInfoErr
	IsFairLaunchMintTimeRightErr
	IsTimeRightErr
	IsFairLaunchIssued
	ProcessFairLaunchMintedInfoErr
	SetFairLaunchMintedInfoErr
	GetInventoryCouldBeMintedByFairLaunchInfoIdErr
	UpdateAndCalculateGasFeeRateByMempoolErr
	GetNumberAndAmountOfInventoryCouldBeMintedErr
	GetFairLaunchInfoByAssetIdErr
	InvalidUserIdErr
	SendFairLaunchReservedErr
	UpdateFairLaunchInfoIsReservedSentErr
	GetIssuedFairLaunchInfosErr
	GetOwnFairLaunchInfosByUserIdErr
	GetOwnFairLaunchMintedInfosByUserIdErr
	GetFairLaunchInfoSimplifiedByUserIdIssuedErr
	GetClosedFairLaunchInfoErr
	GetNotStartedFairLaunchInfoErr
	GetAllUsernameAndAssetIdAssetAddrsErr
	FeeRateAtoiErr
	FeeRateInvalidErr
	GetFairLaunchFollowsByUserIdErr
	SetFollowFairLaunchInfoErr
	SetUnfollowFairLaunchInfoErr
	GetAllFairLaunchFollowSimplifiedErr
	GetFollowedFairLaunchInfoErr
	IsFairLaunchInfoIdAndAssetIdValidErr
	FairLaunchInfoAssetIdInvalidErr
	GetAssetLocalMintHistoriesByUserIdErr
	GetAssetLocalMintHistoryByAssetIdErr
	SetAssetLocalMintHistoryErr
	SetAssetLocalMintHistoriesErr
	GetAllAssetLocalMintHistorySimplifiedErr
	GetAssetManagedUtxosByUserIdErr
	GetAssetManagedUtxoByAssetIdErr
	SetAssetManagedUtxosErr
	GetAllAssetManagedUtxoSimplifiedErr
	ValidateUserIdAndAssetManagedUtxoIdsErr
	GetAmountCouldBeMintByMintedNumberErr
	CreateFairLaunchIncomeOfServerPaySendReservedFeeErr
	GetAssetBalanceByUserIdAndAssetIdErr
	GetAssetTransferByTxidErr
	FormFileErr
	DeviceIdIsNullErr
	OsGetPwdErr
	SaveUploadedFileErr
	CreateLogFileUploadErr
	FileSizeTooLargeErr
	GetAccountAssetBalanceExtendsByAssetIdErr
	BackAmountErr
	GetAllLogFilesErr
	GetFileUploadErr
	OsOpenFileErr
	IoCopyFIleErr
	GetAllAccountAssetTransfersByAssetIdErr
	RefundUserFirstMintByUsernameAndAssetIdErr
	GetAssetHolderBalancePageNumberRequestInvalidErr
	GetAssetHolderBalancePageNumberByPageSizeErr
	GetAccountAssetTransfersLimitAndOffsetErr
	GetAccountAssetTransferPageNumberByPageSizeRequestInvalidErr
	GetAccountAssetTransferPageNumberByPageSizeErr
	GetAccountAssetBalancesLimitAndOffsetErr
	GetAccountAssetBalancePageNumberByPageSizeRequestInvalidErr
	GetAccountAssetBalancePageNumberByPageSizeErr
	GetAssetManagedUtxoLimitAndOffsetErr
	GetAssetManagedUtxoPageNumberByPageSizeRequestInvalidErr
	GetAssetManagedUtxoPageNumberByPageSizeErr
	InvalidTweakedGroupKeyErr
	SetAssetGroupErr
	GetAssetGroupErr
	CreateNftTransferErr
	GetNftTransferByAssetIdErr
	PageNumberExceedsTotalNumberErr
	CreateNftPresaleErr
	CreateNftPresalesErr
	GetNftPresalesByAssetIdErr
	GetLaunchedNftPresalesErr
	GetNftPresalesByBuyerUserIdErr
	BuyNftPresaleErr
	GetNftPresaleByGroupKeyErr
	FetchAssetMetaErr
	GetUserDataErr
	GetUserDataYamlErr
	ReSetFailOrCanceledNftPresaleErr
	GetAccountAssetBalanceUserHoldTotalAmountErr
	ProcessNftPresaleBatchGroupLaunchRequestAndCreateErr
	AddWhitelistsByRequestsErr
	GetBatchGroupsErr
	GetNftPresaleByBatchGroupIdErr
	GetBlockchainInfoErr
	CreateOrUpdateAssetListsErr
	GetAssetListsByUserIdNonZeroErr
	IsAssetListRecordExistErr
	GetUserStatsYamlErr
	GetUserStatsErr
	StatsUserInfoToCsvErr
	TooManyQueryParamsErr
	DateFormatErr
	GetSpecifiedDateUserStatsErr
	InvalidQueryParamErr
	GetUserActiveRecordErr
	GetActiveUserCountBetweenErr
	GetDateLoginCountErr
	PageNumberOutOfRangeErr
	NegativeValueErr
	GetDateIpLoginRecordErr
	GetDateIpLoginRecordCountErr
	GetDateIpLoginCountErr
	GetNewUserCountErr
	GetAssetsNameErr
	DateIpLoginRecordToCsvErr
	GetNewUserRecordAllErr
	GetDateIpLoginRecordAllErr
	GetBackRewardsErr
	RedisGetVerifyErr
	RedisSetRandErr
	AssetBalanceBackupErr
	InvalidHashLengthErr
	UpdateAssetBalanceBackupErr
	GetLatestAssetBalanceHistoriesErr
	CreateAssetBalanceHistoriesErr
	GetGroupFirstImageDataErr
	ProcessPoolAddLiquidityBatchRequestErr
	PoolCreateErr
	ProcessPoolRemoveLiquidityBatchRequestErr
	ProcessPoolSwapExactTokenForTokenNoPathBatchRequestErr
	ProcessPoolSwapTokenForExactTokenNoPathBatchRequestErr
	QueryPoolInfoErr
	PoolDoesNotExistErr
	AtoiErr
	QueryShareRecordsErr
	QueryUserShareRecordsErr
	QueryShareRecordsCountErr
	QueryUserShareRecordsCountErr
	QuerySwapRecordsCountErr
	QueryUserSwapRecordsCountErr
	QuerySwapRecordsErr
	QueryUserSwapRecordsErr
	UsernameEmptyErr
	QueryUserLpAwardBalanceErr
	QueryUserWithdrawAwardRecordsCountErr
	QueryUserWithdrawAwardRecordsErr
	LimitEmptyErr
	RowsEmptyErr
	OffsetEmptyErr
	PageEmptyErr
	LimitLessThanZeroErr
	RowsLessThanZeroErr
	OffsetLessThanZeroErr
	StateLessThanZeroErr
	PageLessThanZeroErr
	UsernameNotMatchErr
	CalcAddLiquidityErr
	CalcRemoveLiquidityErr
	CalcSwapExactTokenForTokenNoPathErr
	CalcSwapTokenForExactTokenNoPathErr
	QueryAddLiquidityBatchCountErr
	QueryAddLiquidityBatchErr
	QueryRemoveLiquidityBatchCountErr
	QueryRemoveLiquidityBatchErr
	QuerySwapExactTokenForTokenNoPathBatchCountErr
	QuerySwapExactTokenForTokenNoPathBatchErr
	QuerySwapTokenForExactTokenNoPathBatchCountErr
	QuerySwapTokenForExactTokenNoPathBatchErr
	QueryWithdrawAwardBatchCountErr
	QueryWithdrawAwardBatchErr
	CalcQuoteErr
	QueryParamEmptyErr
	AddLiquidityErr
	RemoveLiquidityErr
	SwapExactTokenForTokenNoPathErr
	SwapTokenForExactTokenNoPathErr
	WithdrawAwardErr
	CalcAmountOutErr
	CalcAmountInErr
	GetPurchasedNftPresaleInfoErr
	GetPurchasedNftPresaleInfoCountErr
	GetPurchasedNftPresaleInfoLimitAndOffsetErr
	GetBtcBalanceCountErr
	GetBtcBalanceOrderLimitOffsetErr
	QueryUserShareBalanceErr
	GetNftPresaleOfflinePurchaseDataErr
	UpdateNftPresaleOfflinePurchaseDataErr
	CalcBurnLiquidityErr
	QueryUserAllShareRecordsCountErr
	QueryUserAllSwapRecordsCountErr
	QueryUserAllSwapRecordsErr
	QueryLiquidityAndAwardRecordsCountErr
	QueryLiquidityAndAwardRecordsErr
	QueryLpAwardRecordsCountErr
	QueryLpAwardRecordsErr
	SetBtcUtxoErr
	QuerySwapTrsScanCountErr
	QuerySwapTrsErr
	PoolTransferToFeeErr
	GetPoolAccountInfoErr
	ParseUintErr
	QueryPairIdErr
	AssetIdEmptyErr
	GetAssetBalanceLimitAndOffsetErr
	GetAssetBalanceCountErr
	QueryAssetBalanceInfoByUsernameErr
	GetAssetBalanceHistoryCountErr
	QueryAssetBalanceHistoryInfoByUsernameErr
	GetAssetBalanceInfoCountErr
	GetAssetBalanceInfoErr
	GetAccountAssetTransferCountErr
	GetAccountAssetTransferErr
	GetAssetManagedUtxoInfoCountErr
	GetAssetManagedUtxoInfoErr
	GetAssetTransferCombinedSliceByAssetIdLimitErr
	GetAssetLocalMintInfoCountErr
	GetAssetLocalMintInfoErr
	GetAssetLocalMintHistoryInfoCountErr
	GetAssetLocalMintHistoryInfoErr
	GetAssetsNameMapErr
	GetChannelNodeListErr
	GetChannelIdsAndPointsErr
	TradeToUserFundChannelErr
	GetLastProofErr
	TxUpdateErr
	CreateSellOrderErr
	BuySOrderErr
	PublishSOrderTxErr
	QueryPsbtTlSwapCreateInfosErr
	QueryPsbtTlSwapBoughtInfosErr
	QueryPsbtTlSwapQueryCreateInfosErr
	QueryPsbtTlSwapQueryBoughtInfosErr
	QueryPsbtTlSwapQueryCreateInfosByIdErr
	QueryPsbtTlSwapQueryBoughtInfosByIdErr
	QueryPsbtTlSwapCreateInfosCountErr
	QueryPsbtTlSwapBoughtInfosCountErr
	QueryPsbtTlSwapQueryCreateInfosCountErr
	QueryPsbtTlSwapQueryBoughtInfosCountErr
	QueryPsbtTlSwapQueryCreateInfosByIdCountErr
	QueryPsbtTlSwapQueryBoughtInfosByIdCountErr
	InvalidReq
	GetAssetHolderBalanceCountErr
	GetAssetHolderBalanceErr
	GetAccountAssetBalanceCountErr
	GetAccountAssetBalanceErr
	GetAssetManagedUtxoCountErr
	GetAssetManagedUtxoErr
	FeatureIsDisabled
	PureAddLiquidityErr
	GetConfErr
	CreateBoxDeviceInfoErr
	CreateBoxChannelsInfoErr
	GetBoxDeviceActiveInfoToH5Err
	UpdateBoxChannelInfoErr
	HandleForceClosedAssetChannelToOpenErr
	GetMcAndAliasErr
	OpenAssetChannelErr
	GetBoxDeviceAndChannelsInfoErr
	GetTradeChannelsInfoErr
	SetBackAssetsRecordErr
	GetTapAddrsErr
	GetAllBoxDeviceInfoErr
	CreateBoxAssetPushInfoErr
	UploadBoxIPErr
	GetBoxIPErr
	OsUserHomeDirErr
	InvalidFilePath
	BoxFrpErr
	QuerySecretErr
	EncodeDataToBase64Err

	AvailablePortErr
	InvalidPort

	ApiNotAllowedErr
	ValidateBoxProxyErr
	QueryPortErr
	BoxProxyCallRespErr
	BoxProxyCallRespMsgErr
	BoxProxyCallRespCodeErr

	RedisSetErr
	RedisGetErr
	PortAlreadyUsed

	PubKeyEmpty
	SecretEmpty

	GetChannelCountErr
	GetTotalCapacityErr
	GetNodeCountErr
	GetChannelInfoErr
	SearchChannelInfoErr

	CreateDeliAddrErr
	ReadDeliAddrErr
	UpdateDeliAddrErr
	DeleteDeliAddrErr

	GetLoDataHistoryErr

	CreateDeliAddrRecInfoErr

	GetTimesBatchProcessSliceErr
	ProcessChannelInfoErr

	GetCountriesErr
	GetCountriesEnErr

	CreateContactInfoErr
)

const (
	_ ErrCode = iota + 1000
	_
	_
	_

	CustodyAccountPayInsideMissionFaild
	CustodyAccountPayInsideMissionPending
)

func (e ErrCode) Error() string {
	switch {
	case errors.Is(e, SUCCESS):
		return ""
	case errors.Is(e, SUCCESS_2):
		return ""
	case errors.Is(e, DefaultErr):
		return "error"
	case errors.Is(e, CustodyAccountPayInsideMissionFaild):
		return "custody account pay inside mission faild"
	case errors.Is(e, CustodyAccountPayInsideMissionPending):
		return "custody account pay inside mission pending"
	case errors.Is(e, ReadDbErr):
		return "get server data error"

	default:
		return ""
	}
}

func MakeJsonErrorResult(code ErrCode, errorString string, data any) string {
	jsr := JsonResult{
		Error: errorString,
		Code:  code,
		Data:  data,
	}
	if errors.Is(code, SUCCESS) {
		jsr.Success = true
	} else {
		jsr.Success = false
	}
	jstr, err := json.Marshal(jsr)
	if err != nil {
		return MakeJsonErrorResult(DefaultErr, err.Error(), nil)
	}
	return string(jstr)
}
func MakeJsonErrorResultForHttp(code ErrCode, errorString string, data any) *JsonResult {
	jsr := JsonResult{
		Error: errorString,
		Code:  code,
		Data:  data,
	}
	if errors.Is(code, SUCCESS) {
		jsr.Success = true
	} else {
		jsr.Success = false
	}
	return &jsr
}

type Code int

var (
	NullStr = ""
)

func ToCode(c ErrCode) Code {
	if errors.Is(c, SUCCESS) {
		return 0
	}
	return Code(c)
}

func (c Code) Int() int {
	return int(c)
}

type Resp struct {
	Code Code   `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type RespT[T any] struct {
	Code Code   `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type (
	RespStr   RespT[string]
	RespInt   RespT[int]
	RespInt64 RespT[int64]
)

type RespLnc[T any] struct {
	Code Code    `json:"code"`
	Msg  string  `json:"msg"`
	Data LncT[T] `json:"data"`
}

type Lnc struct {
	List  []any `json:"list"`
	Count int64 `json:"count"`
}

type LncT[T any] struct {
	List  []T   `json:"list"`
	Count int64 `json:"count"`
}
