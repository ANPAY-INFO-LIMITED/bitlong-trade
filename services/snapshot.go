package services

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
)

const (
	LndMainnetDataPath = "/root/mainnet-lit/.lnd/data"
)

func SnapshotToZipLast() {

	filePaths := []string{
		LndMainnetDataPath + "/chain/bitcoin/mainnet/block_headers.bin",
		LndMainnetDataPath + "/chain/bitcoin/mainnet/neutrino.db",
		LndMainnetDataPath + "/chain/bitcoin/mainnet/reg_filter_headers.bin",
		LndMainnetDataPath + "/graph/mainnet/channel.db",
	}

	zipFilePath := "/root/neutrino/data.zip"

	errRemove := os.Remove(zipFilePath)
	if errRemove != nil {
		log.Println("delete snapshot zip err:", errRemove)
	} else {
		log.Println("delete snapshot zip ok")
	}

	file, errzip := os.Create(zipFilePath)
	if errzip != nil {
		log.Println("Create Zip File err: ", errzip)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	for _, filePath := range filePaths {

		fileName := filepath.Base(filePath)

		f, err := w.Create(fileName)
		if err != nil {
			log.Println(" w.Create err: ", err)
		}

		file, err := os.Open(filePath)
		if err != nil {
			log.Println("os.Open err: ", err)
		}
		defer file.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			log.Println("io.Copy err: ", err)
		}
	}
	log.Println("creat zip ok")
}
