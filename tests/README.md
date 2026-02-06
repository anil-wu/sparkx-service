# Test Data Files

This directory contains various test files for the SparkPlay API testing.

## File Structure

```
tests/
├── api/
│   ├── data/
│   │   ├── txts/           # Text files
│   │   │   ├── test_document.txt
│   │   │   ├── test_data.json
│   │   │   ├── test_config.xml
│   │   │   ├── test_script.js
│   │   │   └── test_styles.css
│   │   ├── test_image.png  # PNG image
│   │   ├── test_image.jpg  # JPEG image
│   │   └── testdata.go     # Go embed code
│   └── file_upload_test.go # Integration tests
└── README.md               # This file
```

## Test Files

| File | Format | Category | Description |
|------|--------|----------|-------------|
| test_document.txt | txt | text | Plain text file with multiline content |
| test_data.json | json | text | JSON configuration data |
| test_config.xml | xml | text | XML configuration file |
| test_script.js | js | text | JavaScript code file |
| test_styles.css | css | text | CSS stylesheet |
| test_image.png | png | image | 1x1 pixel PNG image |
| test_image.jpg | jpg | image | JPEG test image |

## Usage

These files are used by the integration tests in `api/file_upload_test.go` to test:

1. File upload via pre-signed URLs
2. Different file formats and categories
3. File versioning
4. Project file listing

## Binary Files

Binary files (PNG, JPG) are stored as binary data and embedded using Go's `//go:embed` directive.
