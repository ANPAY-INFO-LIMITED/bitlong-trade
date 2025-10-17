package btlLog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
	"trade/utils"
)

type LogLevel int

const (
	ERROR LogLevel = iota
	WARNING
	DEBUG
	INFO
)

func InitBtlLog() error {
	if err := openLogFile(); err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	loadDefaultLog()
	return nil
}

type ServicesLogger struct {
	logger      *log.Logger
	errorLogger *log.Logger
	level       LogLevel
}

func NewLogger(logName string, level LogLevel, otherErrorWriter io.Writer, hasStdout bool, Writer ...io.Writer) *ServicesLogger {
	var multiWriter io.Writer

	if hasStdout {
		multiWriter = io.MultiWriter(os.Stdout)
	}
	for i := range Writer {
		if multiWriter == nil {
			multiWriter = io.MultiWriter(Writer[i])
			continue
		}
		multiWriter = io.MultiWriter(multiWriter, Writer[i])
	}
	logger := ServicesLogger{
		logger: log.New(multiWriter, "["+logName+"]: ", log.Ldate|log.Ltime),
		level:  level,
	}

	if otherErrorWriter != nil {
		multiWriter = io.MultiWriter(multiWriter, otherErrorWriter)
	}
	logger.errorLogger = log.New(io.MultiWriter(multiWriter, defaultErrorLogFile), "["+logName+"]: ", log.Ldate|log.Ltime)

	return &logger
}

func (ml *ServicesLogger) Debug(format string, v ...interface{}) {
	if ml.level >= DEBUG {
		_, callerFile, callerLine, _ := runtime.Caller(1)
		msg := fmt.Sprintf(format, v...)
		ml.logger.Printf(" %s：%d [Debug]: %s\n", callerFile, callerLine, msg)
	}
}

func (ml *ServicesLogger) Info(format string, v ...any) {
	if ml.level >= INFO {
		_, callerFile, callerLine, _ := runtime.Caller(1)
		msg := fmt.Sprintf(format, v...)
		ml.logger.Printf(" %s：%d [Log]: %s\n", callerFile, callerLine, msg)
	}
}

func (ml *ServicesLogger) Warning(format string, v ...interface{}) {
	if ml.level >= WARNING {
		_, callerFile, callerLine, _ := runtime.Caller(1)
		msg := fmt.Sprintf(format, v...)
		ml.logger.Printf(" %s：%d [Warning]: %s\n", callerFile, callerLine, msg)
	}
}

func (ml *ServicesLogger) Error(format string, v ...any) {
	if ml.level >= ERROR {
		_, callerFile, callerLine, _ := runtime.Caller(1)
		msg := fmt.Sprintf(format, v...)
		ml.errorLogger.Printf(" %s：%d [Error]: %s\n", callerFile, callerLine, msg)
	}
}

var (
	defaultLogFile      *os.File
	defaultErrorLogFile *os.File

	startLogFile *os.File

	custErrorLogFile *os.File
	CaccountLogFile  *os.File
	CLimitLogFile    *os.File
	CSwapLogFile     *os.File

	fwdtLogFile *os.File

	presaleLogFile     *os.File
	mintNftFile        *os.File
	userDataLogFile    *os.File
	userStatsLogFile   *os.File
	cpAmmLogFile       *os.File
	dateIpLoginLogFile *os.File
	pushQueueLogFile   *os.File

	fairLaunchLogFile                  *os.File
	sendFairLaunchMintedAssetLogFile   *os.File
	feeLogFile                         *os.File
	mintNftLogFile                     *os.File
	scheduledTaskLogFile               *os.File
	poolPairTokenAccountBalanceLogFile *os.File
	openChannelLogFile                 *os.File
	psbtTlSwapLogFile                  *os.File
	PoolPureLogFile                    *os.File
	boxAssetPushLogFile                *os.File
	boxAssetPushTaskLogFile            *os.File
	boxDeviceBackupLogFile             *os.File
	boxDeviceLogFile                   *os.File
	boxChannelInfosLogFile             *os.File

	lntLogFile *os.File
)

func getLogFile(dirPath string, fileName string) (*os.File, error) {
	filePath := filepath.Join(dirPath, fileName)
	backupLogFile(filePath)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func openLogFile() error {
	var err error
	dirPath := "./logs"
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {

		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return err
		}
		fmt.Println("目录已创建:", dirPath)
	}

	defaultLogFile, err = getLogFile(dirPath, "output.log")
	if err != nil {
		return err
	}
	defaultErrorLogFile, err = getLogFile(dirPath, "error.log")
	if err != nil {
		return err
	}

	startLogFile, err = getLogFile(dirPath, "start.log")
	if err != nil {
		return err
	}

	custErrorLogFile, err = getLogFile(dirPath, "cust_error.log")
	if err != nil {
		return err
	}
	CaccountLogFile, err = getLogFile(dirPath, "caccount.log")
	if err != nil {
		return err
	}
	CLimitLogFile, err = getLogFile(dirPath, "cLimit.log")
	if err != nil {
		return err
	}
	CSwapLogFile, err = getLogFile(dirPath, "cSwap.log")
	if err != nil {
		return err
	}

	fwdtLogFile, err = getLogFile(dirPath, "fwdt.log")
	if err != nil {
		return err
	}

	presaleLogFile, err = utils.GetLogFile("./logs/trade.presale.log")
	if err != nil {
		return err
	}
	mintNftFile, err = utils.GetLogFile("./logs/trade.mint_nft.log")
	if err != nil {
		return err
	}
	userDataLogFile, err = utils.GetLogFile("./logs/trade.userdata.log")
	if err != nil {
		return err
	}
	userStatsLogFile, err = utils.GetLogFile("./logs/trade.user_stats.log")
	if err != nil {
		return err
	}
	cpAmmLogFile, err = utils.GetLogFile("./logs/trade.cp_amm.log")
	if err != nil {
		return err
	}
	dateIpLoginLogFile, err = utils.GetLogFile("./logs/trade.date_ip_login.log")
	if err != nil {
		return err
	}
	pushQueueLogFile, err = utils.GetLogFile("./logs/trade.push_queue.log")
	if err != nil {
		return err
	}
	fairLaunchLogFile, err = utils.GetLogFile("./logs/trade.fair_launch.log")
	if err != nil {
		return err
	}
	sendFairLaunchMintedAssetLogFile, err = utils.GetLogFile("./logs/trade.send_fair_launch_minted_asset.log")
	if err != nil {
		return err
	}
	feeLogFile, err = utils.GetLogFile("./logs/trade.fee.log")
	if err != nil {
		return err
	}
	mintNftLogFile, err = utils.GetLogFile("./logs/trade.mint_nft.log")
	if err != nil {
		return err
	}
	scheduledTaskLogFile, err = utils.GetLogFile("./logs/trade.scheduled_task.log")
	if err != nil {
		return err
	}
	poolPairTokenAccountBalanceLogFile, err = utils.GetLogFile("./logs/trade.pool_pair_token_account_balance.log")
	if err != nil {
		return err
	}
	openChannelLogFile, err = utils.GetLogFile("./logs/trade.open_channel.log")
	if err != nil {
		return err
	}
	psbtTlSwapLogFile, err = utils.GetLogFile("./logs/trade.psbt_tl_swap.log")
	if err != nil {
		return err
	}
	PoolPureLogFile, err = utils.GetLogFile("./logs/trade.pool.pure.log")
	if err != nil {
		return err
	}
	boxAssetPushLogFile, err = utils.GetLogFile("./logs/trade.box_asset_push.log")
	if err != nil {
		return err
	}
	boxAssetPushTaskLogFile, err = utils.GetLogFile("./logs/trade.box_asset_push_task.log")
	if err != nil {
		return err
	}

	boxDeviceBackupLogFile, err = utils.GetLogFile("./logs/trade.box_device_backup.log")
	if err != nil {
		return err
	}
	boxDeviceLogFile, err = utils.GetLogFile("./logs/trade.box_device.log")
	if err != nil {
		return err
	}
	boxChannelInfosLogFile, err = utils.GetLogFile("./logs/trade.box_channel_infos.log")
	if err != nil {
		return err
	}

	lntLogFile, err = utils.GetLogFile("./logs/trade.lnt.log")
	if err != nil {
		return err
	}
	return nil
}

func backupLogFile(filePath string) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return
	}
	if fileInfo.Size() > 20*1024*1024 {
		newName := filePath + "." + time.Now().Format("200601021504") + ".bak"
		err := os.Rename(filePath, newName)
		if err != nil {
			fmt.Printf("Backup file failed: %v", err)
		}
	}
}

var (
	START                       *ServicesLogger
	CUST                        *ServicesLogger
	CACC                        *ServicesLogger
	CLMT                        *ServicesLogger
	CSWAP                       *ServicesLogger
	SWIC                        *ServicesLogger
	NODE                        *ServicesLogger
	FairLaunchDebugLogger       *ServicesLogger
	SendFairLaunchMintedAsset   *ServicesLogger
	FEE                         *ServicesLogger
	ScheduledTask               *ServicesLogger
	PreSale                     *ServicesLogger
	MintNft                     *ServicesLogger
	UserData                    *ServicesLogger
	UserStats                   *ServicesLogger
	CPAmm                       *ServicesLogger
	DateIpLogin                 *ServicesLogger
	PushQueue                   *ServicesLogger
	PoolPairTokenAccountBalance *ServicesLogger
	OpenChannel                 *ServicesLogger
	PsbtTlSwap                  *ServicesLogger
	PoolPure                    *ServicesLogger
	BoxAssetPush                *ServicesLogger
	BoxAssetPushTask            *ServicesLogger
	BoxChannelInfos             *ServicesLogger
	BoxDevice                   *ServicesLogger
	BoxDeviceBackup             *ServicesLogger
	Lnt                         *ServicesLogger
	FWDT                        *ServicesLogger
)

func loadDefaultLog() {
	Level := INFO
	{
		START = NewLogger("START", INFO, nil, true, startLogFile)
	}
	{
		CUST = NewLogger("CUST", Level, custErrorLogFile, true, defaultLogFile)
		CACC = NewLogger("CACC", Level, custErrorLogFile, false, CaccountLogFile)
		CLMT = NewLogger("CLMT", Level, custErrorLogFile, false, CLimitLogFile)
		CSWAP = NewLogger("CSWP", Level, custErrorLogFile, true, CSwapLogFile)
	}
	{
		FWDT = NewLogger("FWDT", Level, nil, false, fwdtLogFile)
	}
	{
		SWIC = NewLogger("SWIC", Level, nil, true, defaultLogFile)
		NODE = NewLogger("NODE", Level, nil, true, defaultLogFile)
	}
	{
		FairLaunchDebugLogger = NewLogger("FLDL", Level, nil, false, defaultLogFile, fairLaunchLogFile)
		SendFairLaunchMintedAsset = NewLogger("SFML", Level, nil, true, defaultLogFile, sendFairLaunchMintedAssetLogFile)
		FEE = NewLogger("FEE", Level, nil, true, defaultLogFile, feeLogFile)
		ScheduledTask = NewLogger("CRON", Level, nil, true, defaultLogFile)
		PreSale = NewLogger("PRSL", Level, nil, true, defaultLogFile, presaleLogFile)
		MintNft = NewLogger("MINT", Level, nil, false, mintNftFile, mintNftLogFile)
		UserData = NewLogger("URDT", Level, nil, true, defaultLogFile, userDataLogFile)
		UserStats = NewLogger("USTS", Level, nil, true, defaultLogFile, userStatsLogFile)
		CPAmm = NewLogger("CPAM", Level, nil, true, defaultLogFile, cpAmmLogFile)
		DateIpLogin = NewLogger("DILR", Level, nil, true, defaultLogFile, dateIpLoginLogFile)
		PushQueue = NewLogger("PUSH", Level, nil, true, defaultLogFile, pushQueueLogFile)
		PoolPairTokenAccountBalance = NewLogger("PTAB", Level, nil, true, defaultLogFile, poolPairTokenAccountBalanceLogFile)
		OpenChannel = NewLogger("OPCH", Level, nil, true, defaultLogFile, openChannelLogFile)
		PsbtTlSwap = NewLogger("PTLS", Level, nil, true, defaultLogFile, psbtTlSwapLogFile)
		PoolPure = NewLogger("PLPR", Level, nil, true, defaultLogFile, PoolPureLogFile)
		BoxAssetPush = NewLogger("BAPU", Level, nil, true, defaultLogFile, boxAssetPushLogFile)
		BoxAssetPushTask = NewLogger("BAPT", Level, nil, true, defaultLogFile, boxAssetPushTaskLogFile)
		BoxDeviceBackup = NewLogger("BDBK", Level, nil, true, defaultLogFile, boxDeviceBackupLogFile)
		BoxDevice = NewLogger("BDVC", Level, nil, true, defaultLogFile, boxDeviceLogFile)
		BoxChannelInfos = NewLogger("BCIN", Level, nil, true, defaultLogFile, boxChannelInfosLogFile)
		Lnt = NewLogger("LNT", Level, nil, true, defaultLogFile, lntLogFile)
	}
}
