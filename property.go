package notiondb

import (
	"fmt"

	"github.com/jomei/notionapi"
)

// GetProperty は Propertiesから指定のIDのプロパティを返し、見つからなければnilを返します
func GetProperty(props notionapi.Properties, id notionapi.PropertyID) notionapi.Property {
	for _, prop := range props {
		if prop.GetID() == id.String() {
			return prop
		}
	}
	return nil
}

// getPropertyFromPage は、nameOrIDで指定された文字列がプロパティ名もしくはプロパティIDであると仮定して
// 既存のページからプロパティを取得します
// これは、ユーザーが指定したSynchronizeRequest.Propertiesのキーに対し、対応するプロパティを取得する際に使われます
// IDでの探索が名前の探索よりも優先されます
func getPropertyFromPage(page *notionapi.Page, nameOrID string) notionapi.Property {
	if prop := GetProperty(page.Properties, notionapi.PropertyID(nameOrID)); prop != nil {
		return prop
	}
	return page.Properties[nameOrID]
}

func compareProperty(newProp notionapi.Property, oldProp notionapi.Property) bool {
	switch newProp := newProp.(type) {
	case *notionapi.TitleProperty:
		oldProp := oldProp.(*notionapi.TitleProperty)
		return RichTextToString(newProp.Title) == RichTextToString(oldProp.Title)
	case *notionapi.NumberProperty:
		oldProp := oldProp.(*notionapi.NumberProperty)
		return newProp.Number == oldProp.Number
	case *notionapi.SelectProperty:
		oldProp := oldProp.(*notionapi.SelectProperty)
		if oldProp.Select.ID != "" && newProp.Select.ID == oldProp.Select.ID {
			return true
		}
		return newProp.Select.Name == oldProp.Select.Name
	case *notionapi.DateProperty:
		oldProp := oldProp.(*notionapi.DateProperty)
		compareDate := func(a, b *notionapi.Date) bool {
			if a != nil && b != nil {
				return a.String() == b.String()
			}
			return a == nil && b == nil
		}
		return compareDate(newProp.Date.Start, oldProp.Date.Start) && compareDate(newProp.Date.End, oldProp.Date.End)
	case *notionapi.RelationProperty:
		oldProp := oldProp.(*notionapi.RelationProperty)
		if len(oldProp.Relation) != len(newProp.Relation) {
			return false
		}
		for i := range oldProp.Relation {
			if oldProp.Relation[i].ID != newProp.Relation[i].ID {
				return false
			}
		}
		return true
	default:
		panic(fmt.Sprintf("未対応のプロパティ: %#v", newProp))
	}
}
