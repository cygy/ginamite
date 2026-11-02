package functions

import (
	"bytes"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/api/ctxsettings"

	"github.com/cygy/ginamite/common/ctxkey"
	"github.com/cygy/ginamite/plugins/mongodb/internal/plugintest"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

// testSettings are the settings of the plugin of the requests, replaced by withImageDirectory.
var testSettings model.Settings

// withImageDirectory saves the user images in a temporary directory for the duration of the test.
func withImageDirectory(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	testSettings = model.DefaultSettings()
	testSettings.UserImagePath = dir

	plugintest.InitLocalization(t, map[string]string{"status.account.public.image.updated": "updated", "status.account.public.image.deleted": "deleted"})

	return dir
}

// pngFile returns the content of a PNG image.
func pngFile(t *testing.T) []byte {
	t.Helper()

	var buffer bytes.Buffer
	if err := png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, 20, 20))); err != nil {
		t.Fatal(err)
	}

	return buffer.Bytes()
}

// uploadImageRequest sends the file as the public image of the user.
func uploadImageRequest(t *testing.T, db *mock.Database, id bson.ObjectID, content []byte) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if content != nil {
		part, err := writer.CreateFormFile("file", "avatar.png")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	c, recorder := plugintest.SignedInContext(t, db, model.User{ID: id})
	ctxkey.Config.Get(c).Hosts.Static = "https://static.example.com"
	ctxsettings.SetSettings(c, testSettings)
	c.Request = httptest.NewRequest(http.MethodPost, "/", &body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	UpdateAccountPublicImage(c)

	return recorder
}

// files returns the names of the files of the directory and its sub-directories.
func files(t *testing.T, dir string) []string {
	t.Helper()

	var names []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			names = append(names, path)
		}

		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	return names
}

func userWithImage(id bson.ObjectID, paths ...string) bson.M {
	image := bson.M{}
	for i, name := range []string{"full", "small", "thumb"} {
		image[name] = bson.M{"size1x": bson.M{"path": paths[i*2], "url": "u"}, "size2x": bson.M{"path": paths[i*2+1], "url": "u"}}
	}

	return bson.M{"_id": id, "publicInfos": bson.M{"image": image}}
}

func TestUpdateAccountPublicImage(t *testing.T) {
	id := bson.NewObjectID()

	t.Run("saves the image", func(t *testing.T) {
		dir := withImageDirectory(t)
		queue := plugintest.CaptureQueue(t)

		// The user has an image already: it is replaced.
		old := []string{"old-full1.png", "old-full2.png", "old-small1.png", "old-small2.png", "old-thumb1.png", "old-thumb2.png"}
		for _, name := range old {
			if err := os.WriteFile(filepath.Join(dir, name), []byte("old"), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).One = mock.Result{Value: userWithImage(id, old...)}

		recorder := uploadImageRequest(t, db, id, pngFile(t))
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "https://static.example.com") || !strings.Contains(recorder.Body.String(), "updated") {
			t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
		}

		set := db.Coll(model.UserCollection).LastCall().Args[1].(bson.M)["$set"].(bson.M)
		if len(set) != 12 || !strings.HasPrefix(set["publicInfos.image.full.size1x.url"].(string), "https://static.example.com") {
			t.Errorf("set = %v", set)
		}

		// The new versions are saved, the old ones are deleted.
		saved := files(t, dir)
		if len(saved) != 6 {
			t.Errorf("files = %v, want the 6 new versions", saved)
		}
		for _, path := range saved {
			if strings.Contains(path, "old-") {
				t.Errorf("the old image %s must be deleted", path)
			}
		}

		if message, ok := queue.Next(); !ok || message.Type != "update_user_public_image" {
			t.Errorf("message = %+v, %v", message, ok)
		}
	})

	t.Run("database error deletes the new image", func(t *testing.T) {
		dir := withImageDirectory(t)
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": id}}
		db.Coll(model.UserCollection).Errs = map[string]error{"UpdateByID": errBoom}

		if recorder := uploadImageRequest(t, db, id, pngFile(t)); recorder.Code != http.StatusInternalServerError {
			t.Errorf("status = %d", recorder.Code)
		}
		if saved := files(t, dir); len(saved) != 0 {
			t.Errorf("files = %v, want none", saved)
		}
	})

	t.Run("no file", func(t *testing.T) {
		withImageDirectory(t)
		if recorder := uploadImageRequest(t, mock.NewDatabase("app"), id, nil); recorder.Code == http.StatusOK {
			t.Error("a missing file must be rejected")
		}
	})

	t.Run("file too big", func(t *testing.T) {
		withImageDirectory(t)
		testSettings.UserMaxImageSize = 10

		if recorder := uploadImageRequest(t, mock.NewDatabase("app"), id, pngFile(t)); recorder.Code != http.StatusPreconditionFailed {
			t.Errorf("status = %d", recorder.Code)
		}
	})

	t.Run("not an image", func(t *testing.T) {
		withImageDirectory(t)
		if recorder := uploadImageRequest(t, mock.NewDatabase("app"), id, []byte(strings.Repeat("text ", 200))); recorder.Code != http.StatusPreconditionFailed {
			t.Errorf("status = %d", recorder.Code)
		}
	})

	t.Run("corrupted image", func(t *testing.T) {
		withImageDirectory(t)
		corrupted := pngFile(t)[:60]

		if recorder := uploadImageRequest(t, mock.NewDatabase("app"), id, corrupted); recorder.Code != http.StatusInternalServerError {
			t.Errorf("status = %d", recorder.Code)
		}
	})

	t.Run("unknown user", func(t *testing.T) {
		dir := withImageDirectory(t)

		if recorder := uploadImageRequest(t, mock.NewDatabase("app"), id, pngFile(t)); recorder.Code != http.StatusInternalServerError {
			t.Errorf("status = %d", recorder.Code)
		}
		if got := files(t, dir); len(got) != 0 {
			t.Errorf("files = %v, want none", got)
		}
	})
}

func TestDeleteAccountPublicImage(t *testing.T) {
	id := bson.NewObjectID()

	deleteImage := func(t *testing.T, db *mock.Database) *httptest.ResponseRecorder {
		t.Helper()

		c, recorder := plugintest.SignedInContext(t, db, model.User{ID: id})
		ctxsettings.SetSettings(c, testSettings)
		DeleteAccountPublicImage(c)

		return recorder
	}

	t.Run("deletes the image", func(t *testing.T) {
		dir := withImageDirectory(t)
		queue := plugintest.CaptureQueue(t)

		names := []string{"full1.png", "full2.png", "small1.png", "small2.png", "thumb1.png", "thumb2.png"}
		for _, name := range names {
			if err := os.WriteFile(filepath.Join(dir, name), []byte("image"), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).One = mock.Result{Value: userWithImage(id, names...)}

		recorder := deleteImage(t, db)
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "deleted") {
			t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
		}

		set := db.Coll(model.UserCollection).LastCall().Args[1].(bson.M)["$set"].(bson.M)
		if set["publicInfos.image.full.size1x.url"] != "" || set["publicInfos.image.full.size1x.path"] != "" {
			t.Errorf("the image must be cleared: %v", set)
		}
		if remaining := files(t, dir); len(remaining) != 0 {
			t.Errorf("files = %v, want none", remaining)
		}
		if message, ok := queue.Next(); !ok || message.Type != "update_user_public_image" {
			t.Errorf("message = %+v, %v", message, ok)
		}
	})

	t.Run("database error keeps the files", func(t *testing.T) {
		dir := withImageDirectory(t)
		if err := os.WriteFile(filepath.Join(dir, "full1.png"), []byte("image"), 0o600); err != nil {
			t.Fatal(err)
		}
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).One = mock.Result{Value: userWithImage(id, "full1.png", "", "", "", "", "")}
		db.Coll(model.UserCollection).Errs = map[string]error{"UpdateByID": errBoom}

		if recorder := deleteImage(t, db); recorder.Code != http.StatusInternalServerError {
			t.Errorf("status = %d", recorder.Code)
		}
		if remaining := files(t, dir); len(remaining) != 1 {
			t.Errorf("files = %v, want the image kept", remaining)
		}
	})

	t.Run("unknown user", func(t *testing.T) {
		withImageDirectory(t)
		if recorder := deleteImage(t, mock.NewDatabase("app")); recorder.Code != http.StatusInternalServerError {
			t.Errorf("status = %d", recorder.Code)
		}
	})
}
