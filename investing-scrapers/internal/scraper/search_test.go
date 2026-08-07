package scraper

import (
	"testing"
)

func TestParseSearchResponse(t *testing.T) {
	body := `
{
  "articles": [
    {"id": 200476379, "url": "/analysis/article-200476379", "description": "黄力晨：黄金继续看跌", "image": "https://d6-invdn-com.investing.com/img1.jpg"},
    {"id": 200476380, "url": "/analysis/article-200476380", "description": "原油市场分析", "image": "https://d6-invdn-com.investing.com/img2.jpg"}
  ],
  "news": [
    {"id": 4835697, "url": "/news/news-4835697", "description": "美联储会议纪要", "image": ""}
  ]
}
`

	resp := parseSearchResponse(body)
	if resp == nil {
		t.Fatal("parseSearchResponse returned nil")
	}

	if len(resp.Articles) != 2 {
		t.Errorf("Articles = %d, want %d", len(resp.Articles), 2)
	}
	if len(resp.News) != 1 {
		t.Errorf("News = %d, want %d", len(resp.News), 1)
	}

	if resp.Articles[0].ID != 200476379 {
		t.Errorf("Article[0].ID = %d, want %d", resp.Articles[0].ID, 200476379)
	}
	if resp.Articles[0].URL != "/analysis/article-200476379" {
		t.Errorf("Article[0].URL = %q, want %q", resp.Articles[0].URL, "/analysis/article-200476379")
	}
	if resp.Articles[0].Description != "黄力晨：黄金继续看跌" {
		t.Errorf("Article[0].Description = %q, want %q", resp.Articles[0].Description, "黄力晨：黄金继续看跌")
	}
	if resp.Articles[0].Image != "https://d6-invdn-com.investing.com/img1.jpg" {
		t.Errorf("Article[0].Image = %q", resp.Articles[0].Image)
	}
	if resp.Articles[0].Type != "articles" {
		t.Errorf("Article[0].Type = %q, want %q", resp.Articles[0].Type, "articles")
	}

	if resp.News[0].ID != 4835697 {
		t.Errorf("News[0].ID = %d, want %d", resp.News[0].ID, 4835697)
	}
	if resp.News[0].Type != "news" {
		t.Errorf("News[0].Type = %q, want %q", resp.News[0].Type, "news")
	}
}

func TestParseSearchResponse_Empty(t *testing.T) {
	body := `{"articles": [], "news": []}`
	resp := parseSearchResponse(body)
	if resp == nil {
		t.Fatal("parseSearchResponse returned nil")
	}
	if len(resp.Articles) != 0 {
		t.Errorf("Articles = %d, want 0", len(resp.Articles))
	}
	if len(resp.News) != 0 {
		t.Errorf("News = %d, want 0", len(resp.News))
	}
}

func TestParseSearchResponse_InvalidJSON(t *testing.T) {
	resp := parseSearchResponse(`{invalid json}`)
	if resp != nil {
		t.Error("expected nil for invalid JSON")
	}
}

func TestParseSearchResponse_NoArticles(t *testing.T) {
	body := `{"news": [{"id": 123, "url": "/news/123", "description": "test"}]}`
	resp := parseSearchResponse(body)
	if resp == nil {
		t.Fatal("parseSearchResponse returned nil")
	}
	if len(resp.News) != 1 {
		t.Errorf("News = %d, want 1", len(resp.News))
	}
}
