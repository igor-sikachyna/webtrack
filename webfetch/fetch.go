package webfetch

import (
	"context"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type Fetcher struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewFetcher() (result Fetcher) {
	result.ctx, result.cancel = chromedp.NewContext(context.Background())
	return
}

func (f *Fetcher) Close() {
	f.cancel()
}

func (f *Fetcher) FetchHtml(url string) (res string, err error) {
	err = chromedp.Run(f.ctx,
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			res, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			return err
		}),
	)

	return
}
