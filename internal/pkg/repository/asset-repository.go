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

func GetAssetRepository(_algodClient *algod.Client) *AssetRepository{
	if  assetRepository == nil {
		assetRepository = &AssetRepository{}
	}
	algodClient = _algodClient
	return assetRepository
}

// sandbox


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

func getAssetID(account string, assetName string) (uint64, error){ 
	act, err := algodClient.AccountInformation(account).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to get account information: %s\n", err)
		return 0,err
	}

	assetID := uint64(0)
	//	find newest (highest) asset for this account
	for _, asset := range act.CreatedAssets {
		if asset.Index > assetID {
			assetID = asset.Index
		}
	}
	return assetID, nil
}


func (r *AssetRepository) Simulate() {
	fmt.Println("We'll need min of 2 accounts in order to simulate a real life event")
	fmt.Println("For this simulation we'll use 3")
	fmt.Println("1 for organization admin, 1 for compliance, 1 for other user")

	accounts := GetAccountRepository().loadAccounts();
	asset := &models.Asset{
		Creator: "Talabi",
		Name: "Talabiii",
		UnitName: "TAL2",
		Note: "Talabi",
		Decimals: 2,
		TotalSupply: 10000,
		Manager: accounts[0].PublicKey,
		ReserveAuthAddress: accounts[1].PublicKey,
		FreezeAuthAddress: accounts[1].PublicKey,
		ClawbackAuthAddress: accounts[1].PublicKey,
	}

	assetResponse, err := assetRepository.Create(asset)
	if err != nil {
		fmt.Printf("Error occurred %v\n", err)
	}
	_  = assetResponse
	fmt.Println("user needs to opt-in to asset before making txns with asset")
	userMnemonic := "image travel claw climb bottom spot path roast century also task cherry address curious save item clean theme amateur loyal apart hybrid steak about blanket"
	interest := assetRepository.MarkAssetInterest(accounts[0].Address, "Talabiii", userMnemonic)
	if interest == false {
		fmt.Println("Unable to opt-in to asset")
	}
	fmt.Println("user being able to trade asset: being able to receive asset")
	senderMnemonic := "portion never forward pill lunch organ biology weird catch curve isolate plug innocent skin grunt bounce clown mercy hole eagle soul chunk type absorb trim"
	txnResponse, txnSuccess := assetRepository.Transfer(senderMnemonic, accounts[2].Address, "Talabiii")
	if txnSuccess == false {
		fmt.Println("Please confirm user opt-in to asset")
	}
	_ = txnResponse
	fmt.Println("a fraud complaint has been raised with respect to user")
	fmt.Println("Compliance within organization decides to freeze user from transacting asset")

	authorizerMnemonic := "place blouse sad pigeon wing warrior wild script problem team blouse camp soldier breeze twist mother vanish public glass code arrow execute convince ability there"
	isFrozen := assetRepository.FreezeAddress(accounts[2].Address, authorizerMnemonic, "Talabiii");
	if isFrozen == false {
		fmt.Println("unable to freeze user at this time")
	}
	fmt.Printf("user with address: %v cannot trade assets for now", accounts[2].Address)
}



func(a *AssetRepository) Create(asset *models.Asset) (string, error) {
	// Get network-related transaction parameters and assign
	txParams, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		fmt.Printf("Error getting suggested tx params: %s\n", err)
		return "", err
	}
	// comment out the next two (2) lines to use suggested fees
	txParams.FlatFee = true
	txParams.Fee = 1000

	// Get pre-defined set of keys for example
	accounts := GetAccountRepository().loadAccounts()
	creator := accounts[0].PublicKey
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
		return "", err
	}
	fmt.Printf("Asset created AssetName: %s\n", txn.AssetConfigTxnFields.AssetParams.AssetName)
	// sign the transaction
	txid, stx, err := crypto.SignTransaction(ed25519.PrivateKey(accounts[0].SecretKey), txn)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %s\n", err)
		return "", err
	}
	fmt.Printf("Transaction ID: %s\n", txid)
	// Broadcast the transaction to the network
	sendResponse, err := algodClient.SendRawTransaction(stx).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to send transaction: %s\n", err)
		return "", err
	}
	fmt.Printf("Submitted transaction %s\n", sendResponse)
	// Wait for transaction to be confirmed
	GetAccountRepository().waitForConfirmation(txid, algodClient)
	act, err := algodClient.AccountInformation(accounts[0].PublicKey).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to get account information: %s\n", err)
		return "", err
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
	return string(rune(assetID)), nil
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
	_ = sendResponse
	if err != nil {
		fmt.Printf("failed to send transaction: %s\n", err)
		return false
	}
	fmt.Printf("Transaction ID raw: %s\n", txid)

	// Wait for transaction to be confirmed
	GetAccountRepository().waitForConfirmation(txid, algodClient)

	// print created assetholding for this asset and Account 3, showing 0 balance
	fmt.Printf("Asset ID: %d\n", assetID)
	fmt.Printf("Account 3: %s\n", user.PublicKey)
	printAssetHolding(assetID, user.PublicKey, algodClient)
	return true
}

func (a *AssetRepository) Transfer(senderMnenomics string, receiverAddress string, assetName string, ) (string, bool){
	// load secretKey from Mnemonics
	sender := GetAccountRepository().GetAccountByMnenomics(senderMnenomics)
	assetID, err := getAssetID(sender.PublicKey, assetName)
	if err != nil {
		return "invalid asset", false
	}
	txParams, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		fmt.Printf("Error getting suggested tx params: %s\n", err)
		return "", false
	}
	// comment out the next two (2) lines to use suggested fees
	txParams.FlatFee = true
	txParams.Fee = 1000

	// sender := accounts[0].PublicKey
	// recipient := accounts[2].PublicKey
	amount := uint64(10)
	closeRemainderTo := ""
	txn, err := transaction.MakeAssetTransferTxn(sender.PublicKey, receiverAddress, amount, []byte(nil), txParams, closeRemainderTo, 
		assetID)
	if err != nil {
		fmt.Printf("Failed to send transaction MakeAssetTransfer Txn: %s\n", err)
		return "", false
	}
	txid, stx, err := crypto.SignTransaction(ed25519.PrivateKey(sender.SecretKey), txn)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %s\n", err)
		return "", false
	}
	fmt.Printf("Transaction ID: %s\n", txid)
	// Broadcast the transaction to the network
	sendResponse, err := algodClient.SendRawTransaction(stx).Do(context.Background())
	_ = sendResponse
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
// Get network-related transaction parameters and assign
func(a *AssetRepository) FreezeAddress(defaulterAddress string, authorizerMnemonics string, assetName string) bool{
	authorizer := GetAccountRepository().GetAccountByMnenomics(authorizerMnemonics);
	assetID, err := getAssetID(authorizer.PublicKey, assetName)
	if err != nil {
		return false
	}
	txParams, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		fmt.Printf("Error getting suggested tx params: %s\n", err)
		return false
	}
	// comment out the next two (2) lines to use suggested fees
	txParams.FlatFee = true
	txParams.Fee = 1000
	newFreezeSetting := true
	txn, err := transaction.MakeAssetFreezeTxn(authorizer.SecretKey, []byte(nil), txParams, assetID, defaulterAddress, newFreezeSetting)
	if err != nil {
		fmt.Printf("Failed to send txn: %s\n", err)
		return false
	}
	txid, stx, err := crypto.SignTransaction(ed25519.PrivateKey(authorizer.PublicKey), txn)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %s\n", err)
		return false
	}
	fmt.Printf("Transaction ID: %s\n", txid)
	// Broadcast the transaction to the network
	sendResponse, err := algodClient.SendRawTransaction(stx).Do(context.Background())
	_ = sendResponse
	if err != nil {
		fmt.Printf("failed to send transaction: %s\n", err)
		return false
	}
	fmt.Printf("Transaction ID raw: %s\n", txid)
	// Wait for transaction to be confirmed
	GetAccountRepository().waitForConfirmation(txid,algodClient)
    // You should now see is-frozen value of true
	fmt.Printf("Asset ID: %d\n", assetID)
	fmt.Printf("Account 3: %s\n", defaulterAddress)
	printAssetHolding(assetID, defaulterAddress, algodClient)
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
// func revoke()	{
// 	txParams, err = algodClient.SuggestedParams().Do(context.Background())
// 	if err != nil {
// 		fmt.Printf("Error getting suggested tx params: %s\n", err)
// 		return
// 	}
// 	// comment out the next two (2) lines to use suggested fees
// 	txParams.FlatFee = true
// 	txParams.Fee = 1000
// 	target = accounts[3].PublicKey
// 	txn, err = transaction.MakeAssetRevocationTxn(clawback, target, amount, creator, note,
// 		txParams, assetID)
// 	if err != nil {
// 		fmt.Printf("Failed to send txn: %s\n", err)
// 		return
// 	}
// 	txid, stx, err = crypto.SignTransaction(ed25519.PrivateKey(accounts[1].SecretKey), txn)
// 	if err != nil {
// 		fmt.Printf("Failed to sign transaction: %s\n", err)
// 		return
// 	}
// 	fmt.Printf("Transaction ID: %s\n", txid)
// 	// Broadcast the transaction to the network
// 	sendResponse, err = algodClient.SendRawTransaction(stx).Do(context.Background())
// 	if err != nil {
// 		fmt.Printf("failed to send transaction: %s\n", err)
// 		return
// 	}
// 	fmt.Printf("Transaction ID raw: %s\n", txid)
// 	// Wait for transaction to be confirmed
// 	GetAccountRepository().waitForConfirmation( txid, algodClient)
// 	// print created assetholding for this asset and Account 3 and Account 1
// 	// You should see amount of 0 in Account 3, and 1000 in Account 1
// 	fmt.Printf("Asset ID: %d\n", assetID)
// 	fmt.Printf("recipient")
// 	fmt.Printf("Account 3: %s\n", accounts[2].PublicKey)
// 	printAssetHolding(assetID, accounts[2].PublicKey, algodClient)
// 	fmt.Printf("target")
// 	fmt.Printf("Account 1: %s\n", accounts[0].PublicKey)
// 	printAssetHolding(assetID, accounts[0].PublicKey, algodClient)
// }

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
// func destroy()	{
// 	txParams, err = algodClient.SuggestedParams().Do(context.Background())
// 	if err != nil {
// 		fmt.Printf("Error getting suggested tx params: %s\n", err)
// 		return
// 	}
// 	// comment out the next two (2) lines to use suggested fees
// 	txParams.FlatFee = true
// 	txParams.Fee = 1000

// 	txn, err = transaction.MakeAssetDestroyTxn(manager, note, txParams, assetID)
// 	if err != nil {
// 		fmt.Printf("Failed to send txn: %s\n", err)
// 		return
// 	}
// 	txid, stx, err = crypto.SignTransaction(ed25519.PrivateKey(accounts[0].SecretKey), txn)
// 	if err != nil {
// 		fmt.Printf("Failed to sign transaction: %s\n", err)
// 		return
// 	}
// 	fmt.Printf("Transaction ID: %s\n", txid)
// 	// Broadcast the transaction to the network
// 	sendResponse, err = algodClient.SendRawTransaction(stx).Do(context.Background())
// 	if err != nil {
// 		fmt.Printf("failed to send transaction: %s\n", err)
// 		return
// 	}
// 	fmt.Printf("Transaction ID raw: %s\n", txid)
// 	// Wait for transaction to be confirmed
// 	GetAccountRepository().waitForConfirmation(txid,algodClient)
// 	fmt.Printf("Asset ID: %d\n", assetID)	
// 	fmt.Printf("Account 3 must do a transaction for an amount of 0, \n" )
//     fmt.Printf("with a closeRemainderTo to the creator account, to clear it from its accountholdings. \n")
//     fmt.Printf("For Account 1, nothing should print after this as the asset is destroyed on the creator account \n")

// 	// print created asset and asset holding info for this asset (should not print anything)

// 	printCreatedAsset(assetID, accounts[0].PublicKey, algodClient)
// 	printAssetHolding(assetID, accounts[0].PublicKey, algodClient)

// 	// Your terminal output should look similar to this...

// 	// Transaction PI4U7DJZYDKEZS2PKTNGB6DFNVCCEYN5FNLZBBWNONTWMA7RH6AA confirmed in round 4086093
// 	// Asset ID: 2654040
// 	// Account 3 must do a transaction for an amount of 0, 
// 	// with a closeRemainderTo to the creator account, to clear it from its accountholdings.
// 	// For Account 1, nothing should print after this as the asset is destroyed on the creator account
// }
