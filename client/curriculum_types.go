package client

import (
	"encoding/json"
	"fmt"
)

type Curriculum struct {
	Count    int             `json:"count"`
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Results  CurriculumItems `json:"results"`
}

// CurriculumItem contains either *Chapter or *Lecture items
type CurriculumItems []interface{}

func (c *CurriculumItems) UnmarshalJSON(data []byte) error {
	// parse the "array", but delay parsing of internal items
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	for idx, i := range items {
		// chapter have all "common" fields, so we parse each item as a chapter
		var chap *Chapter
		err := json.Unmarshal(i, &chap)
		if err != nil {
			return err
		}
		if chap.Class == "chapter" {
			// ok it's a chapter
			*c = append(*c, chap)
		} else if chap.Class == "lecture" {
			// lecture: we need to parse the "other" fields
			var lf struct{
				TitleCleaned        string   `json:"title_cleaned"`
				Asset               *Asset   `json:"asset"`
				SupplementaryAssets []*Asset `json:"supplementary_assets"`
			}
			if err := json.Unmarshal(i, &lf); err != nil {
				return err
			}
			*c = append(*c, &Lecture{
				Class:               "lecture",
				ID:                  chap.ID,
				Title:               chap.Title,
				ObjectIndex:         chap.ObjectIndex,
				TitleCleaned:        lf.TitleCleaned,
				Asset:               lf.Asset,
				SupplementaryAssets: lf.SupplementaryAssets,
			})
		} else {
			return fmt.Errorf("unknown type for curriculum item at position %d: want \"chapter\" or \"lecture\", got %q", idx, chap.Class)
		}
	}
	return nil
}
