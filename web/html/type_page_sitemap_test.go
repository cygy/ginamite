package html

import "testing"

func TestPageSitemapCopy(t *testing.T) {
	original := Sitemap{
		Included:        true,
		Priority:        1.0,
		UpdateFrequency: SitemapUpdateFrequencyDaily,
	}

	copy := original

	if original.Included != copy.Included {
		t.Error("Copying a meta sitemap struct must not modify the Included property.")
	}
	if original.Priority != copy.Priority {
		t.Error("Copying a meta sitemap struct must not modify the Priority property.")
	}
	if original.UpdateFrequency != copy.UpdateFrequency {
		t.Error("Copying a meta sitemap struct must not modify the UpdateFrequency property.")
	}

	original.Included = !original.Included
	original.Priority = original.Priority * 2.0
	original.UpdateFrequency = SitemapUpdateFrequencyHourly

	if original.Included == copy.Included {
		t.Error("Modifying the Included property of a meta sitemap struct must not modify the Included property of the copied struct.")
	}
	if original.Priority == copy.Priority {
		t.Error("Modifying the Priority property of a meta sitemap struct must not modify the Priority property of the copied struct.")
	}
	if original.UpdateFrequency == copy.UpdateFrequency {
		t.Error("Modifying the UpdateFrequency property of a meta sitemap struct must not modify the UpdateFrequency property of the copied struct.")
	}
}

func TestPageSitemapIsDefined(t *testing.T) {
	sitemap := Sitemap{}

	if sitemap.IsDefined() {
		t.Error("An empty struct 'Sitemap' must not be considered as defined.")
	}

	sitemap.Included = true
	sitemap.Priority = 1.0
	sitemap.UpdateFrequency = SitemapUpdateFrequencyAlways

	if !sitemap.IsDefined() {
		t.Error("An initialized struct 'Sitemap' must be considered as defined.")
	}
}
