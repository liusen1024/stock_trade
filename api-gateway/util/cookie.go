package util

import (
	"context"
	"net/http"
	"stock/api-gateway/db"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

func XueQiuCookie() ([]*http.Cookie, error) {
	ctx := context.Background()
	var cookies []*http.Cookie
	if err := db.GetOrLoad(ctx, "xueqiu_cookies", 10*time.Minute, &cookies, func() error {
		newFetcher := func() *colly.Collector {
			fetcher := colly.NewCollector()
			fetcher.AllowURLRevisit = true
			extensions.Referer(fetcher)
			extensions.RandomUserAgent(fetcher)
			return fetcher
		}
		getCookies := func() ([]*http.Cookie, error) {
			fetcher := newFetcher()
			var cookie []*http.Cookie
			fetcher.OnResponse(func(response *colly.Response) {
				cookie = fetcher.Cookies(response.Request.URL.String())
			})
			if err := fetcher.Visit("https://xueqiu.com/"); err != nil {
				return nil, err
			}
			return cookie, nil
		}
		result, err := getCookies()
		if err != nil {
			return err
		}
		cookies = result
		return nil
	}); err != nil {
		return nil, err
	}

	return cookies, nil
}
