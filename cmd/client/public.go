/**
*  @file
*  @copyright defined in go-seele/LICENSE
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/seeleteam/go-seele/cmd/util"
	"github.com/seeleteam/go-seele/common"
	"github.com/seeleteam/go-seele/common/keystore"
	"github.com/seeleteam/go-seele/core/types"
	"github.com/seeleteam/go-seele/rpc2"
	"github.com/seeleteam/go-seele/seele"
	"github.com/urfave/cli"
)

func RPCAction(handler func(client *rpc.Client) (interface{}, error)) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		client, err := rpc.DialTCP(context.Background(), addressValue)
		if err != nil {
			return err
		}

		result, err := handler(client)
		if err != nil {
			return fmt.Errorf("get error when call rpc %s", err)
		}

		if result != nil {
			resultStr, err := json.MarshalIndent(result, "", "\t")
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", resultStr)
		}

		return nil
	}
}

func GetInfoAction(client *rpc.Client) (interface{}, error) {
	return util.GetInfo(client)
}

func getBalanceAction(client *rpc.Client) (interface{}, error) {
	account, err := MakeAddress(accountValue)
	if err != nil {
		return nil, err
	}

	var result seele.GetBalanceResponse
	err = client.Call(&result, "seele_getBalance", account)
	return result, err
}

func GetAccountNonceAction(client *rpc.Client) (interface{}, error) {
	account, err := MakeAddress(accountValue)
	if err != nil {
		return nil, err
	}

	return util.GetAccountNonce(client, account)
}

func GetBlockHeightAction(client *rpc.Client) (interface{}, error) {
	var result uint64
	err := client.Call(&result, "seele_getBlockHeight")
	return result, err
}

func GetBlockAction(client *rpc.Client) (interface{}, error) {
	var result map[string]interface{}
	var err error

	if hashValue != "" {
		err = client.Call(&result, "seele_getBlockByHash", hashValue, fulltxValue)
	} else {
		err = client.Call(&result, "seele_getBlockByHeight", heightValue, fulltxValue)
	}

	return result, err
}

func GetLogsAction(client *rpc.Client) (interface{}, error) {
	var result []seele.GetLogsResponse
	err := client.Call(&result, "seele_getLogs", heightValue, contractValue, topicValue)

	return result, err
}

func callAction(client *rpc.Client) (interface{}, error) {
	result := make(map[string]interface{})
	err := client.Call(&result, "seele_call", toValue, paloadValue, heightValue)

	return result, err
}

func AddTxAction(client *rpc.Client) (interface{}, error) {
	tx, err := MakeTransaction(client)
	if err != nil {
		return nil, err
	}

	var result bool
	if err = client.Call(&result, "seele_addTx", *tx); err != nil || !result {
		fmt.Println("failed to send transaction")
		return nil, err
	}

	fmt.Println("transaction sent successfully")
	return tx, nil
}

func MakeAddress(value string) (common.Address, error) {
	if value == "" {
		return common.EmptyAddress, nil
	} else {
		return common.HexToAddress(value)
	}
}

func MakeTransaction(client *rpc.Client) (*types.Transaction, error) {
	pass, err := common.GetPassword()
	if err != nil {
		return nil, fmt.Errorf("failed to get password %s\n", err)
	}

	key, err := keystore.GetKey(fromValue, pass)
	if err != nil {
		return nil, fmt.Errorf("invalid sender key file. it should be a private key: %s\n", err)
	}

	txd, err := checkParameter(&key.PrivateKey.PublicKey, client)
	if err != nil {
		return nil, err
	}

	return util.GenerateTx(key.PrivateKey, txd.To, txd.Amount, txd.Fee, txd.AccountNonce, txd.Payload)
}
