package webfetch

import (
	"bytes"
	"context"
	"log"

	"github.com/chromedp/chromedp"
)

func FetchHtml(url string) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf bytes.Buffer
	if err := chromedp.Run(ctx,
		chromedp.Navigate("https://en.wikipedia.org/wiki/Main_Page"),
		//chromedp.WaitVisible(`#content`),
		//chromedp.Evaluate(s, nil),
		//chromedp.WaitVisible(`#thing`),
		chromedp.Dump(`document`, &buf, chromedp.ByJSPath),
	); err != nil {
		log.Fatal(err)
	}
}
