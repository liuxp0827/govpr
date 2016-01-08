package gmm

type Distributor struct {
	index int
	score float64
}


func (d *Distributor) less(dis *Distributor) bool {
	if d.score < dis.score {
		return true
	} else if d.score == dis.score {
		if d.index < dis.index {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}


type Distributors []*Distributor

func (a Distributors) Len() int {
	return len(a)
}

func (a Distributors) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a Distributors) Less(i, j int) bool {
	return a[i].less(a[j])
}