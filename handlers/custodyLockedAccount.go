package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"trade/models"
	"trade/services/custodyAccount"
	"trade/services/custodyAccount/defaultAccount/custodyAssets"
)

func QueryLockedPayments(c *gin.Context) {

	userName := c.MustGet("username").(string)
	invoiceRequest := struct {
		AssetId string `json:"asset_id"`
		Page    int    `json:"page"`
		Size    int    `json:"size"`
		Away    int    `json:"away"`
	}{}
	if err := c.ShouldBindJSON(&invoiceRequest); err != nil {
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	e, err := custodyAssets.NewAssetEvent(userName, invoiceRequest.AssetId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.MakeJsonErrorResultForHttp(models.DefaultErr, "用户不存在", nil))
		return
	}

	p, err := custodyAccount.LockPaymentToPaymentList(e.UserInfo, invoiceRequest.AssetId, invoiceRequest.Page, invoiceRequest.Size, invoiceRequest.Away)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.DefaultErr, err.Error(), nil))
		return
	}
	p.Sort()

	c.JSON(http.StatusOK, models.MakeJsonErrorResultForHttp(models.SUCCESS, "", p))
}
