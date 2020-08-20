package html

import (
	"fmt"
)

func (meta *Meta) setUpHrefLangs(paths map[string]string, fallbackLocales map[string]string, host string) {
	meta.HrefLang = []HrefLang{}

	for locale, path := range paths {
		href := fmt.Sprintf("%s%s", host, path)
		hreflang := HrefLang{
			Href:   href,
			Locale: locale,
		}
		meta.HrefLang = append(meta.HrefLang, hreflang)
	}

	for language, locale := range fallbackLocales {
		path, ok := paths[locale]
		if !ok {
			continue
		}

		href := fmt.Sprintf("%s%s", host, path)
		hreflang := HrefLang{
			Href:   href,
			Locale: language,
		}
		meta.HrefLang = append(meta.HrefLang, hreflang)
	}
}

func (meta *Meta) setUpCanonicalURL(currentLocale string) {
	for _, hrefLang := range meta.HrefLang {
		if hrefLang.Locale == currentLocale {
			meta.CanonicalURL = hrefLang.Href
			break
		}
	}
}
