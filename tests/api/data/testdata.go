package data

import (
	_ "embed"
)

//go:embed txts/test_document.txt
var TestDocumentTXT []byte

//go:embed txts/test_data.json
var TestDataJSON []byte

//go:embed txts/test_config.xml
var TestConfigXML []byte

//go:embed txts/test_script.js
var TestScriptJS []byte

//go:embed txts/test_styles.css
var TestStylesCSS []byte

//go:embed test_image.png
var TestImagePNG []byte

//go:embed test_image.jpg
var TestImageJPG []byte

// TestFile 表示一个测试文件的信息
type TestFile struct {
	Name        string
	Content     []byte
	Format      string
	Category    string
	ContentType string
}

// GetTextFiles 返回所有文本类型的测试文件
func GetTextFiles() []TestFile {
	return []TestFile{
		{Name: "test_document.txt", Content: TestDocumentTXT, Format: "txt", Category: "text", ContentType: "text/plain"},
		{Name: "test_data.json", Content: TestDataJSON, Format: "json", Category: "text", ContentType: "application/json"},
		{Name: "test_config.xml", Content: TestConfigXML, Format: "xml", Category: "text", ContentType: "application/xml"},
		{Name: "test_script.js", Content: TestScriptJS, Format: "js", Category: "text", ContentType: "application/javascript"},
		{Name: "test_styles.css", Content: TestStylesCSS, Format: "css", Category: "text", ContentType: "text/css"},
	}
}

// GetImageFiles 返回所有图片类型的测试文件
func GetImageFiles() []TestFile {
	return []TestFile{
		{Name: "test_image.png", Content: TestImagePNG, Format: "png", Category: "image", ContentType: "image/png"},
		{Name: "test_image.jpg", Content: TestImageJPG, Format: "jpg", Category: "image", ContentType: "image/jpeg"},
	}
}

// GetAllFiles 返回所有测试文件
func GetAllFiles() []TestFile {
	return append(GetTextFiles(), GetImageFiles()...)
}
