package htadaptor

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Encoder[T any] interface {
	Encode(http.ResponseWriter, T) error
}

type Decoder[T any, V Validatable[T], O any] interface {
	Decode(http.ResponseWriter, *http.Request) (V, Encoder[O], error)
}

type Codec[T any, V Validatable[T], O any] interface {
	Encoder[O]
	Decoder[T, V, O]
}

type Finalizer[T any, V Validatable[T]] func(V, *http.Request) error

func Finalize[T any, V Validatable[T], O any](
	decoder Decoder[T, V, O],
	finalizers ...Finalizer[T, V],
) (Decoder[T, V, O], error) {
	var zero Decoder[T, V, O]
	if decoder == zero {
		return zero, errors.New("cannot use a <nil> decoder")
	}

	for _, f := range finalizers {
		if f == nil {
			return zero, errors.New("cannot use a <nil> finalizer")
		}
	}

	switch len(finalizers) {
	case 0:
		return zero, errors.New("at least one finalizer is required")
	case 1:
		return &singleFinalizer[T, V, O]{
			Decoder:   decoder,
			Finalizer: finalizers[0],
		}, nil
	default:
		return &multiFinalizer[T, V, O]{
			Decoder:    decoder,
			Finalizers: finalizers,
		}, nil
	}
}

type singleFinalizer[T any, V Validatable[T], O any] struct {
	Decoder   Decoder[T, V, O]
	Finalizer Finalizer[T, V]
}

func (f *singleFinalizer[T, V, O]) Decode(
	w http.ResponseWriter,
	r *http.Request,
) (V, Encoder[O], error) {
	result, encoder, err := f.Decoder.Decode(w, r)
	if err != nil {
		return nil, nil, err
	}
	if err = f.Finalizer(result, r); err != nil {
		return nil, nil, err
	}
	return result, encoder, nil
}

type multiFinalizer[T any, V Validatable[T], O any] struct {
	Decoder    Decoder[T, V, O]
	Finalizers []Finalizer[T, V]
}

func (f *multiFinalizer[T, V, O]) Decode(
	w http.ResponseWriter,
	r *http.Request,
) (V, Encoder[O], error) {
	result, encoder, err := f.Decoder.Decode(w, r)
	if err != nil {
		return nil, nil, err
	}
	for _, finalizer := range f.Finalizers {
		if err = finalizer(result, r); err != nil {
			return nil, nil, err
		}
	}
	return result, encoder, nil
}

type jsonCodec[T any, V Validatable[T], O any] struct{}

func NewJSONCodec[T any, V Validatable[T], O any]() Codec[T, V, O] {
	return jsonCodec[T, V, O]{}
}

func (j jsonCodec[T, V, O]) Encode(w http.ResponseWriter, value O) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(&value)
}

func (j jsonCodec[T, V, O]) Decode(
	w http.ResponseWriter,
	r *http.Request,
) (V, Encoder[O], error) {
	var request V
	// err := json.NewDecoder(http.MaxBytesReader(w, r.Body, int64(j))).Decode(&request)
	err := json.NewDecoder(r.Body).Decode(&request)
	defer func() {
		r.Body.Close()
	}()
	if err != nil {
		return nil, nil, err
	}
	return request, j, nil
}

type jsonEncoder[O any] struct{}

func NewJSONEncoder[O any]() Encoder[O] {
	return jsonEncoder[O]{}
}

func (j jsonEncoder[O]) Encode(w http.ResponseWriter, value O) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(&value)
}
