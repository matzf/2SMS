package types

type Mapping struct {
	Path string
	Port string
}

type EndpointMappings map[string]string

func (mp *Mapping) Equal(mp_b *Mapping) bool {
	return mp.Path == mp_b.Path && mp.Port == mp_b.Port
}
