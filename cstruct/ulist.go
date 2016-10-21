package cstruct

type Appendable interface {
}

func AppendElements(slice []*ufile, data ...*ufile) []*ufile {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) {
		newSlice := make([]*ufile, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

func IndexById(slice []*ufile, id int64) []*ufile {
	filter := make([]*ufile, 0, 10)
	for _, uf := range slice {
		if uf.GetPointer().GetId() == id {
			filter = AppendElements(filter, uf)
		}
	}
	return filter
}
