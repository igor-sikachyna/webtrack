package webfetch

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type Fetcher struct {
	ctx     context.Context
	cancel  context.CancelFunc
	backend string
}

func NewFetcher(backend string) (result Fetcher) {
	result.backend = backend
	switch backend {
	case "chrome":
		result.ctx, result.cancel = chromedp.NewContext(context.Background())
	case "go":
		// No action required
	default:
		// Unknown backend
		log.Fatal("Unknown backend: ", backend)
	}
	return
}

func (f *Fetcher) Close() {
	if f.cancel != nil {
		f.cancel()
	}
}

func (f *Fetcher) FetchHtml(url string) (res string, err error) {
	// NewFetcher should have validated backend field
	switch f.backend {
	case "chrome":
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
	default:
		var resp *http.Response
		resp, err = http.Get(url)
		if err != nil {
			return
		}
		var resBytes []byte
		resBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return
		}
		res = string(resBytes[:])
	}

	return
}
