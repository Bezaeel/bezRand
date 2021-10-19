package main

import (
	"fmt"

	"github.com/bezaeel/algorand/bez-rand/internal/pkg/models/assets"
	"github.com/bezaeel/algorand/bez-rand/internal/pkg/repository"
)

// Main function to demonstrate ASA examples
func main() {
	var asset repository.AssetRepository

	asset.Create(&assets.Asset{
		Creator: "Talabi",
		Name: "Talabi",
		UnitName: "TAL",
		Note: "Talabi",
		Decimals: 2,
		TotalSupply: 10000,
	})

	fmt.Scanln()
}







// package main

// import (
// 	"fmt"
// 	"github.com/bezaeel/algorand/bez-rand/internal/pkg/models/assets"
// 	"github.com/bezaeel/algorand/bez-rand/internal/pkg/repository"
// )

// func main(){
// 	// var asset repository.AssetRepository

// 	// asset.Create(&assets.Asset{
// 	// 	Creator: "Talabi",
// 	// 	Name: "Talabi",
// 	// 	UnitName: "TAL",
// 	// 	Note: "Talabi",
// 	// 	Decimals: 2,
// 	// 	TotalSupply: 10000,

// 	// });

// 	// fmt.Scanln()
// }