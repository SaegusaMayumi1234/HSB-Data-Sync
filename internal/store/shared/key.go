package shared

import (
	"encoding/json"

	"github.com/saegusamayumi1234/hsb-data-sync/internal/domain/hypixel"
	"github.com/saegusamayumi1234/hsb-data-sync/internal/domain/systemkv"
)

var (
	KeyNeuRepoConstantsPets = NewKey[json.RawMessage](systemkv.Keys.NeuRepoConstantsPets)
	KeySkyblockVersion = NewKey[string]("skyblock_version")
	KeySkyblockItemsReferenceLookupMap = NewKey[*hypixel.SkyblockItemReferenceLookupMap]("skyblock_items_lookup_map")
)
