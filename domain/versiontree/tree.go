package versiontree

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hanchen1103/iteration_engine/domain/model"
)

func Build(versions []*model.Version) ([]*model.VersionNode, error) {
	nodes := make([]*model.VersionNode, 0, len(versions))
	byID := map[string]*model.VersionNode{}
	for _, version := range versions {
		if version == nil {
			continue
		}
		id := strings.TrimSpace(version.ID)
		if id == "" {
			return nil, fmt.Errorf("version id is required")
		}
		if _, ok := byID[id]; ok {
			return nil, fmt.Errorf("duplicate version id: %s", id)
		}
		node := &model.VersionNode{Version: version}
		nodes = append(nodes, node)
		byID[id] = node
	}

	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].Version.VersionNo < nodes[j].Version.VersionNo
	})

	var roots []*model.VersionNode
	for _, node := range nodes {
		parentID := strings.TrimSpace(node.Version.BaseVersionID)
		if parentID == "" {
			roots = append(roots, node)
			continue
		}
		parent, ok := byID[parentID]
		if !ok {
			return nil, fmt.Errorf("version %s references missing base version %s", node.Version.ID, parentID)
		}
		parent.Children = append(parent.Children, node)
	}

	if err := detectVersionTreeCycle(nodes); err != nil {
		return nil, err
	}
	sortVersionTree(roots)
	return roots, nil
}

func sortVersionTree(nodes []*model.VersionNode) {
	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].Version.VersionNo < nodes[j].Version.VersionNo
	})
	for _, node := range nodes {
		sortVersionTree(node.Children)
	}
}

func detectVersionTreeCycle(nodes []*model.VersionNode) error {
	const (
		unvisited = iota
		visiting
		visited
	)
	state := map[string]int{}

	var visit func(*model.VersionNode) error
	visit = func(node *model.VersionNode) error {
		id := node.Version.ID
		switch state[id] {
		case visiting:
			return fmt.Errorf("version tree contains a cycle at version %s", id)
		case visited:
			return nil
		}
		state[id] = visiting
		for _, child := range node.Children {
			if err := visit(child); err != nil {
				return err
			}
		}
		state[id] = visited
		return nil
	}

	for _, node := range nodes {
		if err := visit(node); err != nil {
			return err
		}
	}
	return nil
}
