package htadaptor

import (
	"errors"
	"fmt"
)

func WithReadLimit(n int64) UnaryOrVoidOption {
	return readLimitOption(n)
}

type readLimitOption int64

func (r readLimitOption) validate() error {
	if r < 1 || r > 1<<40 {
		return fmt.Errorf("invalid read limit: %d", r)
	}
	return nil
}

func (r readLimitOption) applyUnaryOption(o *unaryOptions) (err error) {
	if o.ReadLimit != 0 {
		return errors.New("read limit is already set")
	}
	if err = r.validate(); err != nil {
		return err
	}
	o.ReadLimit = int64(r)
	return nil
}

func (r readLimitOption) applyVoidOption(o *voidOptions) (err error) {
	if o.ReadLimit != 0 {
		return errors.New("read limit is already set")
	}
	if err = r.validate(); err != nil {
		return err
	}
	o.ReadLimit = int64(r)
	return nil
}
