package custodyAssets

import (
	"encoding/hex"
	"errors"
	"fmt"
	"trade/btlLog"
	"trade/middleware"
	"trade/models"
	"trade/models/custodyModels"
	"trade/services/btldb"
	cBase "trade/services/custodyAccount/custodyBase"
	"trade/services/custodyAccount/custodyBase/custodyFee"
	"trade/services/custodyAccount/custodyBase/custodyLimit"
	"trade/services/custodyAccount/defaultAccount/custodyBalance"
	"trade/services/custodyAccount/defaultAccount/custodyBtc/mempool"
	rpc "trade/services/servicesrpc"

	"github.com/lightninglabs/taproot-assets/taprpc"
	"github.com/lightninglabs/taproot-assets/taprpc/rfqrpc"
	"github.com/lightninglabs/taproot-assets/taprpc/tapchannelrpc"
	"github.com/lightningnetwork/lnd/lnrpc"
	"gorm.io/gorm"
)

type AssetAddressApplyRequest struct {
	Amount int64
}

func (req *AssetAddressApplyRequest) GetPayReqAmount() int64 {
	return req.Amount
}

type AssetApplyAddress struct {
	Addr   *taprpc.Addr
	Amount int64
}

func (a *AssetApplyAddress) GetAmount() int64 {
	return a.Amount
}
func (a *AssetApplyAddress) GetPayReq() string {
	return a.Addr.Encoded
}

type AssetInvoiceApplyResponse struct {
	Amount int64
}

func (r *AssetInvoiceApplyResponse) GetPayReqAmount() int64 {
	return r.Amount
}

type AssetApplyInvoice struct {
	RfqInfo *rfqrpc.PeerAcceptedBuyQuote
	Invoice *lnrpc.AddInvoiceResponse
	Amount  int64
}

func (a *AssetApplyInvoice) GetAmount() int64 {
	return a.Amount
}
func (a *AssetApplyInvoice) GetPayReq() string {
	return a.Invoice.GetPaymentRequest()
}

type AssetPacket struct {
	PayReq          string
	DecodeAddr      *taprpc.Addr
	DecodeInvoice   *tapchannelrpc.AssetPayReqResponse
	isInsideMission *isInsideMission
	err             chan error
}

func (p *AssetPacket) VerifyPayReq(event *AssetEvent) error {
	ServerFee := uint64(mempool.GetCustodyAssetFee())

	i, err := btldb.GetInvoiceByReq(p.PayReq)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		btlLog.CUST.Error("验证本地发票失败", err)
		return models.ReadDbErr
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		p.isInsideMission = nil
	} else {
		if i.AssetId != *event.AssetId {
			return fmt.Errorf("this invoice can only be paid using the %s", i.AssetId)
		}
		p.isInsideMission = &isInsideMission{
			isInside:      true,
			insideInvoice: i,
		}
		ServerFee = custodyFee.AssetInsideFee
	}
	var assetId string
	var amount uint64

	flag := p.PayReq[0:2]
	switch flag {
	case "ln":
		if i.ID != 0 {
			if i.Status != models.InvoiceStatusPending {
				return fmt.Errorf("invoice is paid or canceled(pay_request=%s)", p.PayReq)
			}
			assetId = i.AssetId
			amount = uint64(i.Amount)
		} else {
			if event.AssetId == nil || *event.AssetId == "" {
				return fmt.Errorf("asset_id_is_empty(pay_request=%s)", p.PayReq)
			}
			assetId = *event.AssetId
			p.DecodeInvoice, err = rpc.DecodeAssetInvoice(p.PayReq, assetId)
			if err != nil {
				btlLog.CUST.Error("发票解析失败", err)
				return fmt.Errorf("%w(pay_request=%s)", cBase.DecodeInvoiceFail, p.PayReq)
			}
			amount = p.DecodeInvoice.AssetAmount
		}
	case "ta":
		p.DecodeAddr, err = rpc.DecodeAddr(p.PayReq)
		if err != nil {
			btlLog.CUST.Error("地址解析失败", err)
			return fmt.Errorf("%w(pay_request=%s)", cBase.DecodeAddressFail, p.PayReq)
		}
		assetId = hex.EncodeToString(p.DecodeAddr.AssetId)
		if *event.AssetId != "" && *event.AssetId != assetId {
			return fmt.Errorf("请使用资产%s进行支付", assetId)
		}
		amount = p.DecodeAddr.Amount
	default:
		return fmt.Errorf("unknown_payreq_type(pay_request=%s)", p.PayReq)
	}

	limitType := custodyModels.LimitType{
		AssetId:      assetId,
		TransferType: custodyModels.LimitTransferTypeLocal,
	}
	if p.isInsideMission == nil {
		limitType.TransferType = custodyModels.LimitTransferTypeOutside
	}
	err = custodyLimit.CheckLimit(middleware.DB, event.UserInfo, &limitType, float64(amount))
	if err != nil {
		return err
	}

	if !custodyBalance.CheckAssetBalance(middleware.DB, event.UserInfo, assetId, float64(amount)) {
		return cBase.NotEnoughAssetFunds
	}
	if !custodyBalance.CheckBtcBalance(middleware.DB, event.UserInfo, float64(ServerFee)) {
		return cBase.NotEnoughFeeFunds
	}
	return nil
}

type isInsideMission struct {
	isInside      bool
	insideInvoice *models.Invoice
	err           chan error
}

type OutsideMission struct {
	AddrTarget       []*target
	AssetID          string
	TotalAmount      int64
	RollBackNumber   int64
	MinPaymentNumber int64
}

type target struct {
	Mission *custodyModels.PayOutside
}
