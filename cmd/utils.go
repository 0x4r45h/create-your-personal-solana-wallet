package cmd

import (
	"context"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/client/rpc"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/sysprog"
	"github.com/portto/solana-go-sdk/types"
	"io/ioutil"
	"log"
)

type Wallet struct {
	account types.Account
	c       *client.Client
}

func CreateNewWallet(RPCEndpoint string) Wallet {
	// create a new wallet using types.NewAccount()
	newAccount := types.NewAccount()
	data := []byte(newAccount.PrivateKey)

	err := ioutil.WriteFile("data", data, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return Wallet{
		newAccount,
		client.NewClient(RPCEndpoint),
	}
}
func GetWalletFromFile(filename string) []byte {
	privateKey, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	return privateKey
}
func ImportOldWallet(privateKey []byte, RPCEndpoint string) (Wallet, error) {
	// import a wallet with bytes slice private key
	wallet, err := types.AccountFromBytes(privateKey)
	if err != nil {
		return Wallet{}, err
	}
	return Wallet{
		wallet,
		client.NewClient(RPCEndpoint),
	}, nil
}

func GetBalance() (uint64, error) {
	wallet, _ := ImportOldWallet(GetWalletFromFile("data"), rpc.DevnetRPCEndpoint)
	balance, err := wallet.c.GetBalance(
		context.TODO(),                      // request context
		wallet.account.PublicKey.ToBase58(), // wallet to fetch balance for
	)
	if err != nil {
		return 0, nil
	}
	return balance, nil
}

func RequestAirdrop(amount uint64) (string, error) {
	// request for SOL using RequestAirdrop()
	wallet, _ := ImportOldWallet(GetWalletFromFile("data"), rpc.DevnetRPCEndpoint)
	amount = amount * 1e9 // turning SOL into lamports
	txhash, err := wallet.c.RequestAirdrop(
		context.TODO(),                      // request context
		wallet.account.PublicKey.ToBase58(), // wallet address requesting airdrop
		amount,                              // amount of SOL in lamport
	)
	if err != nil {
		return "", err
	}
	return txhash, nil
}

func Transfer(receiver string, amount uint64) (string, error) {
	// fetch the most recent blockhash
	wallet, _ := ImportOldWallet(GetWalletFromFile("data"), rpc.DevnetRPCEndpoint)
	response, err := wallet.c.GetRecentBlockhash(context.TODO())
	if err != nil {
		return "", err
	}

	// make a transfer message with the latest block hash
	message := types.NewMessage(
		wallet.account.PublicKey, // public key of the transaction signer
		[]types.Instruction{
			sysprog.Transfer(
				wallet.account.PublicKey,             // public key of the transaction sender
				common.PublicKeyFromString(receiver), // wallet address of the transaction receiver
				amount,                               // transaction amount in lamport
			),
		},
		response.Blockhash, // recent block hash
	)

	// create a transaction with the message and TX signer
	tx, err := types.NewTransaction(message, []types.Account{wallet.account, wallet.account})
	if err != nil {
		return "", err
	}

	// send the transaction to the blockchain
	txhash, err := wallet.c.SendTransaction2(context.TODO(), tx)
	if err != nil {
		return "", err
	}
	return txhash, nil
}
