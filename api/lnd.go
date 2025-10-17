package api

import (
	"encoding/json"
	"errors"
	"strconv"
	"trade/btlLog"
	"trade/utils"

	"github.com/lightninglabs/taproot-assets/rfqmsg"
	"github.com/lightningnetwork/lnd/lnrpc"
)

func ListChainTxnsAndGetResponse() (*lnrpc.TransactionDetails, error) {
	return listChainTxns()
}

func GetListChainTransactions() (*[]ChainTransaction, error) {
	response, err := listChainTxns()
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "listChainTxns")
	}
	result := processChainTransactions(response)
	return result, nil
}

func WalletBalanceAndGetResponse() (*lnrpc.WalletBalanceResponse, error) {
	return walletBalance()
}

func GetListChainTransactionsOutpointAddress(outpoint string) (address string, err error) {
	response, err := GetListChainTransactions()
	if err != nil {
		return "", utils.AppendErrorInfo(err, "GetListChainTransactions")
	}
	tx, indexStr := utils.GetTransactionAndIndexByOutpoint(outpoint)
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return "", utils.AppendErrorInfo(err, "GetTransactionAndIndexByOutpoint")
	}
	for _, transaction := range *response {
		if transaction.TxHash == tx {
			return transaction.DestAddresses[index], nil
		}
	}
	err = errors.New("did not match transaction outpoint")
	return "", err
}

func OpenBtcChannel(amount int, peerPubkey string, feeRate int, pushSat int) (string, error) {
	return openBtcChannel(amount, peerPubkey, feeRate, pushSat)
}

type ChannelIdsAndPoints struct {
	BtcChanIDs      []uint64
	BtcChanPoints   []string
	AssetChanIDs    []uint64
	AssetChanPoints []string
}

func GetChannelIdsAndPoints(assetId string, peerPubkey string) (*ChannelIdsAndPoints, error) {
	resp, err := getChannelList()
	if err != nil {
		return nil, err
	}

	result := &ChannelIdsAndPoints{
		BtcChanIDs:      make([]uint64, 0, len(resp.Channels)),
		BtcChanPoints:   make([]string, 0, len(resp.Channels)),
		AssetChanIDs:    make([]uint64, 0, len(resp.Channels)),
		AssetChanPoints: make([]string, 0, len(resp.Channels)),
	}
	for _, channel := range resp.Channels {
		if channel.RemotePubkey != peerPubkey {
			continue
		}
		switch {
		case assetId == "00" && channel.CustomChannelData == nil:
			result.BtcChanIDs = append(result.BtcChanIDs, channel.ChanId)
			result.BtcChanPoints = append(result.BtcChanPoints, channel.ChannelPoint)

		case channel.CustomChannelData != nil:
			var customData rfqmsg.JsonAssetChannel
			if err := json.Unmarshal(channel.CustomChannelData, &customData); err != nil {
				btlLog.OpenChannel.Error("\n chanlist unmarshal \n %v", err)
				continue
			}

			if len(customData.FundingAssets) == 0 {
				continue
			}
			if customData.FundingAssets[0].AssetGenesis.AssetID == assetId {
				result.AssetChanIDs = append(result.AssetChanIDs, channel.ChanId)
				result.AssetChanPoints = append(result.AssetChanPoints, channel.ChannelPoint)
			}
		}
	}

	return result, nil
}

func GetChannelPeer(peerPubkey string) (bool, error) {
	peers, err := getChannelPeer()
	if err != nil {
		return false, err
	}
	for _, peer := range peers.Peers {
		if peer.PubKey == peerPubkey {
			return true, nil
		}
	}
	return false, nil
}

func OnlyBtcChannelNode(serverIdentityPubkey string) ([]string, error) {
	resp, err := getRelatedChannelList(serverIdentityPubkey)
	if err != nil {
		return nil, err
	}

	pubkeyCount := make(map[string]int)
	assetPubkeys := make(map[string]bool)

	for _, channel := range resp.Channels {
		pubkey := channel.RemotePubkey
		pubkeyCount[pubkey]++
		if channel.CustomChannelData != nil {
			assetPubkeys[pubkey] = true
		}
	}

	var btcOnlyPubkeys []string
	for pubkey, count := range pubkeyCount {
		if count == 1 && !assetPubkeys[pubkey] {
			btcOnlyPubkeys = append(btcOnlyPubkeys, pubkey)
		}
	}

	return btcOnlyPubkeys, nil
}

type GetBoxAssetInfo struct {
	IsActive     bool
	ChannelPoint string
	RemotePubkey string
	ChanId       uint64
}

func GetBoxAssetChannels1(serverIdentityPubkey string) ([]GetBoxAssetInfo, error) {
	resp, err := getRelatedChannelList(serverIdentityPubkey)
	if err != nil {
		return nil, err
	}

	pubkeyCount := make(map[string]int)
	assetChannelCount := make(map[string]int)
	for _, channel := range resp.Channels {
		pubkeyCount[channel.RemotePubkey]++
		if channel.CustomChannelData != nil {
			assetChannelCount[channel.RemotePubkey]++
		}
	}

	var assetChannels []GetBoxAssetInfo
	for _, channel := range resp.Channels {
		if channel.CustomChannelData != nil &&
			(assetChannelCount[channel.RemotePubkey] == 1 ||
				(pubkeyCount[channel.RemotePubkey] == 2 && assetChannelCount[channel.RemotePubkey] == 1)) {
			assetChannels = append(assetChannels, GetBoxAssetInfo{
				IsActive:     channel.Active,
				ChannelPoint: channel.ChannelPoint,
				RemotePubkey: channel.RemotePubkey,
				ChanId:       channel.ChanId,
			})
		}
	}

	return assetChannels, nil
}

func GetBoxAssetChannels(serverIdentityPubkey string) ([]GetBoxAssetInfo, error) {
	resp, err := getRelatedChannelList(serverIdentityPubkey)
	if err != nil {
		return nil, err
	}

	pubkeyCount := make(map[string]int)
	assetChannelCount := make(map[string]int)
	for _, channel := range resp.Channels {
		pubkeyCount[channel.RemotePubkey]++
		if channel.CustomChannelData != nil {
			assetChannelCount[channel.RemotePubkey]++
		}
	}

	var assetChannels []GetBoxAssetInfo
	for _, channel := range resp.Channels {
		if channel.CustomChannelData != nil &&
			pubkeyCount[channel.RemotePubkey] == 2 &&
			assetChannelCount[channel.RemotePubkey] == 1 {
			assetChannels = append(assetChannels, GetBoxAssetInfo{
				IsActive:     channel.Active,
				ChannelPoint: channel.ChannelPoint,
				RemotePubkey: channel.RemotePubkey,
				ChanId:       channel.ChanId,
			})
		}
	}

	return assetChannels, nil
}

func BoxPendingChannels(serverIdentityPubkey string) (*lnrpc.PendingChannelsResponse, error) {
	return boxPendingChannels(serverIdentityPubkey)
}

func GetListAssetChannelsToBox() (*lnrpc.ListChannelsResponse, error) {
	return getAssetChannelListToBox()
}

func GetListChannelsToBox(serverIdentityPubkey string) (*lnrpc.ListChannelsResponse, error) {
	return getChannelListToBox(serverIdentityPubkey)
}

func BoxClosedChannels(serverIdentityPubkey string) (*lnrpc.ClosedChannelsResponse, error) {
	return boxClosedChannels(serverIdentityPubkey)
}

func GetServerSatsBalance(serverIdentityPubkey string) (int64, error) {
	return ServerBalance(serverIdentityPubkey)
}

func GetServerStatus(serverIdentityPubkey string) (string, error) {
	resp, err := getServerStatus(serverIdentityPubkey)
	if err != nil {
		return "", err
	}
	if resp.State == lnrpc.WalletState_SERVER_ACTIVE {
		return "SERVER_ACTIVE", nil
	}
	return "", errors.New("server status is not active")
}
