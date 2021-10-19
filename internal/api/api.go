package api

import (
	"fmt"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/bezaeel/algorand/bez-rand/internal/pkg/config"
	"github.com/bezaeel/algorand/bez-rand/internal/pkg/repository"
)

func setConfiguration(configPath string){
	config.Setup(configPath)

}

func Run(configPath string){
	if configPath == ""{
		configPath = "../data/"
	}
	setConfiguration(configPath)
	conf := config.GetConfig()

	fmt.Println("Simulation started ")
	fmt.Println("==================>")

	algodAddress := conf.Algod.Address
	algodToken := conf.Algod.Token
	_algodClient, err := algod.MakeClient(algodAddress, algodToken)
	if err != nil {
		fmt.Printf("Unable to connect to Algod \n %s\n", err)
	}
	asset := repository.GetAssetRepository(_algodClient)
	
	asset.Simulate()
	
	
}