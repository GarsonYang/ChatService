package indexes

type int64set map[int64]struct{}

//add adds a value to the set and returns
//true if the value didn't already exist in the set
func (s int64set) add(value int64) bool {
	//'ok' is true if value is within set
	_, ok := s[value]
	if ok {
		return false
	}

	s[value] = struct{}{}
	return true
}

//remove removes a value from the set and returns
//true if that value was in the set, false otehrwise
func (s int64set) remove(value int64) bool {
	_, ok := s[value]
	if !ok {
		return false
	}
	delete(s, value)
	return true
}

//has returns true if value is in the set,
//or false if it is not in the set
func (s int64set) has(value int64) bool {
	//'ok' is true if value is within set
	_, ok := s[value]
	return ok
}

func (s int64set) all() []int64 {
	values := make([]int64, 0, len(s))
	for k := range s {
		values = append(values, k)
	}

	return values
}
