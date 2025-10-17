package custodyBtc

import (
	"errors"
	"fmt"
	"trade/btlLog"
	"trade/middleware"
	"trade/models"
	"trade/models/custodyModels"
	"trade/services/btldb"
	caccount "trade/services/custodyAccount/account"
	"trade/services/custodyAccount/custodyBase/custodyFee"
	"trade/services/custodyAccount/custodyBase/custodyLimit"
	"trade/services/custodyAccount/defaultAccount/custodyBalance"
	rpc "trade/services/servicesrpc"

	"github.com/lightningnetwork/lnd/lnrpc"
	"gorm.io/gorm"
)

type BtcApplyInvoice struct {
	LnInvoice *lnrpc.AddInvoiceResponse
	Amount    int64
}

func (in *BtcApplyInvoice) GetAmount() int64 {
	return in.Amount
}
func (in *BtcApplyInvoice) GetPayReq() string {
	return in.LnInvoice.PaymentRequest
}

type BtcApplyInvoiceRequest struct {
	Amount int64
	Memo   string
}

func (req *BtcApplyInvoiceRequest) GetPayReqAmount() int64 {
	return req.Amount
}

type BtcPacketErr error

var (
	NotSufficientFunds BtcPacketErr = errors.New("not sufficient funds")
	DecodeInvoiceFail  BtcPacketErr = errors.New("decode invoice fail")
)

type BtcPacket struct {
	PayReq          string
	FeeLimit        int64
	DecodePayReq    *lnrpc.PayReq
	isInsideMission *isInsideMission
	err             chan error
}

func (p *BtcPacket) VerifyPayReq(userinfo *caccount.UserInfo) error {
	ServerFee := custodyFee.ChannelBtcServiceFee

	i, err := btldb.GetInvoiceByReq(p.PayReq)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		btlLog.CUST.Error("验证本地发票失败", err)
		return models.ReadDbErr
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		p.isInsideMission = nil
	} else {
		if i.Status != models.InvoiceStatusPending {
			return fmt.Errorf("发票已被使用")
		}
		if i.AssetId != "00" {
			return fmt.Errorf("该发票不能使用btc支付，请使用资产%s支付", i.AssetId)
		}
		p.isInsideMission = &isInsideMission{
			isInside:      true,
			insideInvoice: i,
		}
		ServerFee = custodyFee.ChannelBtcInsideServiceFee
	}

	p.DecodePayReq, err = rpc.InvoiceDecode(p.PayReq)
	if err != nil {
		btlLog.CUST.Error("发票解析失败", err)
		return fmt.Errorf("(pay_request=%s)", "发票解析失败：", p.PayReq)
	}
	if p.isInsideMission == nil && p.FeeLimit == 0 {
		p.FeeLimit = p.DecodePayReq.NumSatoshis / 10
		if p.FeeLimit < 1 {
			p.FeeLimit = 1
		}
	}

	endAmount := p.DecodePayReq.NumSatoshis + p.FeeLimit + int64(ServerFee)

	limitType := custodyModels.LimitType{
		AssetId:      "00",
		TransferType: custodyModels.LimitTransferTypeLocal,
	}
	if p.isInsideMission == nil {
		limitType.TransferType = custodyModels.LimitTransferTypeOutside
	}
	err = custodyLimit.CheckLimit(middleware.DB, userinfo, &limitType, float64(endAmount))
	if err != nil {
		return err
	}
	if !custodyBalance.CheckBtcBalance(middleware.DB, userinfo, float64(endAmount)) {
		return NotSufficientFunds
	}
	return nil
}

type isInsideMission struct {
	isInside      bool
	insideInvoice *models.Invoice
	err           chan error
}
type InvoiceResponce struct {
	Invoice string               `json:"invoice"`
	AssetId string               `json:"asset_id"`
	Amount  int64                `json:"amount"`
	Status  models.InvoiceStatus `json:"status"`
	Time    int64                `json:"time"`
}
