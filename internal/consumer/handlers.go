package consumer

// ContentTypeHandler takes care of data format translations from zipped, encoded or other data to JSON byte array
type ContentTypeHandler func([]byte) ([]byte, error)
