package indexes

import (
	"reflect"
	"sort"
	"testing"
)

func TestAdd(t *testing.T) {
	cases := []struct {
		strings []string
		IDs     []int64
	}{
		{
			[]string{"abcd"},
			[]int64{1},
		},
		{
			[]string{"abcd", "abc"},
			[]int64{1, 2},
		},
		{
			[]string{"abcd", "cba", "你好"},
			[]int64{1, 2, 3},
		},
	}

	for _, c := range cases {
		root := NewTrieNode()
		for i, s := range c.strings {
			root.Add(s, c.IDs[i])
		}

		for i, s := range c.strings {
			curNode := root
			for _, char := range []rune(s) {
				if curNode.children[char] == nil {
					t.Errorf("error adding key: %#U", char)
					break
				}
				curNode = curNode.children[char]
			}
			if !curNode.values.has(c.IDs[i]) {
				t.Errorf("error adding ID: %d", c.IDs[i])
			}
		}

		if root.Len() != len(c.strings) {
			t.Errorf("wrong no. of entires added, expecting %d, but actually got %d", len(c.strings), root.Len())
		}
	}

}

func TestRemove(t *testing.T) {
	cases := []struct {
		strings   []string
		IDs       []int64
		isRemoved []bool
	}{
		{
			[]string{"abcd"},
			[]int64{1},
			[]bool{true},
		},
		{
			[]string{"abcd", "abc"},
			[]int64{1, 2},
			[]bool{false, true},
		},
		{
			[]string{"abcd", "cba", "你好"},
			[]int64{1, 2, 3},
			[]bool{true, false, true},
		},
	}

	for _, c := range cases {
		root := NewTrieNode()
		for i, s := range c.strings {
			root.Add(s, c.IDs[i])
		}

		for i, b := range c.isRemoved {
			if b {
				root.Remove(c.strings[i], c.IDs[i])
			}
		}

		for i, s := range c.strings {
			curNode := root
			if c.isRemoved[i] {
				for _, char := range []rune(s) {
					if curNode == nil {
						break
					}

					if len(curNode.children) == 0 && curNode != root && len(curNode.values) == 0 {
						t.Errorf("error removing empty TrieNode")
						break
					}
					curNode = curNode.children[char]
				}

				if curNode != nil && curNode.values.has(c.IDs[i]) {
					t.Errorf("removal failed for %s", s)
				}
			} else {
				for _, char := range []rune(s) {
					if curNode.children[char] == nil {
						t.Errorf("unexpected removal for key: %#U", char)
						break
					}
					curNode = curNode.children[char]
				}
				if !curNode.values.has(c.IDs[i]) {
					t.Errorf("unexpected removal for ID: %d", c.IDs[i])
				}
			}
		}
	}
}

func TestFind(t *testing.T) {
	cases := []struct {
		strings  []string
		IDs      []int64
		prefix   string
		max      int
		expected []int64
	}{
		{
			[]string{"abc", "a", "b", "abcd"},
			[]int64{1, 2, 3, 4},
			"a",
			3,
			[]int64{1, 2, 4},
		},
		{
			[]string{"abc", "a", "b", "abcd"},
			[]int64{1, 2, 3, 4},
			"a",
			1,
			[]int64{2},
		},
		{
			[]string{"你好", "你不好", "好", "a"},
			[]int64{1, 2, 3, 4},
			"你",
			10,
			[]int64{1, 2},
		},
	}

	for _, c := range cases {
		root := NewTrieNode()
		for i, s := range c.strings {
			root.Add(s, c.IDs[i])
		}

		res := root.Find(c.prefix, c.max)
		sort.Slice(res, func(i, j int) bool {
			return res[i] < res[j]
		})

		if !reflect.DeepEqual(res, c.expected) {
			t.Errorf("Incorrect find result")
		}
	}
}
