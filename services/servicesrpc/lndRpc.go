package servicesrpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"time"
	"trade/config"
	"trade/utils"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwallet"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/lightningnetwork/lnd/record"
	"google.golang.org/grpc"
)

func getLndConn() (*grpc.ClientConn, func()) {
	lndconf := config.GetConfig().ApiConfig.Lnd

	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath
	macaroonPath := lndconf.MacaroonPath

	return utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
}

func GetBlockInfo(hash string) (*chainrpc.GetBlockResponse, error) {

	blockHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, err
	}
	request := &chainrpc.GetBlockRequest{BlockHash: blockHash.CloneBytes()}
	response, err := getBlockInfo(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func getBlockInfo(request *chainrpc.GetBlockRequest) (*chainrpc.GetBlockResponse, error) {
	lndconf := config.GetConfig().ApiConfig.Lnd

	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath
	macaroonPath := lndconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := chainrpc.NewChainKitClient(conn)
	response, err := client.GetBlock(context.Background(), request)
	return response, err
}

func InvoiceCreate(amount int64, memo string) (*lnrpc.AddInvoiceResponse, error) {
	lndconf := config.GetConfig().ApiConfig.Lnd
	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath
	macaroonPath := config.GetConfig().ApiConfig.Lnd.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	request := &lnrpc.Invoice{
		Value: amount,
		Memo:  memo,
	}

	client := lnrpc.NewLightningClient(conn)
	response, err := client.AddInvoice(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return response, err
}

func InvoiceDecode(invoice string) (*lnrpc.PayReq, error) {
	lndconf := config.GetConfig().ApiConfig.Lnd

	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath
	macaroonPath := lndconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	request := &lnrpc.PayReqString{
		PayReq: invoice,
	}
	client := lnrpc.NewLightningClient(conn)
	response, err := client.DecodePayReq(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return response, err
}

func InvoiceFind(rHash []byte) (*lnrpc.Invoice, error) {
	lndconf := config.GetConfig().ApiConfig.Lnd

	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath
	macaroonPath := lndconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	request := &lnrpc.PaymentHash{
		RHash: rHash,
	}
	client := lnrpc.NewLightningClient(conn)
	response, err := client.LookupInvoice(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func InvoicePay(invoice string, amt, feeLimit int64) (*lnrpc.Payment, error) {
	lndconf := config.GetConfig().ApiConfig.Lnd
	macaroonFile := config.GetConfig().ApiConfig.Lnd.MacaroonPath
	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonFile)
	defer connClose()

	var paymentTimeout = time.Second * 60
	request := &routerrpc.SendPaymentRequest{
		PaymentRequest:    invoice,
		DestCustomRecords: make(map[uint64][]byte),

		TimeoutSeconds: int32(paymentTimeout.Seconds()),
		MaxParts:       16,
	}
	if feeLimit > 1 {
		request.FeeLimitSat = feeLimit
	} else {
		amtMsat := lnwire.NewMSatFromSatoshis(btcutil.Amount(amt))
		request.FeeLimitSat = int64(lnwallet.DefaultRoutingFeeLimitForAmount(amtMsat).ToSatoshis())
	}
	client := routerrpc.NewRouterClient(conn)
	stream, err := client.SendPaymentV2(context.Background(), request)
	if err != nil {
		return nil, err
	}
	for {
		payment, err := stream.Recv()
		if err != nil {
			return nil, err
		}

		if payment.Status != lnrpc.Payment_IN_FLIGHT &&
			payment.Status != lnrpc.Payment_INITIATED {
			return payment, nil
		}
	}
}

func PaymentTrack(paymentHash string) (*lnrpc.Payment, error) {
	lndconf := config.GetConfig().ApiConfig.Lnd

	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath
	macaroonPath := lndconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	hash, _ := hex.DecodeString(paymentHash)
	request := &routerrpc.TrackPaymentRequest{
		PaymentHash: hash,
	}
	client := routerrpc.NewRouterClient(conn)
	stream, err := client.TrackPaymentV2(context.Background(), request)
	if err != nil {
		return nil, err
	}
	defer func(stream routerrpc.Router_TrackPaymentV2Client) {
		err := stream.CloseSend()
		if err != nil {

		}
	}(stream)
	for {
		payment, err := stream.Recv()
		if err != nil {
			return nil, err
		}
		if payment != nil {
			if payment.Status == lnrpc.Payment_SUCCEEDED || payment.Status == lnrpc.Payment_FAILED {
				return payment, nil
			}
		}
	}
}

func InvoiceCancel(hash []byte) error {
	lndconf := config.GetConfig().ApiConfig.Lnd

	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath
	macaroonPath := lndconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := invoicesrpc.NewInvoicesClient(conn)
	request := &invoicesrpc.CancelInvoiceMsg{
		PaymentHash: hash,
	}
	_, err := client.CancelInvoice(context.Background(), request)
	if err != nil {
		return err
	}
	return nil
}

func ListUnspent() (*lnrpc.ListUnspentResponse, error) {
	lndconf := config.GetConfig().ApiConfig.Lnd

	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath
	macaroonPath := lndconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := lnrpc.NewLightningClient(conn)
	request := &lnrpc.ListUnspentRequest{
		Account: "default",
	}
	response, err := client.ListUnspent(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func GetBalance() (*lnrpc.WalletBalanceResponse, error) {
	lndconf := config.GetConfig().ApiConfig.Lnd

	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath
	macaroonPath := lndconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := lnrpc.NewLightningClient(conn)
	request := &lnrpc.WalletBalanceRequest{}
	response, err := client.WalletBalance(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func GetChannelInfo() ([]*lnrpc.Channel, error) {
	lndconf := config.GetConfig().ApiConfig.Lnd

	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath
	macaroonPath := lndconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := lnrpc.NewLightningClient(conn)
	request := lnrpc.ListChannelsRequest{
		PeerAliasLookup: true,
	}
	response, err := client.ListChannels(context.Background(), &request)
	if err != nil {
		return nil, err
	}
	return response.Channels, nil
}

func ListInvoices(maxNum int, reversed bool) (*lnrpc.ListInvoiceResponse, error) {
	lndconf := config.GetConfig().ApiConfig.Lnd

	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath
	macaroonPath := lndconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := lnrpc.NewLightningClient(conn)

	request := &lnrpc.ListInvoiceRequest{
		Reversed: reversed,
	}
	if maxNum > 0 {
		request.NumMaxInvoices = uint64(maxNum)
	}
	response, err := client.ListInvoices(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return response, err
}

func ListChannels(activeOnly bool, private bool) (*lnrpc.ListChannelsResponse, error) {
	lndconf := config.GetConfig().ApiConfig.Lnd

	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath
	macaroonPath := lndconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := lnrpc.NewLightningClient(conn)
	request := &lnrpc.ListChannelsRequest{
		ActiveOnly:  activeOnly,
		PrivateOnly: private,
	}
	response, err := client.ListChannels(context.Background(), request)
	if err != nil {
		fmt.Printf("lnrpc ListChannels err: %v\n", err)
		return nil, err
	}
	return response, nil
}

func SendPaymentV2ByKeySend(dest string, amt int, feelimit int, outgoingChanId uint64) (*lnrpc.Payment, error) {
	lndconf := config.GetConfig().ApiConfig.Lnd

	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath
	macaroonPath := lndconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := routerrpc.NewRouterClient(conn)

	destNode, _ := hex.DecodeString(dest)
	if len(destNode) != 33 {
		return nil, fmt.Errorf("dest node pubkey must be exactly 33 bytes, is "+
			"instead: %v", len(dest))
	}
	req := &routerrpc.SendPaymentRequest{
		TimeoutSeconds:    30,
		Dest:              destNode,
		Amt:               int64(amt),
		FeeLimitSat:       int64(feelimit),
		DestCustomRecords: make(map[uint64][]byte),
		OutgoingChanIds:   make([]uint64, 1),
	}
	req.OutgoingChanIds[0] = outgoingChanId

	var rHash []byte
	var preimage lntypes.Preimage
	if _, err := rand.Read(preimage[:]); err != nil {
		return nil, err
	}

	req.DestCustomRecords[record.KeySendType] = preimage[:]
	hash := preimage.Hash()
	rHash = hash[:]
	req.PaymentHash = rHash

	ctx := context.Background()
	ctxt, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	stream, err := client.SendPaymentV2(ctxt, req)
	if err != nil {
		return nil, err
	}
	for {
		response, err := stream.Recv()
		if err != nil {
			return nil, err
		} else if response != nil {
			if response.Status == 2 || response.Status == 3 {
				return response, nil
			}
		}
	}
}

func SubscribeHtlcEvents(f func(e *routerrpc.HtlcEvent)) {
	lndconf := config.GetConfig().ApiConfig.Lnd

	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath
	macaroonPath := lndconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := routerrpc.NewRouterClient(conn)

	if f == nil {
		f = func(e *routerrpc.HtlcEvent) {
			fmt.Println(e)
		}
	}
	request := &routerrpc.SubscribeHtlcEventsRequest{}
	stream, err := client.SubscribeHtlcEvents(context.Background(), request)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		event, err := stream.Recv()
		if err != nil {
			fmt.Println(err)
			return
		}
		if event != nil {
			f(event)
		}
	}
}

func SendCoins(addr string, amount int64, satPerVbyte uint64) (string, error) {
	lndconf := config.GetConfig().ApiConfig.Lnd

	grpcHost := lndconf.Host + ":" + strconv.Itoa(lndconf.Port)
	tlsCertPath := lndconf.TlsCertPath
	macaroonPath := lndconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := lnrpc.NewLightningClient(conn)
	request := &lnrpc.SendCoinsRequest{
		Addr:        addr,
		Amount:      amount,
		SatPerVbyte: satPerVbyte,
	}
	response, err := client.SendCoins(context.Background(), request)
	if err != nil {
		return "", err
	}
	return response.Txid, nil
}

func Getbestblock() (*chainrpc.GetBestBlockResponse, error) {
	coon, closecoon := getLndConn()
	defer closecoon()
	client := chainrpc.NewChainKitClient(coon)
	request := &chainrpc.GetBestBlockRequest{}
	response, err := client.GetBestBlock(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func GetTransactions(startHeight int32, endHeight int32) (*lnrpc.TransactionDetails, error) {
	coon, closecoon := getLndConn()
	defer closecoon()
	client := lnrpc.NewLightningClient(coon)
	request := &lnrpc.GetTransactionsRequest{
		StartHeight: startHeight,
		EndHeight:   endHeight,
		Account:     "default",
	}
	response, err := client.GetTransactions(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func NewAddress(addressType lnrpc.AddressType) (string, error) {
	coon, closecoon := getLndConn()
	defer closecoon()
	client := lnrpc.NewLightningClient(coon)
	request := &lnrpc.NewAddressRequest{
		Type: addressType,
	}
	response, err := client.NewAddress(context.Background(), request)
	if err != nil {
		return "", err
	}
	return response.Address, nil
}

func EstimateRouteFeeByPayReq(payReq string, timeout uint32) (*routerrpc.RouteFeeResponse, error) {
	coon, closecoon := getLndConn()
	defer closecoon()
	client := routerrpc.NewRouterClient(coon)
	request := &routerrpc.RouteFeeRequest{
		PaymentRequest: payReq,
		Timeout:        timeout,
	}
	response, err := client.EstimateRouteFee(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func GetLndInfo() (*lnrpc.GetInfoResponse, error) {
	coon, closecoon := getLndConn()
	defer closecoon()
	client := lnrpc.NewLightningClient(coon)
	request := &lnrpc.GetInfoRequest{}
	response, err := client.GetInfo(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return response, nil
}
