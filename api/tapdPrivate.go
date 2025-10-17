package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
	"strings"
	"trade/btlLog"
	"trade/config"
	"trade/models"
	"trade/rpc/btlchannelrpc"
	"trade/services/nodemanage"
	"trade/utils"

	"github.com/lightninglabs/taproot-assets/rfqmath"
	"github.com/lightninglabs/taproot-assets/taprpc"
	"github.com/lightninglabs/taproot-assets/taprpc/mintrpc"
	"github.com/lightninglabs/taproot-assets/taprpc/tapchannelrpc"
	"github.com/lightninglabs/taproot-assets/taprpc/universerpc"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/record"
)

func assetLeaves(isGroup bool, id string, proofType universerpc.ProofType) (*universerpc.AssetLeafResponse, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	request := &universerpc.ID{
		ProofType: proofType,
	}
	if isGroup {
		groupKey := &universerpc.ID_GroupKeyStr{
			GroupKeyStr: id,
		}
		request.Id = groupKey
	} else {
		AssetId := &universerpc.ID_AssetIdStr{
			AssetIdStr: id,
		}
		request.Id = AssetId
	}
	client := universerpc.NewUniverseClient(conn)
	response, err := client.AssetLeaves(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "AssetLeaves")
	}
	return response, nil
}

func assetLeafKeys(isGroup bool, id string, proofType universerpc.ProofType) (*universerpc.AssetLeafKeyResponse, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	request := &universerpc.AssetLeafKeysRequest{
		Id: &universerpc.ID{
			ProofType: proofType,
		},
	}
	if isGroup {
		groupKey := &universerpc.ID_GroupKeyStr{
			GroupKeyStr: id,
		}
		request.Id.Id = groupKey
	} else {
		AssetId := &universerpc.ID_AssetIdStr{
			AssetIdStr: id,
		}
		request.Id.Id = AssetId
	}
	client := universerpc.NewUniverseClient(conn)
	response, err := client.AssetLeafKeys(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "AssetLeafKeys")
	}
	return response, nil
}

func queryProof(isGroup bool, id string, outpoint string, scriptKey string, proofType universerpc.ProofType) (*universerpc.AssetProofResponse, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	request := &universerpc.UniverseKey{
		Id: &universerpc.ID{
			ProofType: proofType,
		},
		LeafKey: &universerpc.AssetKey{
			Outpoint:  &universerpc.AssetKey_OpStr{OpStr: outpoint},
			ScriptKey: &universerpc.AssetKey_ScriptKeyStr{ScriptKeyStr: scriptKey},
		},
	}
	if isGroup {
		groupKey := &universerpc.ID_GroupKeyStr{
			GroupKeyStr: id,
		}
		request.Id.Id = groupKey
	} else {
		AssetId := &universerpc.ID_AssetIdStr{
			AssetIdStr: id,
		}
		request.Id.Id = AssetId
	}
	client := universerpc.NewUniverseClient(conn)
	response, err := client.QueryProof(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "QueryProof")
	}
	return response, nil
}

func assetLeavesSpecified(id string, proofType string) (*universerpc.AssetLeafResponse, error) {
	var _proofType universerpc.ProofType
	if proofType == "issuance" || proofType == "ISSUANCE" || proofType == "PROOF_TYPE_ISSUANCE" {
		_proofType = universerpc.ProofType_PROOF_TYPE_ISSUANCE
	} else if proofType == "transfer" || proofType == "TRANSFER" || proofType == "PROOF_TYPE_TRANSFER" {
		_proofType = universerpc.ProofType_PROOF_TYPE_TRANSFER
	} else {
		_proofType = universerpc.ProofType_PROOF_TYPE_UNSPECIFIED
	}
	return assetLeaves(false, id, _proofType)
}

func processAssetIssuanceLeaf(response *universerpc.AssetLeafResponse) *models.AssetIssuanceLeaf {
	if response == nil || response.Leaves == nil || len(response.Leaves) == 0 {
		return nil
	}
	var groupKey string
	assetGroup := response.Leaves[0].Asset.AssetGroup
	if assetGroup != nil {
		groupKey = hex.EncodeToString(assetGroup.TweakedGroupKey)
	}
	return &models.AssetIssuanceLeaf{
		Version:            response.Leaves[0].Asset.Version.String(),
		GenesisPoint:       response.Leaves[0].Asset.AssetGenesis.GenesisPoint,
		Name:               response.Leaves[0].Asset.AssetGenesis.Name,
		MetaHash:           hex.EncodeToString(response.Leaves[0].Asset.AssetGenesis.MetaHash),
		AssetID:            hex.EncodeToString(response.Leaves[0].Asset.AssetGenesis.AssetId),
		AssetType:          response.Leaves[0].Asset.AssetGenesis.AssetType,
		GenesisOutputIndex: int(response.Leaves[0].Asset.AssetGenesis.OutputIndex),
		Amount:             int(response.Leaves[0].Asset.Amount),
		LockTime:           int(response.Leaves[0].Asset.LockTime),
		RelativeLockTime:   int(response.Leaves[0].Asset.RelativeLockTime),
		ScriptVersion:      int(response.Leaves[0].Asset.ScriptVersion),
		ScriptKey:          hex.EncodeToString(response.Leaves[0].Asset.ScriptKey),
		ScriptKeyIsLocal:   response.Leaves[0].Asset.ScriptKeyIsLocal,
		TweakedGroupKey:    groupKey,
		IsSpent:            response.Leaves[0].Asset.IsSpent,
		LeaseOwner:         hex.EncodeToString(response.Leaves[0].Asset.LeaseOwner),
		LeaseExpiry:        int(response.Leaves[0].Asset.LeaseExpiry),
		IsBurn:             response.Leaves[0].Asset.IsBurn,
		Proof:              hex.EncodeToString(response.Leaves[0].Proof),
	}
}

func assetLeafIssuanceInfo(id string) (*models.AssetIssuanceLeaf, error) {
	response, err := assetLeavesSpecified(id, universerpc.ProofType_PROOF_TYPE_ISSUANCE.String())
	if response == nil {
		return nil, err
	}
	return processAssetIssuanceLeaf(response), nil
}

func mintAsset(assetVersionIsV1 bool, assetTypeIsCollectible bool, name string, assetMetaData string, AssetMetaTypeIsJsonNotOpaque bool, amount int, newGroupedAsset bool, groupedAsset bool, groupKey string, groupAnchor string, shortResponse bool) (*mintrpc.MintAssetResponse, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := mintrpc.NewMintClient(conn)
	var _assetVersion taprpc.AssetVersion
	if assetVersionIsV1 {
		_assetVersion = taprpc.AssetVersion_ASSET_VERSION_V1
	} else {
		_assetVersion = taprpc.AssetVersion_ASSET_VERSION_V0
	}
	var _assetType taprpc.AssetType
	if assetTypeIsCollectible {
		_assetType = taprpc.AssetType_COLLECTIBLE
	} else {
		_assetType = taprpc.AssetType_NORMAL
	}
	_assetMetaDataByteSlice := []byte(assetMetaData)
	var _assetMetaType taprpc.AssetMetaType
	if AssetMetaTypeIsJsonNotOpaque {
		_assetMetaType = taprpc.AssetMetaType_META_TYPE_JSON
	} else {
		_assetMetaType = taprpc.AssetMetaType_META_TYPE_OPAQUE
	}
	_groupKeyByteSlices := []byte(groupKey)
	request := &mintrpc.MintAssetRequest{
		Asset: &mintrpc.MintAsset{
			AssetVersion: _assetVersion,
			AssetType:    _assetType,
			Name:         name,
			AssetMeta: &taprpc.AssetMeta{
				Data: _assetMetaDataByteSlice,
				Type: _assetMetaType,
			},
			Amount:          uint64(amount),
			NewGroupedAsset: newGroupedAsset,
			GroupedAsset:    groupedAsset,
			GroupKey:        _groupKeyByteSlices,
			GroupAnchor:     groupAnchor,
		},
		ShortResponse: shortResponse,
	}
	response, err := client.MintAsset(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "MintAsset")
	}
	return response, nil
}

func mintAssetByParam(assetVersion taprpc.AssetVersion, assetType taprpc.AssetType, name string, assetMetaData []byte, assetMetaType taprpc.AssetMetaType, amount uint64, newGroupedAsset bool, groupedAsset bool, groupKey []byte, groupAnchor string, shortResponse bool) (*mintrpc.MintAssetResponse, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := mintrpc.NewMintClient(conn)
	request := &mintrpc.MintAssetRequest{
		Asset: &mintrpc.MintAsset{
			AssetVersion: assetVersion,
			AssetType:    assetType,
			Name:         name,
			AssetMeta: &taprpc.AssetMeta{
				Data: assetMetaData,
				Type: assetMetaType,
			},
			Amount:          amount,
			NewGroupedAsset: newGroupedAsset,
			GroupedAsset:    groupedAsset,
			GroupKey:        groupKey,
			GroupAnchor:     groupAnchor,
		},
		ShortResponse: shortResponse,
	}
	response, err := client.MintAsset(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "MintAsset")
	}
	return response, nil
}

func finalizeBatch(shortResponse bool, feeRate int) (*mintrpc.FinalizeBatchResponse, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := mintrpc.NewMintClient(conn)
	request := &mintrpc.FinalizeBatchRequest{
		ShortResponse: shortResponse,
		FeeRate:       uint32(feeRate),
	}
	response, err := client.FinalizeBatch(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "FinalizeBatch")
	}
	return response, nil
}

func cancelBatch() (*mintrpc.CancelBatchResponse, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := mintrpc.NewMintClient(conn)
	request := &mintrpc.CancelBatchRequest{}
	response, err := client.CancelBatch(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "CancelBatch")
	}
	return response, nil
}

func fetchAssetMeta(isHash bool, data string) (*taprpc.AssetMeta, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := taprpc.NewTaprootAssetsClient(conn)
	request := &taprpc.FetchAssetMetaRequest{}
	if isHash {
		request.Asset = &taprpc.FetchAssetMetaRequest_MetaHashStr{
			MetaHashStr: data,
		}
	} else {
		request.Asset = &taprpc.FetchAssetMetaRequest_AssetIdStr{
			AssetIdStr: data,
		}
	}
	response, err := client.FetchAssetMeta(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "FetchAssetMeta")
	}
	return response, nil
}

func newAddr(assetId string, amt int, proofCourierAddr string) (*taprpc.Addr, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
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
return nil, utils.AppendErrorInfo(err, "NewAddr")
}
return response, nil
}

func sendAsset(tapAddrs string, feeRate int) (*taprpc.SendAssetResponse, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := taprpc.NewTaprootAssetsClient(conn)
	addrs := strings.Split(tapAddrs, ",")
	request := &taprpc.SendAssetRequest{
		TapAddrs: addrs,
		FeeRate:  uint32(feeRate),
	}
	response, err := client.SendAsset(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "SendAsset")
	}
	return response, nil
}

func sendAssetAddrSlice(addrSlice []string, feeRate int) (*taprpc.SendAssetResponse, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := taprpc.NewTaprootAssetsClient(conn)
	request := &taprpc.SendAssetRequest{
		TapAddrs: addrSlice,
		FeeRate:  uint32(feeRate),
	}
	response, err := client.SendAsset(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "SendAsset")
	}
	return response, nil
}

func decodeAddr(addr string) (*taprpc.Addr, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := taprpc.NewTaprootAssetsClient(conn)
	request := &taprpc.DecodeAddrRequest{
		Addr: addr,
	}
	response, err := client.DecodeAddr(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "DecodeAddr")
	}
	return response, nil
}

func listAssets(withWitness, includeSpent, includeLeased bool) (*taprpc.ListAssetResponse, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := taprpc.NewTaprootAssetsClient(conn)
	request := &taprpc.ListAssetRequest{
		WithWitness:             withWitness,
		IncludeSpent:            includeSpent,
		IncludeLeased:           includeLeased,
		IncludeUnconfirmedMints: true,
	}
	response, err := client.ListAssets(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "ListAssets")
	}
	return response, nil
}

func listBalances(isGroupByAssetIdOrGroupKey bool) (*taprpc.ListBalancesResponse, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := taprpc.NewTaprootAssetsClient(conn)
	var request *taprpc.ListBalancesRequest
	if isGroupByAssetIdOrGroupKey {
		request = &taprpc.ListBalancesRequest{
			GroupBy: &taprpc.ListBalancesRequest_AssetId{AssetId: true},
		}
	} else {
		request = &taprpc.ListBalancesRequest{
			GroupBy: &taprpc.ListBalancesRequest_GroupKey{GroupKey: true},
		}
	}
	response, err := client.ListBalances(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "ListBalances")
	}
	return response, nil
}

func listTransfers() (*taprpc.ListTransfersResponse, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := taprpc.NewTaprootAssetsClient(conn)
	request := &taprpc.ListTransfersRequest{}
	response, err := client.ListTransfers(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "ListTransfers")
	}
	return response, nil
}

func syncUniverse(universeHost string, syncTargets []*universerpc.SyncTarget, syncMode universerpc.UniverseSyncMode) (*universerpc.SyncResponse, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	request := &universerpc.SyncRequest{
		UniverseHost: universeHost,
		SyncMode:     syncMode,
		SyncTargets:  syncTargets,
	}
	client := universerpc.NewUniverseClient(conn)
	response, err := client.SyncUniverse(context.Background(), request)
	return response, err
}

func addrReceives() (*taprpc.AddrReceivesResponse, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	request := &taprpc.AddrReceivesRequest{}
	client := taprpc.NewTaprootAssetsClient(conn)
	response, err := client.AddrReceives(context.Background(), request)
	return response, err
}

func listUtxos() (*taprpc.ListUtxosResponse, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	request := &taprpc.ListUtxosRequest{}
	client := taprpc.NewTaprootAssetsClient(conn)
	response, err := client.ListUtxos(context.Background(), request)
	return response, err
}

func fetchAssetMetaByAssetId(assetId string) (*taprpc.AssetMeta, error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	request := &taprpc.FetchAssetMetaRequest{
		Asset: &taprpc.FetchAssetMetaRequest_AssetIdStr{
			AssetIdStr: assetId,
		},
	}
	client := taprpc.NewTaprootAssetsClient(conn)
	response, err := client.FetchAssetMeta(context.Background(), request)
	return response, err
}

func fetchAssetMetaByAssetIds(assetIds []string) (*map[string]string, *map[string]error) {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := taprpc.NewTaprootAssetsClient(conn)
	idMapData := make(map[string]string)
	idMapErr := make(map[string]error)
	for _, assetId := range assetIds {
		request := &taprpc.FetchAssetMetaRequest{
			Asset: &taprpc.FetchAssetMetaRequest_AssetIdStr{
				AssetIdStr: assetId,
			},
		}
		response, err := client.FetchAssetMeta(context.Background(), request)
		if err != nil {
			idMapErr[assetId] = err

		}
		if response == nil {
			idMapData[assetId] = ""
		} else {
			idMapData[assetId] = string(response.Data)
		}
	}
	return &idMapData, &idMapErr
}

func queryAssetRoots(assetId string) *universerpc.QueryRootResponse {
	grpcHost := config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
	tlsCertPath := config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
	macaroonPath := config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	conn, connClose := utils.GetConn(grpcHost, tlsCertPath, macaroonPath)
	defer connClose()
	client := universerpc.NewUniverseClient(conn)
	in := &universerpc.AssetRootQuery{}
	in.Id = &universerpc.ID{
		Id: &universerpc.ID_AssetIdStr{
			AssetIdStr: assetId,
		},
	}
	roots, err := client.QueryAssetRoots(context.Background(), in)
	if err != nil {
		return nil
	}
	return roots
}

func fundChannel(assetId string, assetAmount int, peerPubkey string, feeRate int, pushSat int, localAmt int) (*btlchannelrpc.FundBtlChannelResponse, error) {
	connConfiguration := GetConnConfiguration(ClientTypeTapd)
	conn, connClose := utils.GetConn(connConfiguration.GrpcHost, connConfiguration.TlsCertPath, connConfiguration.MacaroonPath)
	defer connClose()
	assetIdBytes, err := hex.DecodeString(assetId)
	if err != nil {
		return nil, err
	}

	PubKey, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return nil, err
	}

	if pushSat != 0 {
		pushSat = 0
	}

	client := btlchannelrpc.NewBtlChannelsClient(conn)
	req := &btlchannelrpc.FundBtlChannelRequest{
		AssetAmount:        uint64(assetAmount),
		AssetId:            assetIdBytes,
		PeerPubkey:         PubKey,
		FeeRateSatPerVbyte: uint32(feeRate),
		LocalAmt:           uint64(localAmt),
	}

	response, err := client.FundBtlChannel(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return response, err
}

func boxFundChannel(assetId string, assetAmount int64, peerPubkey string, feeRate int, pushSat int, localAmt int, serverIdentityPubkey string) (*btlchannelrpc.FundBtlChannelResponse, error) {

	node, err := nodemanage.GetNodeCoonPubKey(serverIdentityPubkey)
	if err != nil {
		return nil, err
	}
	assetIdBytes, err := hex.DecodeString(assetId)
	if err != nil {
		return nil, err
	}

	PubKey, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return nil, err
	}

	client := btlchannelrpc.NewBtlChannelsClient(node.TapCon)
	req := &btlchannelrpc.FundBtlChannelRequest{
		AssetAmount:        uint64(assetAmount),
		AssetId:            assetIdBytes,
		PeerPubkey:         PubKey,
		FeeRateSatPerVbyte: uint32(feeRate),
		LocalAmt:           uint64(localAmt),
	}

	if pushSat != 0 {
		req.PushSat = int64(pushSat)
	}
	btlLog.BoxAssetPush.Info("\nboxFundChannel\n %v", req)
	response, err := client.FundBtlChannel(context.Background(), req)
	if err != nil {
		btlLog.BoxAssetPush.Error("\nboxFundChannel\n %v", err)
		return nil, err
	}
	return response, err
}

func keySendToAssetChannel(assetId string, amount int64, pubkey string, outgoingChanId int) (*lnrpc.Payment, error) {
	connConfiguration := GetConnConfiguration(ClientTypeTapd)
	conn, connClose := utils.GetConn(connConfiguration.GrpcHost, connConfiguration.TlsCertPath, connConfiguration.MacaroonPath)
	defer connClose()

	if assetId == "" || amount == 0 || pubkey == "" {
		return nil, errors.New("invalid params")
	}

	client := tapchannelrpc.NewTaprootAssetChannelsClient(conn)

	assetIdStr, err := hex.DecodeString(assetId)
	if err != nil {
		return nil, err
	}

	peerPubkey, err := hex.DecodeString(pubkey)
	if err != nil {
		return nil, err
	}

	req := &tapchannelrpc.SendPaymentRequest{
		AssetId:     assetIdStr,
		AssetAmount: uint64(amount),
		PaymentRequest: &routerrpc.SendPaymentRequest{
			Dest:              peerPubkey,
			Amt:               int64(rfqmath.DefaultOnChainHtlcSat),
			TimeoutSeconds:    30,
			DestCustomRecords: make(map[uint64][]byte),
		},
	}

	if outgoingChanId != 0 {
		req.PaymentRequest.OutgoingChanIds = []uint64{uint64(outgoingChanId)}
	}

	destRecords := req.PaymentRequest.DestCustomRecords
	_, isKeysend := destRecords[record.KeySendType]
	var rHash []byte
	var preimage lntypes.Preimage
	if _, err := rand.Read(preimage[:]); err != nil {
		return nil, err
	}
	if !isKeysend {
		destRecords[record.KeySendType] = preimage[:]
		hash := preimage.Hash()
		rHash = hash[:]

		req.PaymentRequest.PaymentHash = rHash

	}
	resp, err := client.SendPayment(context.Background(), req)
	if err != nil {
		btlLog.BoxAssetPush.Error("\nkeysend\n %v", err)
		return nil, err
	}
	for {
		resp1, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				btlLog.BoxAssetPush.Error("\nkeysend io.EOF\n %v", err)
				return nil, err
			}
			btlLog.BoxAssetPush.Error("\nkeysend stream Recv err:\n %v", err)
			return nil, err
		} else if resp1 != nil {
			resp2 := resp1.GetPaymentResult()
			if resp2 != nil {
				if resp2.Status == 2 {
					return resp2, nil
				} else if resp2.Status == 3 {
					return nil, err
				}
			}
		}
	}
}

func boxKeySendToAssetChannel(client tapchannelrpc.TaprootAssetChannelsClient, assetId string, amount int64, pubkey string, outgoingChanId int64, serverIdentityPubkey string) (*lnrpc.Payment, error) {

	if assetId == "" || amount == 0 || pubkey == "" {
		return nil, errors.New("invalid params")
	}

	assetIdStr, err := hex.DecodeString(assetId)
	if err != nil {
		return nil, err
	}

	peerPubkey, err := hex.DecodeString(pubkey)
	if err != nil {
		return nil, err
	}

	req := &tapchannelrpc.SendPaymentRequest{
		AssetId:     assetIdStr,
		AssetAmount: uint64(amount),
		PaymentRequest: &routerrpc.SendPaymentRequest{
			Dest:              peerPubkey,
			Amt:               int64(rfqmath.DefaultOnChainHtlcSat),
			TimeoutSeconds:    30,
			DestCustomRecords: make(map[uint64][]byte),
		},
	}

	if outgoingChanId != 0 {
		req.PaymentRequest.OutgoingChanIds = []uint64{uint64(outgoingChanId)}
	}

	destRecords := req.PaymentRequest.DestCustomRecords
	_, isKeysend := destRecords[record.KeySendType]
	var rHash []byte
	var preimage lntypes.Preimage
	if _, err := rand.Read(preimage[:]); err != nil {
		return nil, err
	}
	if !isKeysend {
		destRecords[record.KeySendType] = preimage[:]
		hash := preimage.Hash()
		rHash = hash[:]

		req.PaymentRequest.PaymentHash = rHash

	}
	resp, err := client.SendPayment(context.Background(), req)
	if err != nil {
		btlLog.BoxAssetPush.Error("\nkeysend\n %v", err)
		return nil, err
	}
	for {
		resp1, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				btlLog.BoxAssetPush.Error("\nkeysend io.EOF\n %v", err)
				return nil, err
			}
			return nil, err
		} else if resp1 != nil {
			resp2 := resp1.GetPaymentResult()
			if resp2 != nil {
				if resp2.Status == 2 {
					return resp2, nil
				} else if resp2.Status == 3 {
					return nil, err
				}
			}
		}
	}
}

func boxGetTapAddrs(serverIdentityPubkey string, assetId string, amt int64) (string, error) {
	assetIdStr, err := hex.DecodeString(assetId)
	if err != nil {
		return "", err
	}
	node, err := nodemanage.GetNodeCoonPubKey(serverIdentityPubkey)
	if err != nil {
		return "", err
	}

	client := taprpc.NewTaprootAssetsClient(node.TapCon)
	resp, err := client.NewAddr(context.Background(), &taprpc.NewAddrRequest{
		AssetId: assetIdStr,
		Amt:     uint64(amt),
	})
	if err != nil {
		return "", err
	}
	return resp.Encoded, nil
}

func boxAddrReceives(serverIdentityPubkey string, assetAddr string) (*taprpc.AddrReceivesResponse, error) {
	node, err := nodemanage.GetNodeCoonPubKey(serverIdentityPubkey)
	if err != nil {
		return nil, err
	}
	client := taprpc.NewTaprootAssetsClient(node.TapCon)
	resp, err := client.AddrReceives(context.Background(), &taprpc.AddrReceivesRequest{
		FilterAddr: assetAddr,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
