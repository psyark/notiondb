package notiondb

import (
	"context"

	"github.com/jomei/notionapi"
)

// ClearBlockChildren はブロックまたはページの子孫ブロックを全て消します
func ClearBlockChildren[pid notionapi.PageID | notionapi.BlockID](ctx context.Context, client *notionapi.Client, parentID pid) error {
	resp, err := client.Block.GetChildren(ctx, notionapi.BlockID(parentID), nil)
	if err != nil {
		return err
	}
	for _, b := range resp.Results {
		if _, err := client.Block.Delete(ctx, b.GetID()); err != nil {
			return err
		}
	}
	return nil
}
