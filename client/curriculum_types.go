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
	// Load all the "possible" fields
	var items []struct {
		Class               string   `json:"_class"`
		ID                  int      `json:"id"`
		Title               string   `json:"title"`
		TitleCleaned        string   `json:"title_cleaned"`
		Asset               *Asset   `json:"asset"`
		SupplementaryAssets []*Asset `json:"supplementary_assets"`
		ObjectIndex         int      `json:"object_index"`
	}
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	for idx, i := range items {
		if i.Class == "chapter" {
			// ok it's a chapter
			*c = append(*c, &Chapter{
				ID:          i.ID,
				Title:       i.Title,
				ObjectIndex: i.ObjectIndex,
			})
		} else if i.Class == "lecture" {
			*c = append(*c, &Lecture{
				ID:                  i.ID,
				Title:               i.Title,
				ObjectIndex:         i.ObjectIndex,
				TitleCleaned:        i.TitleCleaned,
				Asset:               i.Asset,
				SupplementaryAssets: i.SupplementaryAssets,
			})
		} else {
			return fmt.Errorf("unknown type for curriculum item at position %d: want \"chapter\" or \"lecture\", got %q", idx, chap.Class)
		}
	}
	return nil
}
