package notiondb

import (
	"fmt"

	"github.com/jomei/notionapi"
)

func GetPageTitleWithEmoji(page *notionapi.Page, placeholder string) string {
	title := RichTextToString(GetProperty(page.Properties, "title").(*notionapi.TitleProperty).Title)
	if title == "" {
		title = placeholder
	}

	if page.Icon != nil && page.Icon.Emoji != nil {
		title = fmt.Sprintf("%s %s", *page.Icon.Emoji, title)
	}
	return title
}
