package anim

import (
	"github.com/xackery/engine/animation"
	"github.com/xackery/engine/graphic"
	"github.com/xackery/engine/math32"
	"github.com/xackery/quail/common"
)

func Generate(in []*common.Animation, meshes []*graphic.RiggedMesh) ([]*animation.Animation, error) {
	anims := make([]*animation.Animation, 0)

	for _, entry := range in {
		for _, boneAnim := range entry.Bones {
			for _, mesh := range meshes {

				anim := animation.NewAnimation()
				anim.SetName(entry.Name)
				anim.SetLoop(true)
				var keyframes math32.ArrayF32
				var posValues math32.ArrayF32
				var rotValues math32.ArrayF32
				var scaleValues math32.ArrayF32

				for i, keyframe := range boneAnim.Frames {
					keyframes = append(keyframes, float32(i))
					posValues = append(posValues, keyframe.Translation.X, keyframe.Translation.Y, keyframe.Translation.Z)
					rotValues = append(rotValues, keyframe.Rotation.X, keyframe.Rotation.Y, keyframe.Rotation.Z, keyframe.Rotation.W)
					scaleValues = append(scaleValues, keyframe.Scale.X, keyframe.Scale.Y, keyframe.Scale.Z)
				}
				posChan := animation.NewPositionChannel(mesh)
				posChan.SetBuffers(keyframes, posValues)
				anim.AddChannel(posChan)

				rotChan := animation.NewRotationChannel(mesh)
				rotChan.SetBuffers(keyframes, rotValues)
				anim.AddChannel(rotChan)

				scaleChan := animation.NewScaleChannel(mesh)
				scaleChan.SetBuffers(keyframes, scaleValues)
				anim.AddChannel(scaleChan)

				anims = append(anims, anim)
			}
		}
	}
	return anims, nil
}
