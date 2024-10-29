package documentloader

import (
	"context"
	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader/pdf/gopdf"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDocumentLoaderConfig_ValidLoader(t *testing.T) {
	cfg, err := GetDocumentLoaderConfig("pdf")
	assert.NoError(t, err)
	assert.IsTypef(t, gopdf.PDFOptions{}, cfg, "cfg is not of type PDFOptions")
}

func TestGetDocumentLoaderConfig_InvalidLoader(t *testing.T) {
	_, err := GetDocumentLoaderConfig("invalid")
	assert.Error(t, err)
}

func TestGetDocumentLoaderFunc_ValidLoaderWithoutConfig(t *testing.T) {
	_, err := GetDocumentLoaderFunc("plaintext", nil)
	assert.NoError(t, err)
}

func TestGetDocumentLoaderFunc_ValidLoaderWithInvalidConfig(t *testing.T) {
	_, err := GetDocumentLoaderFunc("pdf", "invalid")
	assert.Error(t, err)
}

func TestGetDocumentLoaderFunc_ValidLoaderWithValidConfig(t *testing.T) {
	_, err := GetDocumentLoaderFunc("pdf", gopdf.PDFOptions{})
	assert.NoError(t, err)
}

func TestGetDocumentLoaderFunc_InvalidLoader(t *testing.T) {
	_, err := GetDocumentLoaderFunc("invalid", nil)
	assert.Error(t, err)
}

func TestGetDocumentLoaderFunc_LoadPlainText(t *testing.T) {
	loaderFunc, _ := GetDocumentLoaderFunc("plaintext", nil)
	docs, err := loaderFunc(context.Background(), strings.NewReader("test"))
	assert.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "test", docs[0].Content)
}

func TestGetDocumentLoaderFunc_LoadPDF(t *testing.T) {
	loaderFunc, _ := GetDocumentLoaderFunc("pdf", gopdf.PDFOptions{})
	content := `
%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj
3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Resources <<
/Font <<
/F1 <<
/Type /Font
/Subtype /Type1
/BaseFont /Helvetica
>>
>>
/ProcSet [/PDF /Text]
/Contents 4 0 R
>>
endobj
4 0 obj
<<
/Length 46
>>
stream
BT
/F1 18 Tf
100 100 Td
(Hello, this is a fake PDF!) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f
0000000010 00000 n
0000000060 00000 n
0000000115 00000 n
0000000194 00000 n
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
277
%%EOF
`
	_, err := loaderFunc(context.Background(), strings.NewReader(content))
	assert.Error(t, err)
}
