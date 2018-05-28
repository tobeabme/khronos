package khronos

type Processor struct {
	Application       string
	NodeName          string
	IP                string
	Port              int
	Status            bool
	MaxExecutionLimit int
	Undone            int
}

const MaxExecutionLimit = 10

type ProcessorList []*Processor

func (pl ProcessorList) Len() int {
	return len(pl)
}

func (pl ProcessorList) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}

func (pl ProcessorList) Less(i, j int) bool {
	return pl[i].Undone < pl[j].Undone
}

// type int64arr []int64

// func (a int64arr) Len() int           { return len(a) }
// func (a int64arr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
// func (a int64arr) Less(i, j int) bool { return a[i] < a[j] }
