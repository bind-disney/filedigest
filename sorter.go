package main

type byId []*digestResponse

func (collection byId) Len() int {
	return len(collection)
}

func (collection byId) Swap(i, j int) {
	collection[i], collection[j] = collection[j], collection[i]
}

func (collection byId) Less(i, j int) bool {
	return collection[i] != nil && collection[j] != nil && collection[i].id < collection[j].id
}
