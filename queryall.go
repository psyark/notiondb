package notiondb

import (
	"context"

	"github.com/jomei/notionapi"
)

func QueryAll(ctx context.Context, client *notionapi.Client, databaseID notionapi.DatabaseID, filter notionapi.Filter) ([]notionapi.Page, error) {
	pages := []notionapi.Page{}
	dqr := &notionapi.DatabaseQueryRequest{PageSize: 100, Filter: filter}
	for {
		resp, err := client.Database.Query(ctx, databaseID, dqr)
		if err != nil {
			return nil, err
		}

		pages = append(pages, resp.Results...)
		if resp.HasMore {
			dqr.StartCursor = resp.NextCursor
		} else {
			break
		}
	}
	return pages, nil
}
