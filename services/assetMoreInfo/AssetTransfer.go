package assetMoreInfo

import (
	"trade/models"
	"trade/services/btldb"
)

func GetAssetTransferProcessedSliceByAssetIdLimit(assetId string, limit int) (*[]models.AssetTransferProcessedDb, error) {
	return btldb.ReadAssetTransferProcessedSliceByAssetIdLimit(assetId, limit)
}

func GetAssetTransferProcessedInputSliceByAssetId(assetId string) (*[]models.AssetTransferProcessedInputDb, error) {
	return btldb.ReadAssetTransferProcessedInputSliceByAssetId(assetId)
}

func GetAssetTransferProcessedOutputSliceByAssetId(assetId string) (*[]models.AssetTransferProcessedOutputDb, error) {
	return btldb.ReadAssetTransferProcessedOutputSliceByAssetId(assetId)
}

func GetInputsByTxidWithTransfersInputs(transfersInputs *[]models.AssetTransferProcessedInputDb, inputLength int, txid string) ([]models.AssetTransferProcessedInput, error) {
	result := make([]models.AssetTransferProcessedInput, inputLength)
	for _, input := range *transfersInputs {
		if input.Txid == txid && input.Index < inputLength {
			result[input.Index] = models.AssetTransferProcessedInput{
				Address:     input.Address,
				Amount:      input.Amount,
				AnchorPoint: input.AnchorPoint,
				ScriptKey:   input.ScriptKey,
			}
		}
	}
	return result, nil
}

func GetOutputsByTxidWithTransfersOutputs(transfersOutputs *[]models.AssetTransferProcessedOutputDb, outputLength int, txid string) ([]models.AssetTransferProcessedOutput, error) {
	result := make([]models.AssetTransferProcessedOutput, outputLength)
	for _, output := range *transfersOutputs {
		if output.Txid == txid && output.Index < outputLength {
			result[output.Index] = models.AssetTransferProcessedOutput{
				Address:                output.Address,
				Amount:                 output.Amount,
				AnchorOutpoint:         output.AnchorOutpoint,
				AnchorValue:            output.AnchorValue,
				AnchorInternalKey:      output.AnchorInternalKey,
				AnchorTaprootAssetRoot: output.AnchorTaprootAssetRoot,
				AnchorMerkleRoot:       output.AnchorMerkleRoot,
				AnchorTapscriptSibling: output.AnchorTapscriptSibling,
				AnchorNumPassiveAssets: output.AnchorNumPassiveAssets,
				ScriptKey:              output.ScriptKey,
				ScriptKeyIsLocal:       output.ScriptKeyIsLocal,
				NewProofBlob:           output.NewProofBlob,
				SplitCommitRootHash:    output.SplitCommitRootHash,
				OutputType:             output.OutputType,
				AssetVersion:           output.AssetVersion,
			}
		}
	}
	return result, nil
}

func CombineAssetTransfers(transfers *[]models.AssetTransferProcessedDb, transfersInputs *[]models.AssetTransferProcessedInputDb, transfersOutputs *[]models.AssetTransferProcessedOutputDb) (*[]models.AssetTransferProcessedCombined, error) {
	var err error
	var transferCombinedSlice []models.AssetTransferProcessedCombined
	for _, transfer := range *transfers {
		var transferCombined models.AssetTransferProcessedCombined
		inputs := make([]models.AssetTransferProcessedInput, transfer.Inputs)
		inputs, err = GetInputsByTxidWithTransfersInputs(transfersInputs, transfer.Inputs, transfer.Txid)
		if err != nil {
			return nil, err
		}
		outputs := make([]models.AssetTransferProcessedOutput, transfer.Outputs)
		outputs, err = GetOutputsByTxidWithTransfersOutputs(transfersOutputs, transfer.Outputs, transfer.Txid)
		transferCombined = models.AssetTransferProcessedCombined{
			Model:              transfer.Model,
			Txid:               transfer.Txid,
			AssetID:            transfer.AssetID,
			TransferTimestamp:  transfer.TransferTimestamp,
			AnchorTxHash:       transfer.AnchorTxHash,
			AnchorTxHeightHint: transfer.AnchorTxHeightHint,
			AnchorTxChainFees:  transfer.AnchorTxChainFees,
			Inputs:             inputs,
			Outputs:            outputs,
			DeviceID:           transfer.DeviceID,
			UserID:             transfer.UserID,
			Username:           transfer.Username,
			Status:             transfer.Status,
		}
		transferCombinedSlice = append(transferCombinedSlice, transferCombined)
	}
	return &transferCombinedSlice, nil
}

// 交易

func GetAssetTransferCombinedSliceByAssetIdLimit(assetId string, limit int) ([]models.AssetTransferProcessedCombined, error) {
	var err error
	var transferCombinedSlice *[]models.AssetTransferProcessedCombined
	// @dev: Use limit only here
	transfers, err := GetAssetTransferProcessedSliceByAssetIdLimit(assetId, limit)
	if err != nil {
		return nil, err
	}
	transfersInputs, err := GetAssetTransferProcessedInputSliceByAssetId(assetId)
	if err != nil {
		return nil, err
	}
	transfersOutputs, err := GetAssetTransferProcessedOutputSliceByAssetId(assetId)
	if err != nil {
		return nil, err
	}
	transferCombinedSlice, err = CombineAssetTransfers(transfers, transfersInputs, transfersOutputs)
	if err != nil {
		return nil, err
	}
	if transferCombinedSlice == nil {
		return []models.AssetTransferProcessedCombined{}, nil
	} else {
		return *transferCombinedSlice, nil
	}
}
