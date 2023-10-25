package skeleton

import (
	"sort"

	"github.com/xackery/engine/core"
	"github.com/xackery/engine/graphic"
	"github.com/xackery/engine/math32"
	"github.com/xackery/quail/common"
)

type jointEntry struct {
	index int
	node  *core.Node
	bone  common.Bone
}

var (
	joints []*jointEntry
)

func Generate(in []common.Bone) (*graphic.Skeleton, error) {
	skel := graphic.NewSkeleton()

	joints = make([]*jointEntry, 0)

	rootBone := core.NewNode()
	rootBone.SetName("root")
	traverse(in, rootBone, in[0])

	sort.Slice(joints, func(i, j int) bool {
		return joints[i].index < joints[j].index
	})

	for _, joint := range joints {
		tm := math32.NewMatrix4()
		var pivot math32.Vector3
		pivot.X = joint.bone.Pivot.X
		pivot.Y = joint.bone.Pivot.Y
		pivot.Z = joint.bone.Pivot.Z
		var rotation math32.Quaternion
		rotation.X = joint.bone.Rotation.X
		rotation.Y = joint.bone.Rotation.Y
		rotation.Z = joint.bone.Rotation.Z
		rotation.W = joint.bone.Rotation.W
		var scale math32.Vector3
		scale.X = joint.bone.Scale.X
		scale.Y = joint.bone.Scale.Y
		scale.Z = joint.bone.Scale.Z
		tm.Compose(&pivot, &rotation, &scale)

		ibm := math32.NewMatrix4()
		ibm.GetInverse(tm)

		skel.AddBone(joint.node, ibm)
	}

	return skel, nil
}

func traverse(bones []common.Bone, ptr *core.Node, focus common.Bone) {
	if focus.ChildrenCount > 0 {
		child := bones[focus.ChildIndex]
		childNode := core.NewNode()
		childNode.SetName(child.Name)

		ptr.Add(childNode)
		je := &jointEntry{
			index: int(focus.ChildIndex),
			node:  childNode,
			bone:  child,
		}
		joints = append(joints, je)
		traverse(bones, childNode, child)
	}

	if focus.Next > -1 {
		next := bones[focus.Next]
		nextNode := core.NewNode()
		nextNode.SetName(next.Name)
		ptr.Add(nextNode)
		je := &jointEntry{
			index: int(focus.Next),
			node:  nextNode,
			bone:  next,
		}
		joints = append(joints, je)
		traverse(bones, nextNode, next)
	}
}
