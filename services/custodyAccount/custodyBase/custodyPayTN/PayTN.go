package custodyPayTN

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

type PayToNpubKey struct {
	NpubKey     string
	AssetId     string
	Amount      float64
	Time        int64
	FromNpubKey string
	Vision      uint8
}

func (p *PayToNpubKey) Encode() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	hexData := hex.EncodeToString(data)
	return "ptn" + hexData, nil
}
func (p *PayToNpubKey) Decode(encoded string) error {
	if !strings.HasPrefix(encoded, "ptn") {
		return errors.New("无效的编码字符串: 缺少前缀 'ptn'")
	}

	hexData := encoded[3:]
	data, err := hex.DecodeString(hexData)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, p)
	if err != nil {
		return err
	}
	if p.Vision == 0 {
		if p.FromNpubKey == "" {
			p.FromNpubKey = encoded
		}
	}
	return nil
}

func HashEncodedString(encoded string) (string, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(encoded))
	if err != nil {
		return "", err
	}
	hashedBytes := hash.Sum(nil)
	return hex.EncodeToString(hashedBytes), nil
}

type PayToNpubKeySwap struct {
	PayAssetId     string
	PayAmount      float64
	ReceiveAssetId string
	ReceiveAmount  float64
	Time           int64
	TargetNpubKey  string
	Vision         uint8
}

func (p *PayToNpubKeySwap) Encode() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	hexData := hex.EncodeToString(data)
	return "ptns" + hexData, nil
}
func (p *PayToNpubKeySwap) Decode(encoded string) error {
	if !strings.HasPrefix(encoded, "ptns") {
		return errors.New("无效的编码字符串: 缺少前缀 'ptns'")
	}
	hexData := encoded[4:]
	data, err := hex.DecodeString(hexData)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, p)
	if err != nil {
		return err
	}
	return nil
}

type SupplierId struct {
	Supplier string
	AssetId  string
	Amount   float64
	Price    float64
	Time     int64
	Vision   uint8
}

func (p *SupplierId) Encode() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	hexData := hex.EncodeToString(data)
	return "spl" + hexData, nil
}
func (p *SupplierId) Decode(encoded string) error {
	if !strings.HasPrefix(encoded, "spl") {
		return errors.New("无效的编码字符串: 缺少前缀 'spl'")
	}
	hexData := encoded[3:]
	data, err := hex.DecodeString(hexData)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, p)
	if err != nil {
		return err
	}
	return nil
}
