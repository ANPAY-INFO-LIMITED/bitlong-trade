package api

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"trade/btlLog"
	"trade/config"
	"trade/services/nodemanage"
	"trade/utils"

	"github.com/lightningnetwork/lnd/lnrpc"
)

type ClientType int

var (
	ClientTypeLnd  ClientType = 1
	ClientTypeTapd ClientType = 2
	ClientTypeLitd ClientType = 3
)

type ConnConfiguration struct {
	GrpcHost     string
	TlsCertPath  string
	MacaroonPath string
}

func GetConnConfiguration(clientType ClientType) *ConnConfiguration {
	var connConfiguration ConnConfiguration
	if clientType == ClientTypeLnd {
		connConfiguration.GrpcHost = config.GetLoadConfig().ApiConfig.Lnd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Lnd.Port)
		connConfiguration.TlsCertPath = config.GetLoadConfig().ApiConfig.Lnd.TlsCertPath
		connConfiguration.MacaroonPath = config.GetLoadConfig().ApiConfig.Lnd.MacaroonPath
	} else if clientType == ClientTypeTapd {
		connConfiguration.GrpcHost = config.GetLoadConfig().ApiConfig.Tapd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Tapd.Port)
		connConfiguration.TlsCertPath = config.GetLoadConfig().ApiConfig.Tapd.TlsCertPath
		connConfiguration.MacaroonPath = config.GetLoadConfig().ApiConfig.Tapd.MacaroonPath
	} else if clientType == ClientTypeLitd {
		connConfiguration.GrpcHost = config.GetLoadConfig().ApiConfig.Litd.Host + ":" + strconv.Itoa(config.GetLoadConfig().ApiConfig.Litd.Port)
		connConfiguration.TlsCertPath = config.GetLoadConfig().ApiConfig.Litd.TlsCertPath
		connConfiguration.MacaroonPath = config.GetLoadConfig().ApiConfig.Litd.MacaroonPath
	} else {
		return nil
	}
	return &connConfiguration
}

func listChainTxns() (*lnrpc.TransactionDetails, error) {
	connConfiguration := GetConnConfiguration(ClientTypeLnd)
	conn, connClose := utils.GetConn(connConfiguration.GrpcHost, connConfiguration.TlsCertPath, connConfiguration.MacaroonPath)
	defer connClose()
	client := lnrpc.NewLightningClient(conn)
	request := &lnrpc.GetTransactionsRequest{}
	response, err := client.GetTransactions(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "GetTransactions")
	}
	return response, nil
}

func walletBalance() (*lnrpc.WalletBalanceResponse, error) {
	connConfiguration := GetConnConfiguration(ClientTypeLnd)
	conn, connClose := utils.GetConn(connConfiguration.GrpcHost, connConfiguration.TlsCertPath, connConfiguration.MacaroonPath)
	defer connClose()
	client := lnrpc.NewLightningClient(conn)
	request := &lnrpc.WalletBalanceRequest{}
	response, err := client.WalletBalance(context.Background(), request)
	if err != nil {
		return nil, utils.AppendErrorInfo(err, "WalletBalance")
	}
	return response, nil
}

type ChainTransaction struct {
	TxHash            string             `json:"tx_hash"`
	Amount            int                `json:"amount"`
	NumConfirmations  int                `json:"num_confirmations"`
	BlockHash         string             `json:"block_hash"`
	BlockHeight       int                `json:"block_height"`
	TimeStamp         int                `json:"time_stamp"`
	TotalFees         int                `json:"total_fees"`
	DestAddresses     []string           `json:"dest_addresses"`
	OutputDetails     []OutputDetail     `json:"output_details"`
	RawTxHex          string             `json:"raw_tx_hex"`
	Label             string             `json:"label"`
	PreviousOutpoints []PreviousOutpoint `json:"previous_outpoints"`
}

type OutputDetail struct {
	OutputType   string `json:"output_type"`
	Address      string `json:"address"`
	PkScript     string `json:"pk_script"`
	OutputIndex  int    `json:"output_index"`
	Amount       int    `json:"amount"`
	IsOurAddress bool   `json:"is_our_address"`
}

type PreviousOutpoint struct {
	Outpoint    string `json:"outpoint"`
	IsOurOutput bool   `json:"is_our_output"`
}

func processChainTransactions(response *lnrpc.TransactionDetails) *[]ChainTransaction {
	var chainTransactions []ChainTransaction
	for _, transaction := range response.Transactions {
		var outputDetails []OutputDetail
		for _, outputDetail := range transaction.OutputDetails {
			outputDetails = append(outputDetails, OutputDetail{
				OutputType:   outputDetail.OutputType.String(),
				Address:      outputDetail.Address,
				PkScript:     outputDetail.PkScript,
				OutputIndex:  int(outputDetail.OutputIndex),
				Amount:       int(outputDetail.Amount),
				IsOurAddress: outputDetail.IsOurAddress,
			})
		}
		var previousOutpoints []PreviousOutpoint
		for _, previousOutpoint := range transaction.PreviousOutpoints {
			previousOutpoints = append(previousOutpoints, PreviousOutpoint{
				Outpoint:    previousOutpoint.Outpoint,
				IsOurOutput: previousOutpoint.IsOurOutput,
			})
		}
		chainTransactions = append(chainTransactions, ChainTransaction{
			TxHash:            transaction.TxHash,
			Amount:            int(transaction.Amount),
			NumConfirmations:  int(transaction.NumConfirmations),
			BlockHash:         transaction.BlockHash,
			BlockHeight:       int(transaction.BlockHeight),
			TimeStamp:         int(transaction.TimeStamp),
			TotalFees:         int(transaction.TotalFees),
			DestAddresses:     transaction.GetDestAddresses(),
			OutputDetails:     outputDetails,
			RawTxHex:          transaction.RawTxHex,
			Label:             transaction.Label,
			PreviousOutpoints: previousOutpoints,
		})
	}
	return &chainTransactions
}

func openBtcChannel(amount int, peerPubkey string, feeRate int, pushSat int) (string, error) {
	connConfiguration := GetConnConfiguration(ClientTypeLnd)
	conn, connClose := utils.GetConn(connConfiguration.GrpcHost, connConfiguration.TlsCertPath, connConfiguration.MacaroonPath)
	defer connClose()

	PubKey, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return "", err
	}
	if pushSat != 0 {
		pushSat = 0
	}
	client := lnrpc.NewLightningClient(conn)
	stream, err := client.OpenChannel(context.Background(), &lnrpc.OpenChannelRequest{
		SatPerVbyte:        uint64(feeRate),
		NodePubkey:         PubKey,
		LocalFundingAmount: int64(amount),
		PushSat:            int64(pushSat),
	})
	if err != nil {
		return "", err
	}
	conversion := func(b []byte) string {
		for i := 0; i < len(b)/2; i++ {
			temp := b[i]
			b[i] = b[len(b)-i-1]
			b[len(b)-i-1] = temp
		}
		txHash := hex.EncodeToString(b)
		return txHash
	}
	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return "", err
			}
			return "", err
		} else if response.PendingChanId != nil {
			txid := response.GetChanPending().Txid
			hash := conversion(txid)
			OutputIndex := response.GetChanPending().OutputIndex
			resp := fmt.Sprintf("%s:%d", hash, OutputIndex)
			return resp, nil
		}
	}
}

func getChannelList() (*lnrpc.ListChannelsResponse, error) {
	connConfiguration := GetConnConfiguration(ClientTypeLnd)
	conn, connClose := utils.GetConn(connConfiguration.GrpcHost, connConfiguration.TlsCertPath, connConfiguration.MacaroonPath)
	defer connClose()

	client := lnrpc.NewLightningClient(conn)
	resp, err := client.ListChannels(context.Background(), &lnrpc.ListChannelsRequest{})
	if err != nil {
		btlLog.OpenChannel.Error("\ngetChannelList\n %v", err)
		return nil, err
	}

	return resp, nil
}

func getRelatedChannelList(serverIdentityPubkey string) (*lnrpc.ListChannelsResponse, error) {

	node, err := nodemanage.GetNodeCoonPubKey(serverIdentityPubkey)
	if err != nil {
		return nil, err
	}

	client := lnrpc.NewLightningClient(node.LndCon)
	resp, err := client.ListChannels(context.Background(), &lnrpc.ListChannelsRequest{})
	if err != nil {
		btlLog.BoxAssetPush.Error("\ngetRelatedChannelList\n %v", err)
		return nil, err
	}

	return resp, nil
}

func getChannelPeer() (*lnrpc.ListPeersResponse, error) {
	connConfiguration := GetConnConfiguration(ClientTypeLnd)
	conn, connClose := utils.GetConn(connConfiguration.GrpcHost, connConfiguration.TlsCertPath, connConfiguration.MacaroonPath)
	defer connClose()

	client := lnrpc.NewLightningClient(conn)
	peers, err := client.ListPeers(context.Background(), &lnrpc.ListPeersRequest{})
	if err != nil {
		btlLog.OpenChannel.Error("\ngetChannelPeer\n %v", err)
		return nil, err
	}
	return peers, nil
}

func boxPendingChannels(serverIdentityPubkey string) (*lnrpc.PendingChannelsResponse, error) {

	node, err := nodemanage.GetNodeCoonPubKey(serverIdentityPubkey)
	if err != nil {
		return nil, err
	}

	client := lnrpc.NewLightningClient(node.LndCon)
	response, err := client.PendingChannels(context.Background(), &lnrpc.PendingChannelsRequest{})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func getAssetChannelListToBox() (*lnrpc.ListChannelsResponse, error) {
	connConfiguration := BoxBoxConnConfiguration(ClientTypeLnd)
	conn, connClose := utils.GetConn(connConfiguration.GrpcHost, connConfiguration.TlsCertPath, connConfiguration.MacaroonPath)
	defer connClose()

	client := lnrpc.NewLightningClient(conn)
	resp, err := client.ListChannels(context.Background(), &lnrpc.ListChannelsRequest{
		PrivateOnly: true,
	})
	if err != nil {
		btlLog.OpenChannel.Error("\ngetAssetChannelList\n %v", err)
		return nil, err
	}

	return resp, nil
}

func getChannelListToBox(serverIdentityPubkey string) (*lnrpc.ListChannelsResponse, error) {

	node, err := nodemanage.GetNodeCoonPubKey(serverIdentityPubkey)
	if err != nil {
		return nil, err
	}

	client := lnrpc.NewLightningClient(node.LndCon)
	resp, err := client.ListChannels(context.Background(), &lnrpc.ListChannelsRequest{})
	if err != nil {
		btlLog.OpenChannel.Error("\ngetChannelList\n %v", err)
		return nil, err
	}

	return resp, nil
}

func boxClosedChannels(serverIdentityPubkey string) (*lnrpc.ClosedChannelsResponse, error) {

	node, err := nodemanage.GetNodeCoonPubKey(serverIdentityPubkey)
	if err != nil {
		return nil, err
	}

	client := lnrpc.NewLightningClient(node.LndCon)
	resp, err := client.ClosedChannels(context.Background(), &lnrpc.ClosedChannelsRequest{})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func ServerBalance(serverIdentityPubkey string) (int64, error) {
	node, err := nodemanage.GetNodeCoonPubKey(serverIdentityPubkey)
	if err != nil {
		return 0, err
	}

	client := lnrpc.NewLightningClient(node.LndCon)
	resp, err := client.WalletBalance(context.Background(), &lnrpc.WalletBalanceRequest{})
	if err != nil {
		return 0, err
	}
	return resp.ConfirmedBalance, nil
}

func getServerStatus(serverIdentityPubkey string) (*lnrpc.GetStateResponse, error) {
	node, err := nodemanage.GetNodeCoonPubKey(serverIdentityPubkey)
	if err != nil {
		return nil, err
	}

	client := lnrpc.NewStateClient(node.LndCon)
	resp, err := client.GetState(context.Background(), &lnrpc.GetStateRequest{})
	if err != nil {
		return nil, err
	}

	return resp, err
}
