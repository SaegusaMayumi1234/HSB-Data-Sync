package shared

import (
	"encoding/json"

	"github.com/saegusamayumi1234/hsb-data-sync/internal/constant"
	"github.com/saegusamayumi1234/hsb-data-sync/internal/domain/hypixel"
)

var (
	KeyNeuRepoConstantsPets = NewKey[json.RawMessage](constant.SystemKVKeys.NeuRepoConstantsPets)
	KeySkyblockVersion = NewKey[string]("skyblock_version")
	KeySkyblockItemsReferenceLookupMap = NewKey[*hypixel.SkyblockItemReferenceLookupMap]("skyblock_items_lookup_map")
)
