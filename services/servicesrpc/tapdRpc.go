package servicesrpc

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"trade/config"
	"trade/utils"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/lightninglabs/taproot-assets/fn"
	"github.com/lightninglabs/taproot-assets/proof"
	"github.com/lightninglabs/taproot-assets/rfq"
	"github.com/lightninglabs/taproot-assets/rpcutils"
	"github.com/lightninglabs/taproot-assets/taprpc"
	"github.com/lightninglabs/taproot-assets/taprpc/tapchannelrpc"
	"github.com/lightninglabs/taproot-assets/taprpc/universerpc"
	"github.com/lightningnetwork/lnd/cmd/commands"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"google.golang.org/grpc"
)

func getTapdCoon() (*grpc.ClientConn, func()) {
	tapdconf := config.GetConfig().ApiConfig.Tapd
	grpcHost := tapdconf.Host + ":" + strconv.Itoa(tapdconf.Port)
	tlsCertPath := tapdconf.TlsCertPath
	macaroonPath := tapdconf.MacaroonPath
	return utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
}
func GetAssetLeaves(ID string, isGroup bool, proofType string) (*universerpc.AssetLeafResponse, error) {
	requset := universerpc.ID{}
	var p universerpc.ProofType
	switch proofType {
	case "issuance":
		p = universerpc.ProofType_PROOF_TYPE_ISSUANCE
	case "transfer":
		p = universerpc.ProofType_PROOF_TYPE_TRANSFER
	default:
		return nil, fmt.Errorf("unknown proof type: %s", proofType)
	}
	requset.ProofType = p

	if isGroup {
		groupId := universerpc.ID_GroupKeyStr{
			GroupKeyStr: ID,
		}
		requset.Id = &groupId
	} else {
		assetId := universerpc.ID_AssetIdStr{
			AssetIdStr: ID,
		}
		requset.Id = &assetId
	}

	leaves, err := getAssetLeaves(&requset)
	if err != nil {
		return nil, err
	}
	return leaves, nil

}

func GetAssetMeta(ID string, isHash bool) (*taprpc.AssetMeta, error) {
	var request taprpc.FetchAssetMetaRequest
	if isHash {
		assetHast := taprpc.FetchAssetMetaRequest_MetaHashStr{
			MetaHashStr: ID,
		}
		request.Asset = &assetHast
	} else {
		assetId := taprpc.FetchAssetMetaRequest_AssetIdStr{
			AssetIdStr: ID,
		}
		request.Asset = &assetId
	}
	assetMeta, err := getAssetMeta(&request)
	if err != nil {
		return nil, err
	}
	return assetMeta, nil
}

func SyncAsset(universe string, id string, isGroupKey bool, proofType string) (*universerpc.SyncResponse, error) {
	request := universerpc.SyncRequest{}
	var p universerpc.ProofType
	switch proofType {
	case "issuance":
		p = universerpc.ProofType_PROOF_TYPE_ISSUANCE
	case "transfer":
		p = universerpc.ProofType_PROOF_TYPE_TRANSFER
	default:
		return nil, fmt.Errorf("unknown proof type: %s", proofType)
	}

	if isGroupKey {
		groupKey := universerpc.ID_GroupKeyStr{
			GroupKeyStr: id,
		}
		request.SyncTargets = append(request.SyncTargets, &universerpc.SyncTarget{
			Id: &universerpc.ID{Id: &groupKey,
				ProofType: p},
		})
	} else {
		assetId := universerpc.ID_AssetIdStr{
			AssetIdStr: id,
		}
		request.SyncTargets = append(request.SyncTargets, &universerpc.SyncTarget{
			Id: &universerpc.ID{Id: &assetId,
				ProofType: p},
		})
	}
	request.UniverseHost = universe
	request.SyncMode = universerpc.UniverseSyncMode_SYNC_ISSUANCE_ONLY
	response, err := syncAsset(&request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func InsertProof(annotatedProof *proof.AnnotatedProof) error {

	proofFile := &proof.File{}
	err := proofFile.Decode(bytes.NewReader(annotatedProof.Blob))
	if err != nil {
		return err
	}

	for i := 0; i < proofFile.NumProofs(); i++ {
		transitionProof, err := proofFile.ProofAt(uint32(i))
		if err != nil {
			return err
		}
		proofAsset := transitionProof.Asset

		rpcAsset, err := rpcutils.MarshalAsset(
			context.Background(), &proofAsset, true, true, nil, fn.None[uint32](),
		)
		if err != nil {
			return err
		}

		var proofBuf bytes.Buffer
		if err := transitionProof.Encode(&proofBuf); err != nil {
			return fmt.Errorf("error encoding proof file: %w", err)
		}

		assetLeaf := universerpc.AssetLeaf{
			Asset: rpcAsset,
			Proof: proofBuf.Bytes(),
		}

		outPoint := transitionProof.OutPoint()
		assetKey := rpcutils.MarshalAssetKey(
			outPoint, proofAsset.ScriptKey.PubKey,
		)
		assetID := proofAsset.ID()

		var (
			groupPubKey      *btcec.PublicKey
			groupPubKeyBytes []byte
		)
		if proofAsset.GroupKey != nil {
			groupPubKey = &proofAsset.GroupKey.GroupPubKey
			groupPubKeyBytes = groupPubKey.SerializeCompressed()
		}

		universeID := rpcutils.MarshalUniverseID(
			assetID[:], groupPubKeyBytes,
		)
		universeKey := universerpc.UniverseKey{
			Id:      universeID,
			LeafKey: assetKey,
		}

		err = insertProof(&universerpc.AssetProof{
			Key:       &universeKey,
			AssetLeaf: &assetLeaf,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func getAssetLeaves(request *universerpc.ID) (*universerpc.AssetLeafResponse, error) {
	tapdconf := config.GetConfig().ApiConfig.Tapd
	grpcHost := tapdconf.Host + ":" + strconv.Itoa(tapdconf.Port)
	tlsCertPath := tapdconf.TlsCertPath
	macaroonPath := tapdconf.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := universerpc.NewUniverseClient(conn)
	response, err := client.AssetLeaves(context.Background(), request)
	return response, err
}

func getAssetMeta(request *taprpc.FetchAssetMetaRequest) (*taprpc.AssetMeta, error) {
	tapdconf := config.GetConfig().ApiConfig.Tapd

	grpcHost := tapdconf.Host + ":" + strconv.Itoa(tapdconf.Port)
	tlsCertPath := tapdconf.TlsCertPath
	macaroonPath := tapdconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := taprpc.NewTaprootAssetsClient(conn)
	response, err := client.FetchAssetMeta(context.Background(), request)
	return response, err
}

func syncAsset(request *universerpc.SyncRequest) (*universerpc.SyncResponse, error) {
	tapdconf := config.GetConfig().ApiConfig.Tapd

	grpcHost := tapdconf.Host + ":" + strconv.Itoa(tapdconf.Port)
	tlsCertPath := tapdconf.TlsCertPath
	macaroonPath := tapdconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := universerpc.NewUniverseClient(conn)
	response, err := client.SyncUniverse(context.Background(), request)
	return response, err
}

func insertProof(request *universerpc.AssetProof) error {
	tapdconf := config.GetConfig().ApiConfig.Tapd

	grpcHost := tapdconf.Host + ":" + strconv.Itoa(tapdconf.Port)
	tlsCertPath := tapdconf.TlsCertPath
	macaroonPath := tapdconf.MacaroonPath

	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := universerpc.NewUniverseClient(conn)
	_, err := client.InsertProof(context.Background(), request)
	if err != nil {
		return err
	}
	return nil
}

func NewAddr(assetId string, amt int, proofCourierAddr string) (*taprpc.Addr, error) {
	tapdconf := config.GetConfig().ApiConfig.Tapd
	grpcHost := tapdconf.Host + ":" + strconv.Itoa(tapdconf.Port)
	tlsCertPath := tapdconf.TlsCertPath
	macaroonPath := tapdconf.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := taprpc.NewTaprootAssetsClient(conn)
	_assetIdByteSlice, _ := hex.DecodeString(assetId)
	if !strings.HasPrefix(proofCourierAddr, "universerpc: 
		proofCourierAddr = "universerpc: 
}
request := &taprpc.NewAddrRequest{
AssetId:          _assetIdByteSlice,
Amt:              uint64(amt),
ProofCourierAddr: proofCourierAddr,
}
response, err := client.NewAddr(context.Background(), request)
if err != nil {
return nil, err
}
return response, nil
}

func DecodeAddr(addr string) (*taprpc.Addr, error) {
	tapdconf := config.GetConfig().ApiConfig.Tapd
	grpcHost := tapdconf.Host + ":" + strconv.Itoa(tapdconf.Port)
	tlsCertPath := tapdconf.TlsCertPath
	macaroonPath := tapdconf.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := taprpc.NewTaprootAssetsClient(conn)
	request := &taprpc.DecodeAddrRequest{
		Addr: addr,
	}
	response, err := client.DecodeAddr(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func ListAssets() (*taprpc.ListAssetResponse, error) {
	tapdconf := config.GetConfig().ApiConfig.Tapd
	grpcHost := tapdconf.Host + ":" + strconv.Itoa(tapdconf.Port)
	tlsCertPath := tapdconf.TlsCertPath
	macaroonPath := tapdconf.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := taprpc.NewTaprootAssetsClient(conn)
	request := &taprpc.ListAssetRequest{}
	response, err := client.ListAssets(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func SendAssets(addr []string) (*taprpc.SendAssetResponse, error) {
	tapdconf := config.GetConfig().ApiConfig.Tapd
	grpcHost := tapdconf.Host + ":" + strconv.Itoa(tapdconf.Port)
	tlsCertPath := tapdconf.TlsCertPath
	macaroonPath := tapdconf.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := taprpc.NewTaprootAssetsClient(conn)
	request := &taprpc.SendAssetRequest{
		TapAddrs: addr,
	}
	response, err := client.SendAsset(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func QueryAssetRoot(Id string) *universerpc.QueryRootResponse {
	tapdconf := config.GetConfig().ApiConfig.Tapd
	grpcHost := tapdconf.Host + ":" + strconv.Itoa(tapdconf.Port)
	tlsCertPath := tapdconf.TlsCertPath
	macaroonPath := tapdconf.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)

	defer connClose()
	in := &universerpc.AssetRootQuery{}
	in.Id = &universerpc.ID{
		Id: &universerpc.ID_AssetIdStr{
			AssetIdStr: Id,
		},
	}
	client := universerpc.NewUniverseClient(conn)
	roots, err := client.QueryAssetRoots(context.Background(), in)
	if err != nil {
		return nil
	}
	return roots
}

func SubscribeReceiveEvents() (*taprpc.ReceiveEvent, error) {
	tapdconf := config.GetConfig().ApiConfig.Tapd
	grpcHost := tapdconf.Host + ":" + strconv.Itoa(tapdconf.Port)
	tlsCertPath := tapdconf.TlsCertPath
	macaroonPath := tapdconf.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := taprpc.NewTaprootAssetsClient(conn)
	request := &taprpc.SubscribeReceiveEventsRequest{}
	stream, err := client.SubscribeReceiveEvents(context.Background(), request)
	if err != nil {
		return nil, err
	}
	for {
		event, err := stream.Recv()
		if err != nil {
			return nil, err
		}
		fmt.Println(event)
	}
}

func ListAssetsBalance() (*taprpc.ListBalancesResponse, error) {
	tapdconf := config.GetConfig().ApiConfig.Tapd
	grpcHost := tapdconf.Host + ":" + strconv.Itoa(tapdconf.Port)
	tlsCertPath := tapdconf.TlsCertPath
	macaroonPath := tapdconf.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()

	client := taprpc.NewTaprootAssetsClient(conn)
	request := &taprpc.ListBalancesRequest{}
	request.GroupBy = &taprpc.ListBalancesRequest_AssetId{
		AssetId: true,
	}
	response, err := client.ListBalances(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func AddAssetChannelInvoice(assetId string, assetAmount uint64, PeerPubkey string, memo string) (*tapchannelrpc.AddInvoiceResponse, error) {
	conn, closeConn := getTapdCoon()
	defer closeConn()
	if conn == nil {
		return nil, errors.New("connection is nil")
	}
	assetIDBytes, err := hex.DecodeString(assetId)
	if err != nil {
		return nil, err
	}
	expirySeconds := int64(rfq.DefaultInvoiceExpiry.Seconds())
	rfqPeerKey, err := hex.DecodeString(PeerPubkey)
	if err != nil {
		return nil, err
	}
	channelsClient := tapchannelrpc.NewTaprootAssetChannelsClient(conn)
	resp, err := channelsClient.AddInvoice(context.Background(), &tapchannelrpc.AddInvoiceRequest{
		AssetId:     assetIDBytes,
		AssetAmount: assetAmount,
		PeerPubkey:  rfqPeerKey,
		InvoiceRequest: &lnrpc.Invoice{
			Memo:   memo,
			Expiry: expirySeconds,
		},
	})
	return resp, err
}

func PayAssetInvoice(payReq string, assetIDStr string, peerPubkey string) (*lnrpc.Payment, error) {

	ctx := context.Background()

	conn, closeConn := getTapdCoon()
	defer closeConn()

	assetIDBytes, err := hex.DecodeString(assetIDStr)
	if err != nil {
		return nil, fmt.Errorf("unable to decode assetID: %v", err)
	}

	rfqPeerKey, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return nil, fmt.Errorf("unable to decode RFQ peer public key: "+
			"%w", err)
	}

	req := &routerrpc.SendPaymentRequest{
		PaymentRequest: commands.StripPrefix(payReq),
	}

	pmtTimeout := 60 * time.Second

	req.TimeoutSeconds = int32(pmtTimeout.Seconds())

	req.FeeLimitSat = 1000

	tchrpcClient := tapchannelrpc.NewTaprootAssetChannelsClient(
		conn,
	)

	stream, err := tchrpcClient.SendPayment(
		ctx, &tapchannelrpc.SendPaymentRequest{
			AssetId:        assetIDBytes,
			PeerPubkey:     rfqPeerKey,
			PaymentRequest: req,
			AllowOverpay:   true,
		},
	)
	if err != nil {
		return nil, err
	}
	for {
		resp1, err := stream.Recv()
		if err != nil {
			return nil, err
		} else if resp1 != nil {
			payment := resp1.GetPaymentResult()
			if payment != nil {
				if payment.Status != lnrpc.Payment_IN_FLIGHT &&
					payment.Status != lnrpc.Payment_INITIATED {
					return payment, nil
				}
			}
		}
	}
}

func DecodeAssetInvoice(payReq string, assetIDStr string) (*tapchannelrpc.AssetPayReqResponse, error) {
	assetIDBytes, err := hex.DecodeString(assetIDStr)
	if err != nil {
		return nil, fmt.Errorf("unable to decode assetID: %v", err)
	}

	conn, closeConn := getTapdCoon()
	defer closeConn()

	channelsClient := tapchannelrpc.NewTaprootAssetChannelsClient(conn)
	return channelsClient.DecodeAssetPayReq(context.Background(), &tapchannelrpc.AssetPayReq{
		AssetId:      assetIDBytes,
		PayReqString: payReq,
	})
}
