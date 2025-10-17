package models

import "gorm.io/gorm"

type AssetList struct {
	gorm.Model
	Version string `json:"version"`

	GenesisPoint string `json:"genesis_point"`
	Name         string `json:"name"`
	MetaHash     string `json:"meta_hash"`
	AssetID      string `json:"asset_id"`
	AssetType    string `json:"asset_type"`
	OutputIndex  int    `json:"output_index"`

	Amount           int   `json:"amount"`
	LockTime         int32 `json:"lock_time"`
	RelativeLockTime int32 `json:"relative_lock_time"`

	ScriptKey string `json:"script_key"`

	AnchorOutpoint string `json:"anchor_outpoint"`

	TweakedGroupKey string `json:"tweaked_group_key"`

	DeviceId string `json:"device_id" gorm:"type:varchar(255)"`
	UserId   int    `json:"user_id"`
	Username string `json:"username" gorm:"type:varchar(255)"`
}

type AssetListSetRequest struct {
	Version          string `json:"version" gorm:"type:varchar(255);index"`
	GenesisPoint     string `json:"genesis_point" gorm:"type:varchar(255)"`
	Name             string `json:"name" gorm:"type:varchar(255);index"`
	MetaHash         string `json:"meta_hash" gorm:"type:varchar(255);index"`
	AssetID          string `json:"asset_id" gorm:"type:varchar(255);index"`
	AssetType        string `json:"asset_type" gorm:"type:varchar(255);index"`
	OutputIndex      int    `json:"output_index"`
	Amount           int    `json:"amount"`
	LockTime         int32  `json:"lock_time"`
	RelativeLockTime int32  `json:"relative_lock_time"`
	ScriptKey        string `json:"script_key" gorm:"type:varchar(255);index"`
	AnchorOutpoint   string `json:"anchor_outpoint" gorm:"type:varchar(255);index"`
	TweakedGroupKey  string `json:"tweaked_group_key" gorm:"type:varchar(255);index"`
	DeviceId         string `json:"device_id" gorm:"type:varchar(255);index"`
}
