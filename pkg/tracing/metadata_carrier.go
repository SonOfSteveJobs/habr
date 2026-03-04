package tracing

import "google.golang.org/grpc/metadata"

type MetadataCarrier metadata.MD

func (m MetadataCarrier) Get(key string) string {
	vals := metadata.MD(m).Get(key)
	if len(vals) == 0 {
		return ""
	}

	return vals[0]
}

func (m MetadataCarrier) Set(key, value string) {
	metadata.MD(m).Set(key, value)
}

func (m MetadataCarrier) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}
