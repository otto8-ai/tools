package main

import (
	"github.com/PuerkitoBio/goquery"
)

var excludeNonMainTags = []string{
	"script, style, noscript, meta, head",
	"header", "footer", "nav", "aside", ".header", ".top", ".navbar", "#header",
	".footer", ".bottom", "#footer", ".sidebar", ".side", ".aside", "#sidebar",
	".modal", ".popup", "#modal", ".overlay", ".ad", ".ads", ".advert", "#ad",
	".lang-selector", ".language", "#language-selector", ".social", ".social-media",
	".social-links", "#social", ".menu", ".navigation", "#nav", ".breadcrumbs",
	"#breadcrumbs", "#search-form", ".search", "#search", ".share", "#share",
	".widget", "#widget", ".cookie", "#cookie",
}

func RemoveNonMainContent(doc *goquery.Selection) *goquery.Selection {
	for _, tag := range excludeNonMainTags {
		doc.Find(tag).Remove()
	}
	return doc
}
