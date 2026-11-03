package model

import (
	"bytes"
	"testing"

	mongoBSON "go.mongodb.org/mongo-driver/v2/bson"
)

// marshalOmittingZeroStructs marshals the value like the client of the session does.
func marshalOmittingZeroStructs(t *testing.T, v any) []byte {
	t.Helper()

	var buf bytes.Buffer
	vw := mongoBSON.NewDocumentWriter(&buf)
	encoder := mongoBSON.NewEncoder(vw)
	encoder.OmitZeroStruct()

	if err := encoder.Encode(v); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	return buf.Bytes()
}

// TestUser_BSONOmitsZeroStructs checks that the empty OAuth infos are not stored, otherwise the unique partial
// indexes on identificationMethods.<network>.userId would reject the second user without this network.
func TestUser_BSONOmitsZeroStructs(t *testing.T) {
	data := marshalOmittingZeroStructs(t, NewUser())

	var doc mongoBSON.M
	if err := mongoBSON.Unmarshal(data, &doc); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	for _, key := range []string{"identificationMethods", "publicInfos", "privateInfos", "registrationInfos"} {
		if _, found := doc[key]; found {
			t.Errorf("the zero struct %q must not be stored", key)
		}
	}

	user := NewUser()
	user.IdentificationMethods.Password = "hash"
	data = marshalOmittingZeroStructs(t, user)

	doc = mongoBSON.M{}
	if err := mongoBSON.Unmarshal(data, &doc); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	methods, _ := doc["identificationMethods"].(mongoBSON.M)
	for _, network := range []string{"facebook", "google", "twitter"} {
		if _, found := methods[network]; found {
			t.Errorf("the zero OAuth infos %q must not be stored", network)
		}
	}
}
