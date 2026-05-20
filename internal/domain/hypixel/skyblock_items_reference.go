package hypixel

import (
	"encoding/json"
	"strings"
)

type SkyblockItemReferenceResponse struct {
	Success     bool                             `json:"success"`
	LastUpdated int64                            `json:"lastUpdated"`
	Items       []json.RawMessage `json:"items"`
}

type SkyblockItemReference struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	// RawMessage preserves unknown fields without a strict schema
	Raw json.RawMessage `json:"-"`
}

type SkyblockItemReferenceLookupMap struct {
    ById   map[string]*SkyblockItemReference
    ByName map[string]*SkyblockItemReference
}

func BuildSkyblockItemsReferenceLookupMap(data json.RawMessage) (*SkyblockItemReferenceLookupMap, error) {
	var parsedData SkyblockItemReferenceResponse
	err := json.Unmarshal(data, &parsedData)
	if err != nil {
		return nil, err
	}

	lookup := &SkyblockItemReferenceLookupMap{
		ById:   make(map[string]*SkyblockItemReference, len(parsedData.Items)),
		ByName: make(map[string]*SkyblockItemReference, len(parsedData.Items)),
	}

	for _, raw := range parsedData.Items {
        var item SkyblockItemReference
        if err := json.Unmarshal(raw, &item); err != nil {
            continue
        }
        item.Raw = raw

        lookup.ById[item.Id] = &item
        lookup.ByName[strings.ToLower(item.Name)] = &item
    }

	return lookup, nil
}

func (lm *SkyblockItemReferenceLookupMap) FindByID(query string) *SkyblockItemReference {
    if item, ok := lm.ById[query]; ok {
        return item
    }
    return nil
}

func (lm *SkyblockItemReferenceLookupMap) FindByName(query string) *SkyblockItemReference {
	query = strings.ToLower(query)
	if item, ok := lm.ByName[query]; ok {
		return item
	}
	return nil
}
