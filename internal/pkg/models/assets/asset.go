package assets


type Asset struct {
	Creator string `json:"creator"`
	Name string `json:"name"`
	UnitName string `json:"unit_name"`
	Decimals uint32 `json:"decimals"`
	Note string `json:"note"`
	TotalSupply uint64 `json:"total_supply"`
	Manager string `json:"manager_address"`
	ReserveAuthAddress string `json:"reserve_address"`
	FreezeAuthAddress string `json:"freeze_address"`
	ClawbackAuthAddress string `json:"clawback_address"`
}