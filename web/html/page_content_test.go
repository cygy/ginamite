package html

import "testing"

func TestPageContentHasTitle(t *testing.T) {
	pageContent := PageContent{}

	if pageContent.HasTitle() {
		t.Error("A PageContent struct with no defined title must not return true to 'HasTitle()'.")
	}

	pageContent.Title = "it is a title"
	if !pageContent.HasTitle() {
		t.Error("A PageContent struct with a defined title must return true to 'HasTitle()'.")
	}

	pageContent.Title = ""
	if pageContent.HasTitle() {
		t.Error("A PageContent struct with no defined title must not return true to 'HasTitle()'.")
	}
}

func TestPageContentHasInfo(t *testing.T) {
	pageContent := PageContent{}

	if pageContent.HasInfo() {
		t.Error("A PageContent struct with no defined info must not return true to 'HasInfo()'.")
	}

	pageContent.Info.Message = "it is an info"
	if !pageContent.HasInfo() {
		t.Error("A PageContent struct with a defined info must return true to 'HasInfo()'.")
	}

	pageContent.Info.Message = ""
	if pageContent.HasInfo() {
		t.Error("A PageContent struct with no defined info must not return true to 'HasInfo()'.")
	}
}

func TestPageContentHasError(t *testing.T) {
	pageContent := PageContent{}

	if pageContent.HasError() {
		t.Error("A PageContent struct with no defined error must not return true to 'HasError()'.")
	}

	pageContent.Error.Message = "it is an error"
	if !pageContent.HasError() {
		t.Error("A PageContent struct with a defined error must return true to 'HasError()'.")
	}

	pageContent.Error.Message = ""
	if pageContent.HasError() {
		t.Error("A PageContent struct with no defined error must not return true to 'HasError()'.")
	}
}

func TestPageContentIsUserAuthenticated(t *testing.T) {
	pageContent := PageContent{}

	if pageContent.IsUserAuthenticated() {
		t.Error("A PageContent struct with no defined user ID must not return true to 'IsUserAuthenticated()'.")
	}

	pageContent.CurrentUser.ID = "some user ID"
	if !pageContent.IsUserAuthenticated() {
		t.Error("A PageContent struct with a defined user ID must return true to 'IsUserAuthenticated()'.")
	}

	pageContent.CurrentUser.ID = ""
	if pageContent.IsUserAuthenticated() {
		t.Error("A PageContent struct with no defined user ID must not return true to 'IsUserAuthenticated()'.")
	}

	pageContent.CurrentUser.Name = "some username"
	if pageContent.IsUserAuthenticated() {
		t.Error("A PageContent struct with no defined user ID must not return true to 'IsUserAuthenticated()'.")
	}
}
