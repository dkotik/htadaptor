package htadaptor

const oneMB = 1 << 20

type unaryOptions struct {
	ReadLimit int64
	// Decoders map[string]Decoder
	// Finalizers []Finalizer
	// Encoders map[string]Encoder
}

type voidOptions struct {
	ReadLimit int64
	// Decoders map[string]Decoder
	// Finalizers []Finalizer
}

type UnaryOption interface {
	applyUnaryOption(*unaryOptions) error
}

type VoidOption interface {
	applyVoidOption(*voidOptions) error
}

type UnaryOrVoidOption interface {
	UnaryOption
	VoidOption
}
