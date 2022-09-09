package msgo

import "strings"

type treeNode struct {
	name     string
	children []*treeNode
	rootPath string
	isEnd    bool
}

// /user/info/:id
func (t *treeNode) Put(path string) {
	root := t
	elements := strings.Split(path, "/")
	for index, element := range elements {
		if index == 0 {
			continue
		}
		children := t.children
		isMatch := false
		for _, node := range children {
			if node.name == element {
				// continue to match
				// use next element to match
				isMatch = true
				t = node
				break
			}
		}

		if !isMatch {
			isEnd := false
			if len(elements)-1 == index {
				isEnd = true
			}
			node := &treeNode{
				name:     element,
				children: make([]*treeNode, 0),
				isEnd:    isEnd,
			}
			children = append(children, node)
			t.children = children
			t = node
		}
	}

	t = root
}

// get path: /user/get/1
func (t *treeNode) Get(path string) *treeNode {
	elements := strings.Split(path, "/")
	t.rootPath = ""
	for index, element := range elements {
		if index == 0 {
			continue
		}

		children := t.children
		isMatch := false
		for _, node := range children {
			if node.name == element ||
				node.name == "*" ||
				strings.Contains(node.name, ":") {
				isMatch = true
				t.rootPath += "/" + node.name
				node.rootPath = t.rootPath
				if len(elements)-1 == index { // if this the final element, then return current node
					return node
				}
				t = node // if this element is not a final element, please continue searching
				break
			}
		}

		if !isMatch {
			for _, node := range children { // I don't know
				if node.name == "**" {
					return node
				}
			}
		}
	}

	return nil
}
