package govpr

import "fmt"

type T interface {
	Compare(T) int
	Add(T)
	Sub(T)
	Equal(T) bool
}

type Vector struct {
	_size     int64
	_capacity int64
	_array    []T
}

func NewVector(_capacity int64, _size int64) *Vector {
	if _capacity < 1 {
		_capacity = 1
	}

	vector := &Vector{
		_size:     0,
		_capacity: _capacity,
	}

	vector._array = vector.createArray()
	return vector
}

func NewVector1(v *Vector) *Vector {
	_capacity := 1
	if v._size > 1 {
		_capacity = v._size
	}

	vector := &Vector{
		_size:     v._size,
		_capacity: _capacity,
	}

	copy(vector._array, v._array)
	return vector
}

// 模拟+=运算
func (this *Vector) Add(v *Vector) (*Vector, error) {
	if this._size != v._size {
		return nil, fmt.Errorf("Mismatch vector sizes")
	}

	for i := int64(0); i < this._size; i++ {
		/* 两个向量累加 */
		this._array[i].Add(v._array[i])

		//		this._array[i]+=v._array[i]
	}
	return this, nil
}

// 模拟-=运算
func (this *Vector) Sub(v *Vector) (*Vector, error) {
	if this._size != v._size {
		return nil, fmt.Errorf("Mismatch vector sizes")
	}

	for i := int64(0); i < this._size; i++ {

		this._array[i].Sub(v._array[i])

		//		this._array[i]-=v._array[i]
	}
	return this, nil
}

// 模拟==运算
func (this *Vector) Equal(v *Vector) bool {
	if this._size != v._size {
		return false
	}

	for i := int64(0); i < this._size; i++ {

		if !this._array[i].Equal(v._array[i]) {
			//	if	this._array[i]!=v._array[i]{
			return false
		}

	}
	return true
}

func (this *Vector) Size() int64 {
	return this._size
}

func (this *Vector) Clear() {
	this._size = 0
}

func (this *Vector) SetSize(size int64, updateCapacity bool) error {
	if this._array == nil {
		fmt.Errorf("Vector array is nil")
	}

	if size > this._capacity || (size < this._capacity && updateCapacity) {
		oldSize := this._size
		this._size = size
		this._capacity = this._size
		oldArray := this._array
		var copyLen int64
		if size > oldSize {
			copyLen = oldSize
		} else {
			copyLen = size
		}

		this._array = this.createArray()
		copy(this._array, oldArray[:copyLen])
	}
	return nil
}

func (this *Vector) AddValue(t ...T) error {
	if this._array == nil {
		fmt.Errorf("Vector array is nil")
	}

	this._array = append(this._array, t...)
	return nil
}

func (this *Vector) SetValues(v *Vector) error {
	if this._size != v._size {
		return fmt.Errorf("Cannot set values: vector size mismatch")
	}
	copy(this._array, v._array)
	return nil
}

func (this *Vector) SetAllValues(t T) {
	for i := int64(0); i < this._size; i++ {
		this._array[i] = t
	}
}

func (this *Vector) Multi(s float64) {
	for i := int64(0); i < this._size; i++ {
		this._array[i] *= s
	}
}

func (this *Vector) createArray() []T {
	if this._capacity == 0 {
		panic("Vector capacity can not be 0")
	}
	return make([]T, this._capacity)
}

func (this *Vector) compare(t1 T, t2 T) int {
	return t1.Compare(t2)
}
