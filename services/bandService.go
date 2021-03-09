package services

import (
	"fmt"
	"github.com/joeyave/scala-chords-bot/entities"
	"github.com/joeyave/scala-chords-bot/repositories"
	"github.com/kjk/notionapi"
	"time"
)

type BandService struct {
	bandRepository *repositories.BandRepository
	notionClient   *notionapi.Client
}

func NewBandService(bandRepository *repositories.BandRepository, notionClient *notionapi.Client) *BandService {
	return &BandService{
		bandRepository: bandRepository,
		notionClient:   notionClient,
	}
}

func (s *BandService) FindAll() ([]*entities.Band, error) {
	return s.bandRepository.FindAll()
}

func (s *BandService) UpdateOne(band entities.Band) (*entities.Band, error) {
	return s.bandRepository.UpdateOne(band)
}

func (s *BandService) GetTodayOrAfterEvents(band entities.Band) ([]*entities.Event, error) {
	q := []byte(`{
        "aggregations": [
            {
                "aggregator": "count",
                "property": "title"
            }
        ],
        "filter": {
            "filters": [
                {
                    "filter": {
                        "operator": "date_is_on_or_after",
                        "value": {
                            "type": "relative",
                            "value": "today"
                        }
                    },
					"property": "\\qpk"
                }
            ],
            "operator": "and"
        },
        "sort": [
            {
                "direction": "descending",
                "property": "\\qpk"
            }
        ]
    }`)

	res, _ := s.notionClient.QueryCollection(band.NotionCollection.NotionCollectionID, band.NotionCollection.NotionCollectionViewID, q, nil)

	var events []*entities.Event

	for _, blockID := range res.Result.BlockIDS {
		block := res.RecordMap.Blocks[blockID].Block

		if block == nil || !block.IsPage() || block.IsSubPage() || block.IsLinkToPage() {
			continue
		}

		event := &entities.Event{
			ID: block.ID,
		}

		eventTitleProp := block.GetTitle()
		if len(eventTitleProp) > 0 {
			event.Name = eventTitleProp[0].Text
		}

		eventTimeProp := block.GetProperty("\\qpk")
		eventSetlistProp := block.GetProperty("wZ?W")

		if len(eventTimeProp) < 1 || len(eventSetlistProp) < 1 {
			continue
		}

		date := notionapi.AttrGetDate(eventTimeProp[0].Attrs[0])

		var err error
		if date.StartTime != "" {
			event.Time, err = time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", date.StartDate, date.StartTime))
		} else {
			event.Time, err = time.Parse("2006-01-02", date.StartDate)
		}
		if err != nil {
			continue
		}

		for _, prop := range eventSetlistProp {
			if len(prop.Attrs) < 1 {
				continue
			}

			pageID := notionapi.AttrGetPageID(prop.Attrs[0])

			event.SetlistPageIDs = append(event.SetlistPageIDs, pageID)
		}

		events = append(events, event)
	}

	return events, nil
}
