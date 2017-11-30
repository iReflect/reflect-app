package fields

import (
	"bytes"
	"database/sql/driver"
	"testing"
)

func TestScan(t *testing.T) {

	src := "{'String1', 29}"

	j1 := JSONB{}

	var (
		res driver.Value
		err error
	)

	if err = j1.Scan(interface{}([]byte(src))); err != nil {
		t.Fatalf("Error in scan - %s", err)
	}

	if res, err = j1.Value(); err != nil {
		t.Fatalf("Error in fetching value - %s", err)
	}

	if src != res {
		t.Fatalf("Source & Result Should be equal - %s, %s.", src, res)
	}
}

func TestMarshall(t *testing.T) {

	src := "{'String1', 29}"
	srcBytes := []byte(src)

	j1 := JSONB{}
	var (
		resBytes []byte
		err      error
	)

	if err = j1.UnmarshalJSON(srcBytes); err != nil {
		t.Fatalf("Error in UnmarshalJSON - %s", err)
	}

	if resBytes, err = j1.MarshalJSON(); err != nil {
		t.Fatalf("Error in MarshalJSON - %s", err)
	}

	if bytes.Compare(srcBytes, resBytes) != 0 {
		t.Fatalf("Source & Result Should be equal - %s, %s.", srcBytes, resBytes)
	}
}

func TestNull(t *testing.T) {

	j1 := JSONB{}
	if j1.IsNull() != true {
		t.Fatalf("Should be Null - %s", j1)
	}
}

func TestNotNull(t *testing.T) {

	j1 := JSONB([]byte("{'String1', 29}"))
	if j1.IsNull() == true {
		t.Fatalf("Should not be Null - %s", j1)
	}
}

func TestEquals(t *testing.T) {

	var srcBytes = []byte("{'String1', 29}")
	var j1, j2 JSONB

	j1 = JSONB(srcBytes)
	j2 = JSONB(srcBytes)

	var (
		resBytes []byte
		err      error
	)

	if resBytes, err = j1.MarshalJSON(); err != nil {
		t.Fatalf("Error in MarshalJSON - %s", err)
	}

	if j1.IsNull() || bytes.Compare(srcBytes, resBytes) != 0 || j1.Equals(j2) != true {
		t.Fatalf("Source & Result Should be non-null and equal - %s, %s.", j1, j2)
	}
}

func TestEqualsNul(t *testing.T) {

	var j1, j2 JSONB

	if !j1.IsNull() || j1.Equals(j2) != true {
		t.Fatalf("Source & Result Should be null and equal - %s, %s.", j1, j2)
	}
}
