package model

type ContentForPath struct {
	Path          string
	ContentOpener Opener
}

type Content struct {
	ByPath []ContentForPath
	Closer func() error
}

func (c *Content) Close() error {
	if c.Closer != nil {
		return c.Closer()
	}
	return nil
}
