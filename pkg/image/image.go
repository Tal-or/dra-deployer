package image

import (
	"fmt"

	"github.com/containers/image/v5/docker/reference"
	"k8s.io/klog/v2"
)

type Reference struct {
	Image string
	Tag   string
}

func Parse(image string) (Reference, error) {
	named, err := reference.ParseNamed(image)
	if err != nil {
		return Reference{}, err
	}

	tag := "latest"
	tagged, ok := named.(reference.Tagged)
	if ok {
		tag = tagged.Tag()
	} else {
		klog.V(5).InfoS("Image is not tagged, using latest as tag", "image", image)
	}

	return Reference{
		Image: named.Name(),
		Tag:   tag,
	}, nil
}

func (r Reference) String() string {
	return fmt.Sprintf("%s:%s", r.Image, r.Tag)
}
