package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

type Wallets struct {
	Wallets map[string]*WalletGroup
}

var (
	_, b, _, _ = runtime.Caller(0)

	// Root folder of this project
	Root            = filepath.Join(filepath.Dir(b), "../")
	walletsPath     = path.Join(Root, "/storage/wallets")
	walletsFilename = "wallets.data"
)

func InitializeWallets() (*Wallets, error) {
	wallets := Wallets{map[string]*WalletGroup{}}
	err := wallets.LoadFile()

	return &wallets, err
}
func (ws *Wallets) GetWallet(userId string) (WalletGroup, error) {
	var wallet *WalletGroup
	var ok bool
	w := *ws
	if wallet, ok = w.Wallets[userId]; !ok {
		return *new(WalletGroup), errors.New("Invalid ID")
	}

	return *wallet, nil
}

func (ws *Wallets) AddWallet(userId string) string {
	wallet := MakeWalletGroup()
	userId = fmt.Sprintf("%s", userId)

	ws.Wallets[userId] = wallet

	return userId
}

func (ws *Wallets) LoadFile() error {
	walletsFile := path.Join(walletsPath, walletsFilename)

	if _, err := os.Stat(walletsFile); os.IsNotExist(err) {
		return err
	}
	var wallets Wallets
	fileContent, err := ioutil.ReadFile(walletsFile)
	if err != nil {
		return err
	}

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		return err
	}

	ws.Wallets = wallets.Wallets

	return nil
}
func (ws *Wallets) Save() {
	walletsFile := path.Join(walletsPath, walletsFilename)

	var content bytes.Buffer

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletsFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
