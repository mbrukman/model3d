// Generated from templates/fast_maps.template

package model3d

// CoordMap implements a map-like interface for
// mapping Coord3D to interface{}.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type CoordMap struct {
	slowMap map[Coord3D]interface{}
	fastMap map[uint64]cellForCoordMap
}

// NewCoordMap creates an empty map.
func NewCoordMap() *CoordMap {
	return &CoordMap{fastMap: map[uint64]cellForCoordMap{}}
}

// Len gets the number of elements in the map.
func (m *CoordMap) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Value is like Load(), but without a second return
// value.
func (m *CoordMap) Value(key Coord3D) interface{} {
	res, _ := m.Load(key)
	return res
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *CoordMap) Load(key Coord3D) (interface{}, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[hashForCoordMap(key)]
		if !ok || cell.Key != key {
			return nil, false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *CoordMap) Delete(key Coord3D) {
	if m.fastMap != nil {
		hash := hashForCoordMap(key)
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *CoordMap) Store(key Coord3D, value interface{}) {
	if m.fastMap != nil {
		hash := hashForCoordMap(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForCoordMap{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// KeyRange is like Range, but only iterates over
// keys, not values.
func (m *CoordMap) KeyRange(f func(key Coord3D) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key) {
				return
			}
		}
	} else {
		for k := range m.slowMap {
			if !f(k) {
				return
			}
		}
	}
}

// ValueRange is like Range, but only iterates over
// values only.
func (m *CoordMap) ValueRange(f func(value interface{}) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Value) {
				return
			}
		}
	} else {
		for _, v := range m.slowMap {
			if !f(v) {
				return
			}
		}
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *CoordMap) Range(f func(key Coord3D, value interface{}) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key, cell.Value) {
				return
			}
		}
	} else {
		for k, v := range m.slowMap {
			if !f(k, v) {
				return
			}
		}
	}
}

func (m *CoordMap) fastToSlow() {
	m.slowMap = map[Coord3D]interface{}{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForCoordMap struct {
	Key   Coord3D
	Value interface{}
}

func hashForCoordMap(c Coord3D) uint64 {
	return c.fastHash64()
}

// CoordToFaces implements a map-like interface for
// mapping Coord3D to []*Triangle.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type CoordToFaces struct {
	slowMap map[Coord3D][]*Triangle
	fastMap map[uint64]cellForCoordToFaces
}

// NewCoordToFaces creates an empty map.
func NewCoordToFaces() *CoordToFaces {
	return &CoordToFaces{fastMap: map[uint64]cellForCoordToFaces{}}
}

// Len gets the number of elements in the map.
func (m *CoordToFaces) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Value is like Load(), but without a second return
// value.
func (m *CoordToFaces) Value(key Coord3D) []*Triangle {
	res, _ := m.Load(key)
	return res
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *CoordToFaces) Load(key Coord3D) ([]*Triangle, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[hashForCoordToFaces(key)]
		if !ok || cell.Key != key {
			return nil, false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *CoordToFaces) Delete(key Coord3D) {
	if m.fastMap != nil {
		hash := hashForCoordToFaces(key)
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *CoordToFaces) Store(key Coord3D, value []*Triangle) {
	if m.fastMap != nil {
		hash := hashForCoordToFaces(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForCoordToFaces{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// Append appends x to the value stored for the given key
// and returns the new value.
func (m *CoordToFaces) Append(key Coord3D, x *Triangle) []*Triangle {
	if m.fastMap != nil {
		hash := hashForCoordToFaces(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			return m.Append(key, x)
		} else {
			value := append(cell.Value, x)
			m.fastMap[hash] = cellForCoordToFaces{Key: key, Value: value}
			return value
		}
	} else {
		value := append(m.slowMap[key], x)
		m.slowMap[key] = value
		return value
	}
}

// KeyRange is like Range, but only iterates over
// keys, not values.
func (m *CoordToFaces) KeyRange(f func(key Coord3D) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key) {
				return
			}
		}
	} else {
		for k := range m.slowMap {
			if !f(k) {
				return
			}
		}
	}
}

// ValueRange is like Range, but only iterates over
// values only.
func (m *CoordToFaces) ValueRange(f func(value []*Triangle) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Value) {
				return
			}
		}
	} else {
		for _, v := range m.slowMap {
			if !f(v) {
				return
			}
		}
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *CoordToFaces) Range(f func(key Coord3D, value []*Triangle) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key, cell.Value) {
				return
			}
		}
	} else {
		for k, v := range m.slowMap {
			if !f(k, v) {
				return
			}
		}
	}
}

func (m *CoordToFaces) fastToSlow() {
	m.slowMap = map[Coord3D][]*Triangle{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForCoordToFaces struct {
	Key   Coord3D
	Value []*Triangle
}

func hashForCoordToFaces(c Coord3D) uint64 {
	return c.fastHash64()
}

// CoordToCoord implements a map-like interface for
// mapping Coord3D to Coord3D.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type CoordToCoord struct {
	slowMap map[Coord3D]Coord3D
	fastMap map[uint64]cellForCoordToCoord
}

// NewCoordToCoord creates an empty map.
func NewCoordToCoord() *CoordToCoord {
	return &CoordToCoord{fastMap: map[uint64]cellForCoordToCoord{}}
}

// Len gets the number of elements in the map.
func (m *CoordToCoord) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Value is like Load(), but without a second return
// value.
func (m *CoordToCoord) Value(key Coord3D) Coord3D {
	res, _ := m.Load(key)
	return res
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *CoordToCoord) Load(key Coord3D) (Coord3D, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[hashForCoordToCoord(key)]
		if !ok || cell.Key != key {
			return Coord3D{}, false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *CoordToCoord) Delete(key Coord3D) {
	if m.fastMap != nil {
		hash := hashForCoordToCoord(key)
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *CoordToCoord) Store(key Coord3D, value Coord3D) {
	if m.fastMap != nil {
		hash := hashForCoordToCoord(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForCoordToCoord{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// KeyRange is like Range, but only iterates over
// keys, not values.
func (m *CoordToCoord) KeyRange(f func(key Coord3D) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key) {
				return
			}
		}
	} else {
		for k := range m.slowMap {
			if !f(k) {
				return
			}
		}
	}
}

// ValueRange is like Range, but only iterates over
// values only.
func (m *CoordToCoord) ValueRange(f func(value Coord3D) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Value) {
				return
			}
		}
	} else {
		for _, v := range m.slowMap {
			if !f(v) {
				return
			}
		}
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *CoordToCoord) Range(f func(key Coord3D, value Coord3D) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key, cell.Value) {
				return
			}
		}
	} else {
		for k, v := range m.slowMap {
			if !f(k, v) {
				return
			}
		}
	}
}

func (m *CoordToCoord) fastToSlow() {
	m.slowMap = map[Coord3D]Coord3D{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForCoordToCoord struct {
	Key   Coord3D
	Value Coord3D
}

func hashForCoordToCoord(c Coord3D) uint64 {
	return c.fastHash64()
}

// CoordToInt implements a map-like interface for
// mapping Coord3D to int.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type CoordToInt struct {
	slowMap map[Coord3D]int
	fastMap map[uint64]cellForCoordToInt
}

// NewCoordToInt creates an empty map.
func NewCoordToInt() *CoordToInt {
	return &CoordToInt{fastMap: map[uint64]cellForCoordToInt{}}
}

// Len gets the number of elements in the map.
func (m *CoordToInt) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Value is like Load(), but without a second return
// value.
func (m *CoordToInt) Value(key Coord3D) int {
	res, _ := m.Load(key)
	return res
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *CoordToInt) Load(key Coord3D) (int, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[hashForCoordToInt(key)]
		if !ok || cell.Key != key {
			return 0, false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *CoordToInt) Delete(key Coord3D) {
	if m.fastMap != nil {
		hash := hashForCoordToInt(key)
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *CoordToInt) Store(key Coord3D, value int) {
	if m.fastMap != nil {
		hash := hashForCoordToInt(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForCoordToInt{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// Add adds x to the value stored for the given key and
// returns the new value.
func (m *CoordToInt) Add(key Coord3D, x int) int {
	if m.fastMap != nil {
		hash := hashForCoordToInt(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			return m.Add(key, x)
		} else {
			m.fastMap[hash] = cellForCoordToInt{Key: key, Value: cell.Value + x}
			return cell.Value + x
		}
	} else {
		value := m.slowMap[key] + x
		m.slowMap[key] = value
		return value
	}
}

// KeyRange is like Range, but only iterates over
// keys, not values.
func (m *CoordToInt) KeyRange(f func(key Coord3D) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key) {
				return
			}
		}
	} else {
		for k := range m.slowMap {
			if !f(k) {
				return
			}
		}
	}
}

// ValueRange is like Range, but only iterates over
// values only.
func (m *CoordToInt) ValueRange(f func(value int) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Value) {
				return
			}
		}
	} else {
		for _, v := range m.slowMap {
			if !f(v) {
				return
			}
		}
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *CoordToInt) Range(f func(key Coord3D, value int) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key, cell.Value) {
				return
			}
		}
	} else {
		for k, v := range m.slowMap {
			if !f(k, v) {
				return
			}
		}
	}
}

func (m *CoordToInt) fastToSlow() {
	m.slowMap = map[Coord3D]int{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForCoordToInt struct {
	Key   Coord3D
	Value int
}

func hashForCoordToInt(c Coord3D) uint64 {
	return c.fastHash64()
}

// EdgeMap implements a map-like interface for
// mapping [2]Coord3D to interface{}.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type EdgeMap struct {
	slowMap map[[2]Coord3D]interface{}
	fastMap map[uint64]cellForEdgeMap
}

// NewEdgeMap creates an empty map.
func NewEdgeMap() *EdgeMap {
	return &EdgeMap{fastMap: map[uint64]cellForEdgeMap{}}
}

// Len gets the number of elements in the map.
func (m *EdgeMap) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Value is like Load(), but without a second return
// value.
func (m *EdgeMap) Value(key [2]Coord3D) interface{} {
	res, _ := m.Load(key)
	return res
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *EdgeMap) Load(key [2]Coord3D) (interface{}, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[hashForEdgeMap(key)]
		if !ok || cell.Key != key {
			return nil, false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *EdgeMap) Delete(key [2]Coord3D) {
	if m.fastMap != nil {
		hash := hashForEdgeMap(key)
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *EdgeMap) Store(key [2]Coord3D, value interface{}) {
	if m.fastMap != nil {
		hash := hashForEdgeMap(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForEdgeMap{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// KeyRange is like Range, but only iterates over
// keys, not values.
func (m *EdgeMap) KeyRange(f func(key [2]Coord3D) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key) {
				return
			}
		}
	} else {
		for k := range m.slowMap {
			if !f(k) {
				return
			}
		}
	}
}

// ValueRange is like Range, but only iterates over
// values only.
func (m *EdgeMap) ValueRange(f func(value interface{}) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Value) {
				return
			}
		}
	} else {
		for _, v := range m.slowMap {
			if !f(v) {
				return
			}
		}
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *EdgeMap) Range(f func(key [2]Coord3D, value interface{}) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key, cell.Value) {
				return
			}
		}
	} else {
		for k, v := range m.slowMap {
			if !f(k, v) {
				return
			}
		}
	}
}

func (m *EdgeMap) fastToSlow() {
	m.slowMap = map[[2]Coord3D]interface{}{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForEdgeMap struct {
	Key   [2]Coord3D
	Value interface{}
}

func hashForEdgeMap(c [2]Coord3D) uint64 {
	h1 := c[0].fastHash()
	h2 := c[1].fastHash()
	return uint64(h1) | (uint64(h2) << 32)
}

// EdgeToBool implements a map-like interface for
// mapping [2]Coord3D to bool.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type EdgeToBool struct {
	slowMap map[[2]Coord3D]bool
	fastMap map[uint64]cellForEdgeToBool
}

// NewEdgeToBool creates an empty map.
func NewEdgeToBool() *EdgeToBool {
	return &EdgeToBool{fastMap: map[uint64]cellForEdgeToBool{}}
}

// Len gets the number of elements in the map.
func (m *EdgeToBool) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Value is like Load(), but without a second return
// value.
func (m *EdgeToBool) Value(key [2]Coord3D) bool {
	res, _ := m.Load(key)
	return res
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *EdgeToBool) Load(key [2]Coord3D) (bool, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[hashForEdgeToBool(key)]
		if !ok || cell.Key != key {
			return false, false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *EdgeToBool) Delete(key [2]Coord3D) {
	if m.fastMap != nil {
		hash := hashForEdgeToBool(key)
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *EdgeToBool) Store(key [2]Coord3D, value bool) {
	if m.fastMap != nil {
		hash := hashForEdgeToBool(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForEdgeToBool{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// KeyRange is like Range, but only iterates over
// keys, not values.
func (m *EdgeToBool) KeyRange(f func(key [2]Coord3D) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key) {
				return
			}
		}
	} else {
		for k := range m.slowMap {
			if !f(k) {
				return
			}
		}
	}
}

// ValueRange is like Range, but only iterates over
// values only.
func (m *EdgeToBool) ValueRange(f func(value bool) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Value) {
				return
			}
		}
	} else {
		for _, v := range m.slowMap {
			if !f(v) {
				return
			}
		}
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *EdgeToBool) Range(f func(key [2]Coord3D, value bool) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key, cell.Value) {
				return
			}
		}
	} else {
		for k, v := range m.slowMap {
			if !f(k, v) {
				return
			}
		}
	}
}

func (m *EdgeToBool) fastToSlow() {
	m.slowMap = map[[2]Coord3D]bool{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForEdgeToBool struct {
	Key   [2]Coord3D
	Value bool
}

func hashForEdgeToBool(c [2]Coord3D) uint64 {
	h1 := c[0].fastHash()
	h2 := c[1].fastHash()
	return uint64(h1) | (uint64(h2) << 32)
}

// EdgeToInt implements a map-like interface for
// mapping [2]Coord3D to int.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type EdgeToInt struct {
	slowMap map[[2]Coord3D]int
	fastMap map[uint64]cellForEdgeToInt
}

// NewEdgeToInt creates an empty map.
func NewEdgeToInt() *EdgeToInt {
	return &EdgeToInt{fastMap: map[uint64]cellForEdgeToInt{}}
}

// Len gets the number of elements in the map.
func (m *EdgeToInt) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Value is like Load(), but without a second return
// value.
func (m *EdgeToInt) Value(key [2]Coord3D) int {
	res, _ := m.Load(key)
	return res
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *EdgeToInt) Load(key [2]Coord3D) (int, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[hashForEdgeToInt(key)]
		if !ok || cell.Key != key {
			return 0, false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *EdgeToInt) Delete(key [2]Coord3D) {
	if m.fastMap != nil {
		hash := hashForEdgeToInt(key)
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *EdgeToInt) Store(key [2]Coord3D, value int) {
	if m.fastMap != nil {
		hash := hashForEdgeToInt(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForEdgeToInt{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// Add adds x to the value stored for the given key and
// returns the new value.
func (m *EdgeToInt) Add(key [2]Coord3D, x int) int {
	if m.fastMap != nil {
		hash := hashForEdgeToInt(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			return m.Add(key, x)
		} else {
			m.fastMap[hash] = cellForEdgeToInt{Key: key, Value: cell.Value + x}
			return cell.Value + x
		}
	} else {
		value := m.slowMap[key] + x
		m.slowMap[key] = value
		return value
	}
}

// KeyRange is like Range, but only iterates over
// keys, not values.
func (m *EdgeToInt) KeyRange(f func(key [2]Coord3D) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key) {
				return
			}
		}
	} else {
		for k := range m.slowMap {
			if !f(k) {
				return
			}
		}
	}
}

// ValueRange is like Range, but only iterates over
// values only.
func (m *EdgeToInt) ValueRange(f func(value int) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Value) {
				return
			}
		}
	} else {
		for _, v := range m.slowMap {
			if !f(v) {
				return
			}
		}
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *EdgeToInt) Range(f func(key [2]Coord3D, value int) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key, cell.Value) {
				return
			}
		}
	} else {
		for k, v := range m.slowMap {
			if !f(k, v) {
				return
			}
		}
	}
}

func (m *EdgeToInt) fastToSlow() {
	m.slowMap = map[[2]Coord3D]int{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForEdgeToInt struct {
	Key   [2]Coord3D
	Value int
}

func hashForEdgeToInt(c [2]Coord3D) uint64 {
	h1 := c[0].fastHash()
	h2 := c[1].fastHash()
	return uint64(h1) | (uint64(h2) << 32)
}

// EdgeToFaces implements a map-like interface for
// mapping [2]Coord3D to []*Triangle.
//
// This can be more efficient than using a map directly,
// since it uses a special hash function for coordinates.
// The speed-up is variable, but was ~2x as of mid-2021.
type EdgeToFaces struct {
	slowMap map[[2]Coord3D][]*Triangle
	fastMap map[uint64]cellForEdgeToFaces
}

// NewEdgeToFaces creates an empty map.
func NewEdgeToFaces() *EdgeToFaces {
	return &EdgeToFaces{fastMap: map[uint64]cellForEdgeToFaces{}}
}

// Len gets the number of elements in the map.
func (m *EdgeToFaces) Len() int {
	if m.fastMap != nil {
		return len(m.fastMap)
	} else {
		return len(m.slowMap)
	}
}

// Value is like Load(), but without a second return
// value.
func (m *EdgeToFaces) Value(key [2]Coord3D) []*Triangle {
	res, _ := m.Load(key)
	return res
}

// Load gets the value for the given key.
//
// If no value is present, the first return argument is a
// zero value, and the second is false. Otherwise, the
// second return value is true.
func (m *EdgeToFaces) Load(key [2]Coord3D) ([]*Triangle, bool) {
	if m.fastMap != nil {
		cell, ok := m.fastMap[hashForEdgeToFaces(key)]
		if !ok || cell.Key != key {
			return nil, false
		}
		return cell.Value, true
	} else {
		x, y := m.slowMap[key]
		return x, y
	}
}

// Delete removes the key from the map if it exists, and
// does nothing otherwise.
func (m *EdgeToFaces) Delete(key [2]Coord3D) {
	if m.fastMap != nil {
		hash := hashForEdgeToFaces(key)
		if cell, ok := m.fastMap[hash]; ok && cell.Key == key {
			delete(m.fastMap, hash)
		}
	} else {
		delete(m.slowMap, key)
	}
}

// Store assigns the value to the given key, overwriting
// the previous value for the key if necessary.
func (m *EdgeToFaces) Store(key [2]Coord3D, value []*Triangle) {
	if m.fastMap != nil {
		hash := hashForEdgeToFaces(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			m.slowMap[key] = value
		} else {
			m.fastMap[hash] = cellForEdgeToFaces{Key: key, Value: value}
		}
	} else {
		m.slowMap[key] = value
	}
}

// Append appends x to the value stored for the given key
// and returns the new value.
func (m *EdgeToFaces) Append(key [2]Coord3D, x *Triangle) []*Triangle {
	if m.fastMap != nil {
		hash := hashForEdgeToFaces(key)
		cell, ok := m.fastMap[hash]
		if ok && cell.Key != key {
			// We must switch to a slow map to store colliding values.
			m.fastToSlow()
			return m.Append(key, x)
		} else {
			value := append(cell.Value, x)
			m.fastMap[hash] = cellForEdgeToFaces{Key: key, Value: value}
			return value
		}
	} else {
		value := append(m.slowMap[key], x)
		m.slowMap[key] = value
		return value
	}
}

// KeyRange is like Range, but only iterates over
// keys, not values.
func (m *EdgeToFaces) KeyRange(f func(key [2]Coord3D) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key) {
				return
			}
		}
	} else {
		for k := range m.slowMap {
			if !f(k) {
				return
			}
		}
	}
}

// ValueRange is like Range, but only iterates over
// values only.
func (m *EdgeToFaces) ValueRange(f func(value []*Triangle) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Value) {
				return
			}
		}
	} else {
		for _, v := range m.slowMap {
			if !f(v) {
				return
			}
		}
	}
}

// Range iterates over the map, calling f successively for
// each value until it returns false, or all entries are
// enumerated.
//
// It is not safe to modify the map with Store or Delete
// during enumeration.
func (m *EdgeToFaces) Range(f func(key [2]Coord3D, value []*Triangle) bool) {
	if m.fastMap != nil {
		for _, cell := range m.fastMap {
			if !f(cell.Key, cell.Value) {
				return
			}
		}
	} else {
		for k, v := range m.slowMap {
			if !f(k, v) {
				return
			}
		}
	}
}

func (m *EdgeToFaces) fastToSlow() {
	m.slowMap = map[[2]Coord3D][]*Triangle{}
	for _, cell := range m.fastMap {
		m.slowMap[cell.Key] = cell.Value
	}
	m.fastMap = nil
}

type cellForEdgeToFaces struct {
	Key   [2]Coord3D
	Value []*Triangle
}

func hashForEdgeToFaces(c [2]Coord3D) uint64 {
	h1 := c[0].fastHash()
	h2 := c[1].fastHash()
	return uint64(h1) | (uint64(h2) << 32)
}
