package notiondb

import (
	"strings"

	"github.com/jomei/notionapi"
)

func RichTextToString(richText []notionapi.RichText) string {
	strs := []string{}
	for _, rt := range richText {
		strs = append(strs, rt.PlainText)
	}
	return strings.Join(strs, "")
}

func StringToRichText(text string, options ...stringToRichTextOption) []notionapi.RichText {
	o := &stringToRichTextOptions{}
	for _, option := range options {
		option(o)
	}
	return []notionapi.RichText{{Text: &notionapi.Text{Content: text}, PlainText: text, Annotations: o.annotation}}
}

type stringToRichTextOptions struct {
	annotation *notionapi.Annotations
}
type stringToRichTextOption func(o *stringToRichTextOptions)

func WithAnnotation(annotation *notionapi.Annotations) stringToRichTextOption {
	return func(o *stringToRichTextOptions) {
		o.annotation = annotation
	}
}
