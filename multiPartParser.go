package vurixwebapiclient

import "sync"

type MultiPartParser struct {
	mutex *sync.Mutex
	buff  []byte
}

func NewMultiPartParser() *MultiPartParser {
	return &MultiPartParser{
		mutex: &sync.Mutex{},
		buff:  make([]byte, 0),
	}
}

func (*MultiPartParser) Put(data []byte) {

}
