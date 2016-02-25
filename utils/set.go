package utils

// TODO: Optimize this for stuff
type Set struct {
	items map[string]interface{}
}

func (set *Set) Add(key string, value interface{}) int {
	if set.items == nil {
		set.items = map[string]interface{}{}
	}
	if _, ok := set.items[key]; ok {
		return 0
	}
	set.items[key] = value
	return 1
}

func (set *Set) Remove(key string) int {
	if set.items == nil {
		set.items = map[string]interface{}{}
	}
	if _, ok := set.items[key]; !ok {
		return 0
	}
	delete(set.items, key)
	return 1
}

func (set *Set) Has(key string) bool {
	if set.items == nil {
		set.items = map[string]interface{}{}
	}
	_, ok := set.items[key]
	return ok
}

func (set *Set) Get(key string) (interface{}, bool) {
	if set.items == nil {
		set.items = map[string]interface{}{}
	}
	if value, ok := set.items[key]; ok {
		return value, true
	}
	return nil, false
}
