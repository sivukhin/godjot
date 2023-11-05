package tokenizer

// Attributes - is a map with support for deterministic order of enumeration
type Attributes struct {
	Keys []string
	Map  map[string]string
}

type AttributeEntry struct{ Key, Value string }

func (a *Attributes) Size() int {
	if a == nil {
		return 0
	}
	return len(a.Keys)
}

func (a *Attributes) Set(key, value string) *Attributes {
	if _, ok := a.Map[key]; !ok {
		a.Keys = append(a.Keys, key)
	}
	if a.Map == nil {
		a.Map = make(map[string]string)
	}
	a.Map[key] = value
	return a
}

func (a *Attributes) TryGet(key string) (string, bool) {
	value, ok := a.Map[key]
	return value, ok
}

func (a *Attributes) Get(key string) string {
	return a.Map[key]
}

func (a *Attributes) MergeWith(other *Attributes) *Attributes {
	if other != nil {
		for _, key := range other.Keys {
			a.Set(key, other.Get(key))
		}
	}
	return a
}

func (a *Attributes) Entries() []AttributeEntry {
	entries := make([]AttributeEntry, 0, len(a.Map))
	for _, key := range a.Keys {
		entries = append(entries, AttributeEntry{Key: key, Value: a.Get(key)})
	}
	return entries
}
