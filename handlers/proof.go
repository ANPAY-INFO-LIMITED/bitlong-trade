package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"trade/models"
	"trade/services"
	"trade/services/assetsyncinfo"
)

func DownloadProof(c *gin.Context) {
	AssetId := c.Param("asset_id")
	ProofName := c.Param("proof_name")
	path, err := services.ValidateAndGetProofFilePath(AssetId, ProofName)
	if err != nil {
		c.JSON(http.StatusOK, models.JsonResult{
			Success: false,
			Error:   "Validate And Get Proof File Path. " + err.Error(),
			Data:    nil,
			Code:    models.ValidateAndGetProofFilePathErr,
		})
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+ProofName)
	c.Header("Content-Disposition", "inline;filename="+ProofName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")
	c.File(path)
	return
}

func DownloadProof2(c *gin.Context) {
	AssetId := c.Param("asset_id")
	ProofName := c.Param("proof_name")
	path, err := services.ValidateAndGetProofFilePath(AssetId, ProofName)
	file, err := os.Open(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File not found"})
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)
	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+ProofName)
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.Writer.WriteHeader(http.StatusOK)
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func SyncAssetInfo(c *gin.Context) {
	req := assetsyncinfo.SyncInfoRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if req.Id == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "request error"})
		return
	}
	assetSyncInfo, err := assetsyncinfo.GetAssetSyncInfo(&req)
	if err != nil && errors.Is(err, assetsyncinfo.AssetNotFoundErr) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	r := struct {
		AssetId        string  `json:"asset_Id"`
		Name           string  `json:"name"`
		Point          string  `json:"point"`
		AssetType      string  `json:"assetType"`
		GroupName      *string `json:"group_name"`
		GroupKey       *string `json:"group_key"`
		Amount         uint64  `json:"amount"`
		Meta           *string `json:"meta"`
		CreateHeight   int64   `json:"create_height"`
		CreateTime     int64   `json:"create_time"`
		Universe       string  `json:"universe"`
		DecimalDisplay uint32  `json:"decimal_display"`
	}{
		AssetId:        assetSyncInfo.AssetId,
		Name:           assetSyncInfo.Name,
		Point:          assetSyncInfo.Point,
		AssetType:      models.AssetType_name[assetSyncInfo.AssetType],
		GroupName:      assetSyncInfo.GroupName,
		GroupKey:       assetSyncInfo.GroupKey,
		Amount:         assetSyncInfo.Amount,
		Meta:           assetSyncInfo.Meta,
		CreateHeight:   assetSyncInfo.CreateHeight,
		CreateTime:     assetSyncInfo.CreateTime.Unix(),
		Universe:       assetSyncInfo.Universe,
		DecimalDisplay: assetSyncInfo.DecimalDisplay,
	}
	c.JSON(http.StatusOK, gin.H{"sync_asset_info": r})
}

func GetAssetsDecimal(c *gin.Context) {
	var req []string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnauthorized, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	if len(req) == 0 {
		c.JSON(http.StatusUnauthorized, models.MakeJsonErrorResultForHttp(models.DefaultErr, "empty request", nil))
		return

	}
	f := func(slice []string) []string {
		keys := make(map[string]bool)
		list := []string{}

		for _, entry := range slice {
			if _, value := keys[entry]; !value {
				keys[entry] = true
				list = append(list, entry)
			}
		}
		return list
	}
	assetIds := f(req)
	infos, err := assetsyncinfo.GetAssetsDecimal(assetIds)
	if err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", infos))
}

func GetLastProof(c *gin.Context) {

	scriptKey := c.Query("script_key")
	outpoint := c.Query("outpoint")
	assetId := c.Query("asset_id")

	if scriptKey == "" {
		err := errors.New("script_key is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.QueryParamEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   "",
		})
		return
	}
	if outpoint == "" {
		err := errors.New("outpoint is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.QueryParamEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   "",
		})
		return
	}
	if assetId == "" {
		err := errors.New("asset_id is empty")
		c.JSON(http.StatusOK, Result2{
			Errno:  models.QueryParamEmptyErr.Code(),
			ErrMsg: err.Error(),
			Data:   "",
		})
		return
	}

	lastProofB64Str, err := services.GetLastProof(scriptKey, outpoint, assetId)
	if err != nil {
		c.JSON(http.StatusOK, Result2{
			Errno:  models.GetLastProofErr.Code(),
			ErrMsg: err.Error(),
			Data:   "",
		})
		return
	}

	c.JSON(http.StatusOK, Result2{
		Errno:  0,
		ErrMsg: models.SUCCESS.Error(),
		Data:   lastProofB64Str,
	})
}
