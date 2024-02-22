package secrets

import (
	"bytes"
	"encoding/gob"
)

type Snapshot struct {
	Future  *Secret
	Present *Secret
	Past    *Secret
}

func (s *Snapshot) Serialize() ([]byte, error) {
	b := &bytes.Buffer{}
	if err := gob.NewEncoder(b).Encode(s); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func NewSnapshot(b []byte) (s *Snapshot, err error) {
	err = gob.NewDecoder(bytes.NewReader(b)).Decode(s)
	return
}
