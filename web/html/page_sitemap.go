package html

// SitemapUpdateFrequency is the frequency of the updates of a page, as in the sitemaps.
type SitemapUpdateFrequency string

// String returns the frequency as a string.
func (frequency SitemapUpdateFrequency) String() string {
	return string(frequency)
}

const (
	// SitemapUpdateFrequencyAlways is the frequency of the pages changed at each access.
	SitemapUpdateFrequencyAlways = "always"
	// SitemapUpdateFrequencyHourly is the hourly frequency.
	SitemapUpdateFrequencyHourly = "hourly"
	// SitemapUpdateFrequencyDaily is the daily frequency.
	SitemapUpdateFrequencyDaily = "daily"
	// SitemapUpdateFrequencyWeekly is the weekly frequency.
	SitemapUpdateFrequencyWeekly = "weekly"
	// SitemapUpdateFrequencyMonthly is the monthly frequency.
	SitemapUpdateFrequencyMonthly = "monthly"
	// SitemapUpdateFrequencyYearly is the yearly frequency.
	SitemapUpdateFrequencyYearly = "yearly"
	// SitemapUpdateFrequencyNever is the frequency of the archived pages.
	SitemapUpdateFrequencyNever = "never"
)

// Sitemap holds the properties of a page in the sitemap.
type Sitemap struct {
	Included        bool
	Priority        float32
	UpdateFrequency SitemapUpdateFrequency `yaml:"update_frequency"`
}

// IsDefined reports whether the sitemap properties are defined, which is when the update frequency is.
func (sitemap Sitemap) IsDefined() bool {
	return len(sitemap.UpdateFrequency) > 0
}
