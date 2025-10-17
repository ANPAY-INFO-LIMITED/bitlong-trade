package btldb

import (
	"gorm.io/gorm"
	"sync"
	"trade/middleware"
	"trade/models"
)

var invoiceMutex sync.Mutex

func CreateInvoice(invoice *models.Invoice) error {
	invoiceMutex.Lock()
	defer invoiceMutex.Unlock()
	return middleware.DB.Create(invoice).Error
}

func GetInvoice(id uint) (*models.Invoice, error) {
	var invoice models.Invoice
	err := middleware.DB.First(&invoice, id).Error
	return &invoice, err
}

func GetInvoiceByReq(invoiceReq string) (*models.Invoice, error) {
	var invoice models.Invoice
	err := middleware.DB.Where("invoice =?", invoiceReq).First(&invoice).Error
	return &invoice, err
}

func UpdateInvoice(db *gorm.DB, invoice *models.Invoice) error {
	invoiceMutex.Lock()
	defer invoiceMutex.Unlock()
	return db.Save(invoice).Error
}

func DeleteInvoice(id uint) error {
	invoiceMutex.Lock()
	defer invoiceMutex.Unlock()
	var invoice models.Invoice
	return middleware.DB.Delete(&invoice, id).Error
}
