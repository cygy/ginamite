package html

// SitemapUpdateFrequency : update frequencies of a sitemap
type SitemapUpdateFrequency string

// String : returns the string value of the frequency.
func (frequency SitemapUpdateFrequency) String() string {
	return string(frequency)
}

const (
	// SitemapUpdateFrequencyAlways always frequency
	SitemapUpdateFrequencyAlways = "always"
	// SitemapUpdateFrequencyHourly hourly frequency
	SitemapUpdateFrequencyHourly = "hourly"
	// SitemapUpdateFrequencyDaily daily frequency
	SitemapUpdateFrequencyDaily = "daily"
	// SitemapUpdateFrequencyWeekly weekly frequency
	SitemapUpdateFrequencyWeekly = "weekly"
	// SitemapUpdateFrequencyMonthly monthly frequency
	SitemapUpdateFrequencyMonthly = "monthly"
	// SitemapUpdateFrequencyYearly yearly frequency
	SitemapUpdateFrequencyYearly = "yearly"
	// SitemapUpdateFrequencyNever never frequency
	SitemapUpdateFrequencyNever = "never"
)

// Sitemap : struct of the sitemap properties.
type Sitemap struct {
	Included        bool
	Priority        float32
	UpdateFrequency SitemapUpdateFrequency `yaml:"update_frequency"`
}

// IsDefined : returns true if the struct is initialized.
func (sitemap Sitemap) IsDefined() bool {
	return len(sitemap.UpdateFrequency) > 0
}
