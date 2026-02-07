package clipboard

import xclip "golang.design/x/clipboard"

type clipboardImpl struct{}

func (c *clipboardImpl) GetText() (string, error) {
	data := xclip.Read(xclip.FmtText)
	return string(data), nil
}

func (c *clipboardImpl) SetText(text string) error {
	xclip.Write(xclip.FmtText, []byte(text))
	return nil
}
