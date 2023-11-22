# Hyper Text Adaptors

Package `htadaptor` provides generic domain logic adaptors for HTTP handlers. Adaptors come in three flavors:

1. UnaryFunc: func(context, inputStruct) (outputStruct, error)
2. NullaryFunc: func(context) (outputStruct, error)
3. VoidFunc: func(context, inputStruct) error

Each input requires implementation of `htadaptor.Validatable` for safety.
