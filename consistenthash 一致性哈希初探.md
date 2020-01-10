* 引言

[论文地址](http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.147.1879)

* 问题

如何通过一个关键字在分布式系统中定位一台服务器，进而存储或检索数据，同时能很好地应对服务器故障以及网络分裂。

简单的 哈希后对n取余得到服务器的编号 并不能应对 服务器故障或者由于网络分裂 的问题。这会导致几乎全部缓存失效，唯一的解决办法是将全部数据重新映射

* 一致性哈希

在一致性哈希算法中，服务器也和关键字一样进行哈希。哈希空间足够大（一般取[0,2^32）)，并且被当作一个收尾相接的环（哈希环的由来），对服务器进行哈希就相当于将服务器放置在这个哈希环上。当我们需要查找一个关键字时，将它哈希（就是把它也定位到环上），然后沿着哈希环顺时针移动直到找到下一台服务器，当到达哈希环尾端后仍然找不到服务器时，使用第一台服务器。理论上这样就搞定上面说到的问题啦，但是在实践中，经过哈希后的服务器经常在环上聚集起来，这就会使得第一台服务器的压力大于其它服务器。这可以通过让服务器在环上分布得更均匀来改善。具体通过以下做法来实现：引入虚拟节点的概念，通过replica count（副本数）来控制每台物理服务器对应的虚节点数，当我们要添加一台服务器时，从0 到 replica count - 1 循环，将哈希关键字改为服务器关键字加上虚节点编号（hash(ser_str#1),hash(ser_str#2)...）生成虚节点的位置，将虚节点放置到哈希环上。这样做能有效地将服务器均匀分配到环上。注意到这里所谓的服务器副本是虚拟的节点，所以完全不涉及服务器之间的数据同步问题（简单地讲，就是现在变成了二段映射，先找到虚节点，然后再根据虚节点找到对应的物理机）。

优势：取余需要将几乎全部数据重新映射，一致性哈希只需要将环上对应节点到上一个节点之间的数据重新映射。

* 实现

```
/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package consistenthash provides an implementation of a ring hash.
package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int
	keys     []int // Sorted
	hashMap  map[int]string
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// IsEmpty returns true if there are no items available.
func (m *Map) IsEmpty() bool {
	return len(m.keys) == 0
}

// Add adds some keys to the hash.
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if m.IsEmpty() {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })

	// Means we have cycled back to the first replica.
	if idx == len(m.keys) {
		idx = 0
	}

	return m.hashMap[m.keys[idx]]
}

```


