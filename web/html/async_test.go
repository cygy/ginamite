package html

import (
	"reflect"
	"testing"
)

func TestAsync_Prepare(t *testing.T) {
	async := Async{
		Paths: map[string]string{
			"en-us": "/{{.Locale}}/like",
			"fr-fr": "/{{.Locale}}/aimer",
			"de-de": "/{{.Locale}}/like",
		},
		Method: "POST",
		Target: "/likes/:id",
	}

	async.Prepare("https://web.test", []string{"en-us", "fr-fr", "es-es"})

	want := map[string]string{
		"en-us": "https://web.test/en-us/like",
		"fr-fr": "https://web.test/fr-fr/aimer",
	}
	if !reflect.DeepEqual(async.Paths, want) {
		t.Errorf("Paths = %v, want %v", async.Paths, want)
	}
	if async.Method != "POST" || async.Target != "/likes/:id" {
		t.Errorf("method = %q, target = %q", async.Method, async.Target)
	}
}
