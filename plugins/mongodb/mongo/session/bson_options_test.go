package session

import "testing"

func TestBSONOptions(t *testing.T) {
	opts := bsonOptions()

	if !opts.OmitZeroStruct {
		t.Error("the zero structs must be omitted")
	}

	if !opts.NilMapAsEmpty || !opts.NilSliceAsEmpty {
		t.Error("the nil maps and slices must be stored as empty")
	}
}
