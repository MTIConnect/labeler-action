package label

// Operations stores a set of actions that can be used against
// a string slice to remove and add strings as if it was a set.
type Operations struct {
	Remove []string
	Set    []string
}

// Apply will iteratively remove labels in the remove set and
// then add items not already in the passed in string slice.
func (o Operations) Apply(labels []string) []string {
	labels = removeStrings(labels, o.Remove)
	labels = setStrings(labels, o.Set)
	return labels
}

func removeStrings(labels, remove []string) []string {
	n := 0
	for _, label := range labels {
		var keep = true
		for _, other := range remove {
			if label == other {
				keep = false
				break
			}
		}

		if keep {
			labels[n] = label
			n++
		}
	}
	return labels[:n]
}

func setStrings(labels, set []string) []string {
	for _, other := range set {
		var add = true
		for _, label := range labels {
			if label == other {
				add = false
				break
			}
		}

		if add {
			labels = append(labels, other)
		}
	}
	return labels
}
