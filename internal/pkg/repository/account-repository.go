package repository

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/mnemonic"
	"github.com/algorand/go-algorand-sdk/types"
	"github.com/bezaeel/algorand/bez-rand/internal/pkg/models/accounts"
	models "github.com/bezaeel/algorand/bez-rand/internal/pkg/models/accounts"
)

type AccountRepository struct {}
var accountRepository *AccountRepository

func GetAccountRepository() *AccountRepository{
	if  accountRepository == nil {
		accountRepository = &AccountRepository{}
	}
	return accountRepository
}


func (a *AccountRepository) loadAccounts() []*accounts.Account {
	var accts []*accounts.Account
	mnemonic1 := "portion never forward pill lunch organ biology weird catch curve isolate plug innocent skin grunt bounce clown mercy hole eagle soul chunk type absorb trim"
	mnemonic2 := "place blouse sad pigeon wing warrior wild script problem team blouse camp soldier breeze twist mother vanish public glass code arrow execute convince ability there"
	mnemonic3 := "image travel claw climb bottom spot path roast century also task cherry address curious save item clean theme amateur loyal apart hybrid steak about blanket"

	mnemonics := []string{mnemonic1, mnemonic2, mnemonic3}
	pks := map[int]string{1: "", 2: "", 3: ""}
	var sks = make(map[int][]byte)

	for i, m := range mnemonics {
		var err error
		sk, err := mnemonic.ToPrivateKey(m)
		sks[i+1] = sk
		if err != nil {
			fmt.Printf("Issue with account %d private key conversion.", i+1)
		}
		// derive public address from Secret Key.
		pk := sk.Public()
		var a types.Address
		cpk := pk.(ed25519.PublicKey)
		copy(a[:], cpk[:])
		pks[i+1] = a.String()
		fmt.Printf("Loaded Key %d: %s\n", i+1, pks[i+1])
		var acct = &models.Account{
			Address: string(pks[i+1]),
			SecretKey: string(sks[i+1]),
			PublicKey: string(pks[i+1]),
		}
		accts = append(accts, acct)
	}
	
	return accts
}

// Accounts to be used through examples
func (a *AccountRepository) GetAccountByMnenomics(pass string) *accounts.Account {
	var err error
	sk, err := mnemonic.ToPrivateKey(pass)
	if err != nil {
		fmt.Printf("Issue with account %s private key conversion.", pass)
	}
	_pk := sk.Public()
	var addr types.Address
	cpk := _pk.(ed25519.PublicKey)
	copy(addr[:], cpk[:])
	pk := addr.String()

	return &models.Account{
		Address: pk,
		SecretKey: string(sk),
		PublicKey: pk,
	}
}


func (a *AccountRepository) waitForConfirmation(txID string, client *algod.Client) {
	status, err := client.Status().Do(context.Background())
	if err != nil {
		fmt.Printf("error getting algod status: %s\n", err)
		return
	}
	lastRound := status.LastRound
	for {
		pt, _, err := client.PendingTransactionInformation(txID).Do(context.Background())
		if err != nil {
			fmt.Printf("error getting pending transaction: %s\n", err)
			return
		}
		if pt.ConfirmedRound > 0 {
			fmt.Printf("Transaction "+txID+" confirmed in round %d\n", pt.ConfirmedRound)
			break
		}
		fmt.Printf("waiting for confirmation\n")
		lastRound++
		status, err = client.StatusAfterBlock(lastRound).Do(context.Background())
	}
}


// func (a *AccountRepository) CreateAccount() *models.Account{
// 	account := crypto.GenerateAccount()
// 	return &models.Account{
// 		account.Address.String(),
// 		string(account.PrivateKey),
// 		string(account.PublicKey),
// 	}
// }

// func (a *AccountRepository) GetAccount() *models.Account{
// 	mnemonic1 := "portion never forward pill lunch organ biology weird catch curve isolate plug innocent skin grunt bounce clown mercy hole eagle soul chunk type absorb trim"
	
// 	mnemonics := []string{mnemonic1}
// 	pks := map[int]string{1: ""}
// 	var sks = make(map[int][]byte)
// 	var account models.Account

// 	for i, m := range mnemonics {
// 		var err error
// 		sk, err := mnemonic.ToPrivateKey(m)
// 		fmt.Printf("Sk ::%v\n", sk)
// 		sks[i+1] = sk
// 		if err != nil {
// 			fmt.Printf("Issue with account %d private key conversion.", i+1)
// 		}
// 		// derive public address from Secret Key.
// 		pk := sk.Public()
// 		var a types.Address
// 		cpk := pk.(ed25519.PublicKey)
// 		copy(a[:], cpk[:])
// 		pks[i+1] = a.String()
		
// 		fmt.Printf("sks:: %v\n",sks[1])
// 		account.SecretKey = string(sks[1])
// 		account.PublicKey = pks[1]
// 		account.Address = account.PublicKey
// 		fmt.Printf("Loaded Key %d: %s\n", i+1, pks[i+1])
// 	}
// 	return &account
// }