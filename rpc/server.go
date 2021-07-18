package rpc

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	jrpc "github.com/gumeniukcom/golang-jsonrpc2"
	logger "github.com/sirupsen/logrus"
	blockchain "github.com/workspace/evoting/ev-blockchain-protocol/core"
	"github.com/workspace/evoting/ev-blockchain-protocol/database"
	"github.com/workspace/evoting/ev-blockchain-protocol/pkg/config"
)

func getStore() database.Store {
	store, err := database.NewStore("badgerdb", "4000")
	if err != nil {
		logger.Panic(err)
	}
	return store
}

func StartServer(port string) {

	serve := jrpc.New()
	bc := blockchain.NewBlockchain(
		getStore(),
		config.Config{},
	)
	NewHandler(bc, serve)

	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		io.WriteString(res, "RPC SERVER LIVE!")
	})

	http.HandleFunc("/json-rpc", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			logger.Panic(err)
		}
		defer r.Body.Close()

		w.Header().Set("Content-Type", "applicaition/json")
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write(serve.HandleRPCJsonRawMessage(ctx, body)); err != nil {
			logger.Panic(err)
		}
	})

	logger.Infof("Serving rpc on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		logger.Panic(err)
	}
}
