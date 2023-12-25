package notiondb

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/jomei/notionapi"
)

type Synchronizer struct {
	client         *notionapi.Client
	dbID           notionapi.DatabaseID
	unmatchedPages []notionapi.Page
}

func NewSynchronizer(ctx context.Context, client *notionapi.Client, dbID notionapi.DatabaseID, filter notionapi.Filter) (*Synchronizer, error) {
	pages, err := QueryAll(ctx, client, dbID, filter)
	if err != nil {
		return nil, err
	}
	s := &Synchronizer{
		client:         client,
		dbID:           dbID,
		unmatchedPages: pages,
	}
	return s, nil
}

type SynchronizeChildren struct {
	DigestPropertyID string
	Blocks           []notionapi.Block
}
type SynchronizeRequest struct {
	Properties notionapi.Properties
	Children   *SynchronizeChildren
	Icon       *notionapi.Icon
	Cover      *notionapi.Image
	Matcher    func(page *notionapi.Page) bool
}
type SynchronizeResult struct {
	CreateRequest *notionapi.PageCreateRequest
	UpdateRequest *notionapi.PageUpdateRequest
	Page          *notionapi.Page
}

func (s *Synchronizer) Synchronize(ctx context.Context, request SynchronizeRequest) (*SynchronizeResult, error) {
	if request.Matcher == nil {
		return nil, fmt.Errorf("matcher is nil")
	}

	var page *notionapi.Page
	for i, p := range s.unmatchedPages {
		p := p
		if request.Matcher(&p) {
			page = &p
			// 未使用ページから削除
			s.unmatchedPages = append(s.unmatchedPages[0:i], s.unmatchedPages[i+1:]...)
			break
		}
	}

	if page != nil {
		// 更新リクエストを作成
		req := &notionapi.PageUpdateRequest{Properties: notionapi.Properties{}}

		// プロパティの比較
		for k, newProp := range request.Properties {
			if oldProp, ok := page.Properties[k]; !ok || !compareProperty(newProp, oldProp) {
				req.Properties[k] = newProp
			}
		}

		// アイコンの比較
		if request.Icon != nil {
			newIcon, _ := json.Marshal(request.Icon)
			oldIcon, _ := json.Marshal(page.Icon)
			if string(newIcon) != string(oldIcon) {
				req.Icon = request.Icon
			}
		}

		// 内容
		if request.Children != nil {
			oldDigest := RichTextToString(page.Properties[request.Children.DigestPropertyID].(*notionapi.RichTextProperty).RichText)
			if oldDigest != request.Children.getDigest() {
				// 既存のブロックを
				resp, err := s.client.Block.GetChildren(ctx, notionapi.BlockID(page.ID), nil)
				if err != nil {
					return nil, err
				}
				// 全部消す
				for _, b := range resp.Results {
					if _, err := s.client.Block.Delete(ctx, b.GetID()); err != nil {
						return nil, err
					}
				}
				abcReq := &notionapi.AppendBlockChildrenRequest{Children: request.Children.Blocks}
				if _, err := s.client.Block.AppendChildren(ctx, notionapi.BlockID(page.ID), abcReq); err != nil {
					return nil, err
				}
				req.Properties[request.Children.DigestPropertyID] = &notionapi.RichTextProperty{RichText: StringToRichText(request.Children.getDigest())}
			}
		}

		// プロパティ・アイコン・カバーが変更されないなら何もしない
		if len(req.Properties) == 0 && req.Icon == nil && req.Cover == nil {
			return &SynchronizeResult{Page: page}, nil
		}

		// 変更を試みる
		if page, err := s.client.Page.Update(ctx, notionapi.PageID(page.ID), req); err != nil {
			return nil, err
		} else {
			return &SynchronizeResult{UpdateRequest: req, Page: page}, err
		}
	} else {
		// 新規作成
		parent := notionapi.Parent{Type: notionapi.ParentTypeDatabaseID, DatabaseID: s.dbID}
		req := &notionapi.PageCreateRequest{Properties: request.Properties, Parent: parent}
		if request.Icon != nil {
			req.Icon = request.Icon
		}
		if request.Children != nil {
			req.Children = request.Children.Blocks
			req.Properties[request.Children.DigestPropertyID] = &notionapi.RichTextProperty{RichText: StringToRichText(request.Children.getDigest())}
		}
		if page, err := s.client.Page.Create(ctx, req); err != nil {
			return nil, err
		} else {
			return &SynchronizeResult{CreateRequest: req, Page: page}, nil
		}
	}
}

func (s *Synchronizer) UnmatchedPages() []notionapi.Page {
	return s.unmatchedPages
}

func (s *Synchronizer) DeleteUnmatchedPages(ctx context.Context) error {
	index := 0

	req := &notionapi.PageUpdateRequest{
		Properties: notionapi.Properties{},
		Archived:   true,
	}

	for i, p := range s.unmatchedPages {
		if _, err := s.client.Page.Update(ctx, notionapi.PageID(p.ID), req); err != nil {
			return err
		}
		index = i + 1
	}
	s.unmatchedPages = s.unmatchedPages[index:]
	return nil
}

func (c *SynchronizeChildren) getDigest() string {
	data, err := json.Marshal(c.Blocks)
	if err != nil {
		panic(err)
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
