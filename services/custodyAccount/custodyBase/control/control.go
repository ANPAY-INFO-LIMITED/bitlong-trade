package control

import (
	"errors"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"strings"
	"time"
	"trade/btlLog"
	"trade/middleware"
	"trade/models/custodyModels"
)

var controlMap = make(map[string]bool)

func GetTransferControl(assetId string, transferType TransferControl) bool {
	t := transferControlString{
		AssetId: assetId,
		Type:    transferType,
	}
	str := t.toString()
	if value, exists := controlMap[str]; exists {
		return value
	} else {
		return true
	}
}

func SetTransferControl(assetId string, transferType TransferControl, control bool) error {
	t := transferControlString{
		AssetId: assetId,
		Type:    transferType,
	}
	str := t.toString()

	ctrl := custodyModels.Control{
		ControlName: str,
	}
	err := middleware.DB.First(&ctrl).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		btlLog.CLMT.Error("set control failed:%v", err)
		return err
	}
	if ctrl.Status != control || ctrl.Status == false {
		ctrl.Status = control
		err = middleware.DB.Save(&ctrl).Error
		if err != nil {
			btlLog.CLMT.Error("set control failed:%v", err)
			return err
		}
	}
	loadingControlMap()
	return nil
}

func loadingControlMap() {
	var controls []custodyModels.Control
	err := middleware.DB.Find(&controls).Error
	if err != nil {
		btlLog.CLMT.Error("loading control map failed:%v", err)
		return
	}
	for _, control := range controls {
		controlMap[control.ControlName] = control.Status
	}
}

type transferControlString struct {
	AssetId string
	Type    TransferControl
}

func (t *transferControlString) toString() string {
	return t.AssetId + "," + string(t.Type)
}

func (t *transferControlString) fromString(s string) {
	parts := strings.Split(s, ",")
	if len(parts) >= 2 {
		t.AssetId = parts[0]
		t.Type = TransferControl(parts[1])
	}
}

type TransferControl string

const (
	TransferControlOnChain  TransferControl = "OnChain"
	TransferControlOffChain                 = "OffChain"
	TransferControlLocal                    = "Local"
)

func (t *TransferControl) ToInt() string {
	switch *t {
	case TransferControlOnChain:
		return "1"
	case TransferControlOffChain:
		return "2"
	case TransferControlLocal:
		return "3"
	default:
		return "0"
	}
}
func (t *TransferControl) FromInt(s int) error {
	switch s {
	case 1:
		*t = TransferControlOnChain
	case 2:
		*t = TransferControlOffChain
	case 3:
		*t = TransferControlLocal
	default:
		return errors.New("invalid transfer control")
	}
	return nil
}

func ControlTest() {
	var err error
	success, failed := 0, 0
	btlLog.CLMT.Info("set Control function:%v, key:%v", "control,1", true)
	err = middleware.RedisSet("control,1", true, 0)
	if err != nil {
		btlLog.CLMT.Error("set control failed:%v", err)
	}
	for {
		_, err := middleware.RedisGet("control,1")
		if err != nil {
			btlLog.CLMT.Error("get control failed:%v", err)
			failed++
			btlLog.CLMT.Info("success:%d, failed:%d", success, failed)
			time.Sleep(time.Second)
			if failed >= 300 {
				return
			}
			continue
		}
		if errors.Is(err, redis.Nil) {
			failed++
		}
		success++
		btlLog.CLMT.Info("success:%d, failed:%d", success, failed)
		time.Sleep(time.Second)
	}
}
