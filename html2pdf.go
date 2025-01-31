package main

import (
	"bytes"
	"context"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func Html2pdf(html string, buf *bytes.Buffer, margin float64) error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var tasks = chromedp.Tasks{
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}

			return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			b, _, err := page.PrintToPDF().
				WithMarginLeft(margin).
				WithMarginRight(margin).
				WithMarginTop(margin).
				WithMarginBottom(margin).
				WithPrintBackground(true).
				Do(ctx)
			if err != nil {
				return err
			}
			if _, err := buf.Write(b); err != nil {
				return err
			}
			return nil
		}),
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		return err
	}
	return nil
}
