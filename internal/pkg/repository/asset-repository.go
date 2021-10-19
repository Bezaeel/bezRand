package repository

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/crypto"
	models "github.com/bezaeel/algorand/bez-rand/internal/pkg/models/assets"

	transaction "github.com/algorand/go-algorand-sdk/future"
)

type AssetRepository struct {}
var assetRepository *AssetRepository
var algodClient *algod.Client

func GetAssetRepository() *AssetRepository{
	if  assetRepository == nil {
		assetRepository = &AssetRepository{}
	}
	initClient()
	return assetRepository
}

// sandbox
const algodAddress = "http://localhost:4001"
const algodToken = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func initClient(){
	_algodClient, err := algod.MakeClient(algodAddress, algodToken)
	if err != nil{
		return
	}
	algodClient = _algodClient
}

// prettyPrint prints Go structs
func prettyPrint(data interface{}) {
	var p []byte
	//    var err := error
	p, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s \n", p)
}

// printAssetHolding utility to print asset holding for account
func printAssetHolding(assetID uint64, account string, client *algod.Client) {

	act, err := client.AccountInformation(account).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to get account information: %s\n", err)
		return
	}
	for _, assetholding := range act.Assets {
		if assetID == assetholding.AssetId {
			prettyPrint(assetholding)
			break
		}
	}
}

// printCreatedAsset utility to print created assert for account
func printCreatedAsset(assetID uint64, account string, client *algod.Client) {

	act, err := client.AccountInformation(account).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to get account information: %s\n", err)
		return
	}
	for _, asset := range act.CreatedAssets {
		if assetID == asset.Index {
			prettyPrint(asset)
			break
		}
	}
}

func(r *AssetRepository) Create(asset *models.Asset) string {

	

	// Get network-related transaction parameters and assign
	txParams, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		fmt.Printf("Error getting suggested tx params: %s\n", err)
		return ""
	}
	// comment out the next two (2) lines to use suggested fees
	txParams.FlatFee = true
	txParams.Fee = 1000

	// Get pre-defined set of keys for example
	accounts := GetAccountRepository().loadAccounts()
	creator := accounts[1].PublicKey
	assetName := asset.Name
	unitName := asset.UnitName
	assetURL := "https://path/to/my/asset/details"
	assetMetadataHash := "thisIsSomeLength32HashCommitment"
	defaultFrozen := false
	decimals := uint32(0)
	totalIssuance := uint64(1000)
	manager := accounts[0].PublicKey
	reserve := accounts[1].PublicKey
	freeze := accounts[1].PublicKey
	clawback := accounts[1].PublicKey
	note := []byte(nil)
	txn, err := transaction.MakeAssetCreateTxn(creator,
		note,
		txParams, totalIssuance, decimals,
		defaultFrozen, manager, reserve, freeze, clawback,
		unitName, assetName, assetURL, assetMetadataHash)

	if err != nil {
		fmt.Printf("Failed to make asset: %s\n", err)
		return ""
	}
	fmt.Printf("Asset created AssetName: %s\n", txn.AssetConfigTxnFields.AssetParams.AssetName)
	// sign the transaction
	txid, stx, err := crypto.SignTransaction(ed25519.PrivateKey(accounts[0].SecretKey), txn)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %s\n", err)
		return ""
	}
	fmt.Printf("Transaction ID: %s\n", txid)
	// Broadcast the transaction to the network
	sendResponse, err := algodClient.SendRawTransaction(stx).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to send transaction: %s\n", err)
		return ""
	}
	fmt.Printf("Submitted transaction %s\n", sendResponse)
	// Wait for transaction to be confirmed
	GetAccountRepository().waitForConfirmation(txid, algodClient)
	act, err := algodClient.AccountInformation(accounts[0].PublicKey).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to get account information: %s\n", err)
		return ""
	}

	assetID := uint64(0)
	//	find newest (highest) asset for this account
	for _, asset := range act.CreatedAssets {
		if asset.Index > assetID {
			assetID = asset.Index
		}
	}

	// print created asset and asset holding info for this asset
	fmt.Printf("Asset ID: %d\n", assetID)
	printCreatedAsset(assetID, accounts[0].PublicKey, algodClient)
	printAssetHolding(assetID, accounts[0].PublicKey, algodClient)
	return string(assetID)
}

func (a *AssetRepository) MarkAssetInterest(assetCreatorAddress string, assetName string, userMnemonic string) bool{
	// OPT-IN

	// Account 3 opts in to receive latinum
	// Use previously set transaction parameters and update sending address to account 3
	// assetID := uint64(332920)
	// Get network-related transaction parameters and assign
	txParams, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		fmt.Printf("Error getting suggested tx params: %s\n", err)
		return false
	}
	// comment out the next two (2) lines to use suggested fees
	txParams.FlatFee = true
	txParams.Fee = 1000

	//get asset by creator address and asset name
	acct, err := algodClient.AccountInformation(assetCreatorAddress).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to get account information: %s\n", err)
		return false
	}

	assetID := uint64(0)
	//	find newest (highest) asset for this account
	for _, asset := range acct.CreatedAssets {
		if asset.Params.Name == assetName {
			assetID = asset.Index
		}
	}

	user := GetAccountRepository().GetAccountByMnenomics(userMnemonic)

	txn, err := transaction.MakeAssetAcceptanceTxn(user.Address, []byte(nil), txParams, assetID)
	if err != nil {
		fmt.Printf("Failed to send transaction MakeAssetAcceptanceTxn: %s\n", err)
		return false
	}
	txid, stx, err := crypto.SignTransaction(ed25519.PrivateKey(user.SecretKey), txn)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %s\n", err)
		return false
	}

	fmt.Printf("Transaction ID: %s\n", txid)
	// Broadcast the transaction to the network
	sendResponse, err := algodClient.SendRawTransaction(stx).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to send transaction: %s\n", err)
		return false
	}
	fmt.Printf("Transaction ID raw: %s\n", txid)

	// Wait for transaction to be confirmed
	GetAccountRepository().waitForConfirmation(txid, algodClient)

	// print created assetholding for this asset and Account 3, showing 0 balance
	fmt.Printf("Asset ID: %d\n", assetID)
	fmt.Printf("Asset ID: %d\n", sendResponse)
	fmt.Printf("Account 3: %s\n", user.PublicKey)
	printAssetHolding(assetID, user.PublicKey, algodClient)
	return true
}
	// your terminal output should be similar to this...

	// Transaction ID: JYVJEB25YMAVNSAFDTZECWMJTKZHSFJGICGGXF64TH5RTXDICIUA
	// Transaction ID raw: JYVJEB25YMAVNSAFDTZECWMJTKZHSFJGICGGXF64TH5RTXDICIUA
	// waiting for confirmation
	// Transaction JYVJEB25YMAVNSAFDTZECWMJTKZHSFJGICGGXF64TH5RTXDICIUA confirmed in round 4086079
	// Asset ID: 2654040
	// Account 3: 3ZQ3SHCYIKSGK7MTZ7PE7S6EDOFWLKDQ6RYYVMT7OHNQ4UJ774LE52AQCU
	// {
	// 	"amount": 0,
	// 	"asset-id": 2654040,
	// 	"creator": "THQHGD4HEESOPSJJYYF34MWKOI57HXBX4XR63EPBKCWPOJG5KUPDJ7QJCM"
	// } 

	// TRANSFER ASSET
	
	// Send  10 latinum from Account 1 to Account 3
	// assetID := uint64(332920)
	// Get network-related transaction parameters and assign
func (a *AccountRepository) Transfer(senderMnenomics string, receiverAddress string, assetName string, ) (string, bool){
	//load secretKey from Mnemonics
	sender := GetAccountRepository().GetAccountByMnenomics(senderMnenomics)
	txParams, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		fmt.Printf("Error getting suggested tx params: %s\n", err)
		return "", false
	}
	// comment out the next two (2) lines to use suggested fees
	txParams.FlatFee = true
	txParams.Fee = 1000

	sender := accounts[0].PublicKey
	// recipient := accounts[2].PublicKey
	amount := uint64(10)
	closeRemainderTo := ""
	txn, err := transaction.MakeAssetTransferTxn(sender.PublicKey, receiverAddress, amount, []byte(nil), txParams, closeRemainderTo, 
		assetID)
	if err != nil {
		fmt.Printf("Failed to send transaction MakeAssetTransfer Txn: %s\n", err)
		return "", false
	}
	txid, stx, err := crypto.SignTransaction(ed25519.PrivateKey(accounts[0].SecretKey), txn)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %s\n", err)
		return "", false
	}
	fmt.Printf("Transaction ID: %s\n", txid)
	// Broadcast the transaction to the network
	sendResponse, err := algodClient.SendRawTransaction(stx).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to send transaction: %s\n", err)
		return "", false
	}
	fmt.Printf("Transaction ID raw: %s\n", txid)

	// Wait for transaction to be confirmed
	GetAccountRepository().waitForConfirmation(txid,algodClient)

	// print created assetholding for this asset and Account 3 and Account 1
	// You should see amount of 10 in Account 3, and 990 in Account 1
	fmt.Printf("Asset ID: %d\n", assetID)
	fmt.Printf("Account 3: %s\n", receiverAddress)
	printAssetHolding(assetID, receiverAddress, algodClient)
	fmt.Printf("Account 1: %s\n", receiverAddress)
	printAssetHolding(assetID, receiverAddress, algodClient)
	return txid, true
}
	// Your terminal output should look similar to this
	// Transaction ID: 7GPXSVF6YYHHHIGHDCGGR2AS2XXLMDXTUR6GUTSZU4GMIOK2V7TQ
	// Transaction ID raw: 7GPXSVF6YYHHHIGHDCGGR2AS2XXLMDXTUR6GUTSZU4GMIOK2V7TQ
	// waiting for confirmation
	// Transaction 7GPXSVF6YYHHHIGHDCGGR2AS2XXLMDXTUR6GUTSZU4GMIOK2V7TQ confirmed in round 4086081
	// Asset ID: 2654040
	// Account 3: 3ZQ3SHCYIKSGK7MTZ7PE7S6EDOFWLKDQ6RYYVMT7OHNQ4UJ774LE52AQCU
	// {
	// 	"amount": 10,
	// 	"asset-id": 2654040,
	// 	"creator": "THQHGD4HEESOPSJJYYF34MWKOI57HXBX4XR63EPBKCWPOJG5KUPDJ7QJCM"
	// } 
	// Account 1: THQHGD4HEESOPSJJYYF34MWKOI57HXBX4XR63EPBKCWPOJG5KUPDJ7QJCM
	// {
	// 	"amount": 990,
	// 	"asset-id": 2654040,
	// 	"creator": "THQHGD4HEESOPSJJYYF34MWKOI57HXBX4XR63EPBKCWPOJG5KUPDJ7QJCM"
	// } 

	// FREEZE ASSET
	// The freeze address (Account 2) Freeze's asset for Account 3.
	// assetID := uint64(332920)
	// Get network-related transaction parameters and assign
func(a *AccountRepository) FreezeAddress(defaulterAddress string, authorizerAddress string) bool{
	txParams, err = algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		fmt.Printf("Error getting suggested tx params: %s\n", err)
		return
	}
	// comment out the next two (2) lines to use suggested fees
	txParams.FlatFee = true
	txParams.Fee = 1000
	newFreezeSetting := true
	target := accounts[2].PublicKey
	txn, err = transaction.MakeAssetFreezeTxn(authorizerAddress, note, txParams, assetID, defaulterAddress, newFreezeSetting)
	if err != nil {
		fmt.Printf("Failed to send txn: %s\n", err)
		return
	}
	txid, stx, err = crypto.SignTransaction(ed25519.PrivateKey(accounts[1].SecretKey), txn)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %s\n", err)
		return
	}
	fmt.Printf("Transaction ID: %s\n", txid)
	// Broadcast the transaction to the network
	sendResponse, err = algodClient.SendRawTransaction(stx).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to send transaction: %s\n", err)
		return
	}
	fmt.Printf("Transaction ID raw: %s\n", txid)
	// Wait for transaction to be confirmed
	GetAccountRepository().waitForConfirmation(txid,algodClient)
    // You should now see is-frozen value of true
	fmt.Printf("Asset ID: %d\n", assetID)
	fmt.Printf("Account 3: %s\n", accounts[2].PublicKey)
	printAssetHolding(assetID, accounts[2].PublicKey, algodClient)
	return true
}
	

	// Your terminal output should look similar to this:

	// Transaction ID: FHFLUVKQ5Q4S2RRLOA6EJ6NVQDZEVU6TDKNOVJK5ZNKCDYUZFNXQ
	// Transaction ID raw: FHFLUVKQ5Q4S2RRLOA6EJ6NVQDZEVU6TDKNOVJK5ZNKCDYUZFNXQ
	// waiting for confirmation
	// Transaction FHFLUVKQ5Q4S2RRLOA6EJ6NVQDZEVU6TDKNOVJK5ZNKCDYUZFNXQ confirmed in round 4086084
	// Asset ID: 2654040
	// Account 3: 3ZQ3SHCYIKSGK7MTZ7PE7S6EDOFWLKDQ6RYYVMT7OHNQ4UJ774LE52AQCU
	// {
	// 	"amount": 10,
	// 	"asset-id": 2654040,
	// 	"creator": "THQHGD4HEESOPSJJYYF34MWKOI57HXBX4XR63EPBKCWPOJG5KUPDJ7QJCM",
	// 	"is-frozen": true
	// }
	
	// REVOKE ASSET
	// Revoke an Asset
	// The clawback address (Account 2) revokes 10 latinum from Account 3 (target)
	// and places it back with Account 1 (creator).
	// assetID := uint64(332920)
	// Get network-related transaction parameters and assign
func revoke()	{
	txParams, err = algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		fmt.Printf("Error getting suggested tx params: %s\n", err)
		return
	}
	// comment out the next two (2) lines to use suggested fees
	txParams.FlatFee = true
	txParams.Fee = 1000
	target = accounts[3].PublicKey
	txn, err = transaction.MakeAssetRevocationTxn(clawback, target, amount, creator, note,
		txParams, assetID)
	if err != nil {
		fmt.Printf("Failed to send txn: %s\n", err)
		return
	}
	txid, stx, err = crypto.SignTransaction(ed25519.PrivateKey(accounts[1].SecretKey), txn)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %s\n", err)
		return
	}
	fmt.Printf("Transaction ID: %s\n", txid)
	// Broadcast the transaction to the network
	sendResponse, err = algodClient.SendRawTransaction(stx).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to send transaction: %s\n", err)
		return
	}
	fmt.Printf("Transaction ID raw: %s\n", txid)
	// Wait for transaction to be confirmed
	GetAccountRepository().waitForConfirmation( txid, algodClient)
	// print created assetholding for this asset and Account 3 and Account 1
	// You should see amount of 0 in Account 3, and 1000 in Account 1
	fmt.Printf("Asset ID: %d\n", assetID)
	fmt.Printf("recipient")
	fmt.Printf("Account 3: %s\n", accounts[2].PublicKey)
	printAssetHolding(assetID, accounts[2].PublicKey, algodClient)
	fmt.Printf("target")
	fmt.Printf("Account 1: %s\n", accounts[0].PublicKey)
	printAssetHolding(assetID, accounts[0].PublicKey, algodClient)
}

	// Your terminal output should look similar to this...

	// Transaction XH32YUIX2VTEH3QPJECVNVXHVHU2LBQVGIHMPPSRE4XGLFNUG63Q confirmed in round 4086090
	// Asset ID: 2654040
	// recipientAccount 3: 3ZQ3SHCYIKSGK7MTZ7PE7S6EDOFWLKDQ6RYYVMT7OHNQ4UJ774LE52AQCU
	// {
	// 	"amount": 0,
	// 	"asset-id": 2654040,
	// 	"creator": "THQHGD4HEESOPSJJYYF34MWKOI57HXBX4XR63EPBKCWPOJG5KUPDJ7QJCM",
	// 	"is-frozen": true
	// } 
	// targetAccount 1: THQHGD4HEESOPSJJYYF34MWKOI57HXBX4XR63EPBKCWPOJG5KUPDJ7QJCM
	// {
	// 	"amount": 1000,
	// 	"asset-id": 2654040,
	// 	"creator": "THQHGD4HEESOPSJJYYF34MWKOI57HXBX4XR63EPBKCWPOJG5KUPDJ7QJCM"
	// }

	// DESTROY ASSET
	// Destroy the asset
	// Make sure all funds are back in the creator's account. Then use the
	// Manager account (Account 1) to destroy the asset.

	// assetID := uint64(332920)
	// Get network-related transaction parameters and assign
func destroy()	{
	txParams, err = algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		fmt.Printf("Error getting suggested tx params: %s\n", err)
		return
	}
	// comment out the next two (2) lines to use suggested fees
	txParams.FlatFee = true
	txParams.Fee = 1000

	txn, err = transaction.MakeAssetDestroyTxn(manager, note, txParams, assetID)
	if err != nil {
		fmt.Printf("Failed to send txn: %s\n", err)
		return
	}
	txid, stx, err = crypto.SignTransaction(ed25519.PrivateKey(accounts[0].SecretKey), txn)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %s\n", err)
		return
	}
	fmt.Printf("Transaction ID: %s\n", txid)
	// Broadcast the transaction to the network
	sendResponse, err = algodClient.SendRawTransaction(stx).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to send transaction: %s\n", err)
		return
	}
	fmt.Printf("Transaction ID raw: %s\n", txid)
	// Wait for transaction to be confirmed
	GetAccountRepository().waitForConfirmation(txid,algodClient)
	fmt.Printf("Asset ID: %d\n", assetID)	
	fmt.Printf("Account 3 must do a transaction for an amount of 0, \n" )
    fmt.Printf("with a closeRemainderTo to the creator account, to clear it from its accountholdings. \n")
    fmt.Printf("For Account 1, nothing should print after this as the asset is destroyed on the creator account \n")

	// print created asset and asset holding info for this asset (should not print anything)

	printCreatedAsset(assetID, accounts[0].PublicKey, algodClient)
	printAssetHolding(assetID, accounts[0].PublicKey, algodClient)

	// Your terminal output should look similar to this...

	// Transaction PI4U7DJZYDKEZS2PKTNGB6DFNVCCEYN5FNLZBBWNONTWMA7RH6AA confirmed in round 4086093
	// Asset ID: 2654040
	// Account 3 must do a transaction for an amount of 0, 
	// with a closeRemainderTo to the creator account, to clear it from its accountholdings.
	// For Account 1, nothing should print after this as the asset is destroyed on the creator account
}
