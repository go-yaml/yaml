package yaml

// IndexMap0
type IndexMap0 struct {
	Tag   string
	Value interface{}
	Index int
}

func (im *IndexMap0) SetYAML(tag string, value interface{}) bool {
	im.Tag = tag
	im.Value = value
	im.Index += 1
	return true
}

func (im *IndexMap0) Reset() {
	im.Index = 0
}

// IndexMap1
type IndexMap1 struct {
	Tag   string
	Value interface{}
	Index int
}

func (im *IndexMap1) SetYAML(tag string, value interface{}) bool {
	im.Tag = tag
	im.Value = value
	im.Index += 1
	return true
}

func (im *IndexMap1) Reset() {
	im.Index = 0
}
