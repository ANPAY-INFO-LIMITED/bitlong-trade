package control

import (
	"errors"
	"github.com/go-redis/redis/v8"
	"strings"
	"time"
	"trade/btlLog"
	"trade/middleware"
)

func GetTransferControl(assetId string, transferType TransferControl) bool {
	t := transferControlString{
		AssetId: assetId,
		Type:    transferType,
	}
	str := t.toString()

	var control string
	var err error
	var count int
	for i := 3; i > 0; i-- {
		control, err = middleware.RedisGet(str)
		if err == nil {
			break
		} else if errors.Is(err, redis.Nil) {
			count++
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		btlLog.CLMT.Error("get control failed:%v", err)
	}
	if count >= 3 && errors.Is(err, redis.Nil) {
		err = SetTransferControl(assetId, transferType, false)
		if err != nil {
			return false
		}
	}
	return control == "1"
}
func SetTransferControl(assetId string, transferType TransferControl, control bool) error {
	t := transferControlString{
		AssetId: assetId,
		Type:    transferType,
	}
	str := t.toString()
	btlLog.CLMT.Info("set Control function:%v, key:%v", str, control)
	err := middleware.RedisSet(str, control, 0)
	if err != nil {
		return err
	}
	return nil
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
