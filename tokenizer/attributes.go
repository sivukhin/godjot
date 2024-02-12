package tokenizer

// Attributes - is a map with support for deterministic order of enumeration
type Attributes struct {
	Keys []string
	Map  map[string]string
}

type AttributeEntry struct{ Key, Value string }

func NewAttributes(entries ...AttributeEntry) Attributes {
	attributes := Attributes{}
	for _, entry := range entries {
		attributes.Set(entry.Key, entry.Value)
	}
	return attributes
}

func (a *Attributes) Size() int {
	return len(a.Keys)
}

func (a *Attributes) Append(key, value string) {
	if previous, ok := a.TryGet(key); ok {
		a.Set(key, previous+" "+value)
	} else {
		a.Set(key, value)
	}
}

func (a *Attributes) Set(key, value string) {
	if _, ok := a.Map[key]; !ok {
		a.Keys = append(a.Keys, key)
	}
	if a.Map == nil {
		a.Map = make(map[string]string)
	}
	a.Map[key] = value
}

func (a *Attributes) TryGet(key string) (string, bool) {
	value, ok := a.Map[key]
	return value, ok
}

func (a *Attributes) Get(key string) string {
	return a.Map[key]
}

func (a *Attributes) MergeWith(other Attributes) {
	for _, key := range other.Keys {
		a.Set(key, other.Get(key))
	}
}

func (a *Attributes) Entries() []AttributeEntry {
	if len(a.Keys) == 0 {
		return nil
	}
	entries := make([]AttributeEntry, 0, len(a.Map))
	for _, key := range a.Keys {
		entries = append(entries, AttributeEntry{Key: key, Value: a.Get(key)})
	}
	return entries
}

func (a *Attributes) GoMap() map[string]string {
	if len(a.Keys) == 0 {
		return nil
	}
	entries := make(map[string]string)
	for _, key := range a.Keys {
		entries[key] = a.Get(key)
	}
	return entries
}
