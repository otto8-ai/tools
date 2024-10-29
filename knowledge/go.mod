module github.com/gptscript-ai/knowledge

go 1.23.1

toolchain go1.23.2

replace (
	github.com/hupe1980/golc => github.com/iwilltry42/golc v0.0.113-0.20240802113826-d065a3c5b0c7 // nbformat extension
	github.com/ledongthuc/pdf => github.com/iwilltry42/pdf v0.0.0-20240517145113-99fbaebc5dd3 // fix for reading some PDFs: https://github.com/ledongthuc/pdf/pull/36 + https://github.com/iwilltry42/pdf/pull/2
	github.com/philippgille/chromem-go => github.com/iwilltry42/chromem-go v0.0.0-20241025145357-75b37396158f
	github.com/tmc/langchaingo => github.com/StrongMonkey/langchaingo v0.0.0-20240617180437-9af4bee04c8b // Context-Aware Markdown Splitting
)

require (
	code.sajari.com/docconv/v2 v2.0.0-pre.4
	dario.cat/mergo v1.0.0
	github.com/JohannesKaufmann/html-to-markdown v1.6.0
	github.com/PuerkitoBio/goquery v1.9.2
	github.com/acorn-io/cmd v0.0.0-20240728043647-5ee22f336de6
	github.com/acorn-io/z v0.0.0-20231104012607-4cab1b3ec5e5
	github.com/adrg/xdg v0.4.0
	github.com/asg017/sqlite-vec-go-bindings v0.1.4-alpha.2
	github.com/cohere-ai/cohere-go/v2 v2.8.2
	github.com/gabriel-vasile/mimetype v1.4.4
	github.com/gen2brain/go-fitz v1.24.14
	github.com/gin-gonic/gin v1.10.0
	github.com/glebarez/sqlite v1.11.0
	github.com/go-git/go-git/v5 v5.12.0
	github.com/google/uuid v1.6.0
	github.com/gptscript-ai/go-gptscript v0.9.6-0.20241028114905-1fce3cceae35
	github.com/hupe1980/golc v0.0.112
	github.com/iwilltry42/bm25-go v0.0.0-20240909111832-a928590cc9da
	github.com/jackc/pgx/v5 v5.7.1
	github.com/jmcarbo/stopwords v1.1.9
	github.com/joho/godotenv v1.5.1
	github.com/knadh/koanf/maps v0.1.1
	github.com/knadh/koanf/parsers/json v0.1.0
	github.com/knadh/koanf/parsers/yaml v0.1.0
	github.com/knadh/koanf/providers/env v0.1.0
	github.com/knadh/koanf/providers/rawbytes v0.1.0
	github.com/knadh/koanf/v2 v2.1.1
	github.com/ledongthuc/pdf v0.0.0-20240201131950-da5b75280b06
	github.com/lu4p/cat v0.1.5
	github.com/mitchellh/mapstructure v1.5.0
	github.com/ncruces/go-sqlite3 v0.19.0
	github.com/pgvector/pgvector-go v0.2.2
	github.com/philippgille/chromem-go v0.6.1-0.20240811154507-a1944285b284
	github.com/spf13/cobra v1.8.1
	github.com/stretchr/testify v1.9.0
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/swag v1.16.3
	github.com/tmc/langchaingo v0.1.12
	golang.org/x/sync v0.8.0
	gorm.io/driver/postgres v1.5.9
	gorm.io/gorm v1.25.12
	sigs.k8s.io/yaml v1.4.0
)

require (
	cloud.google.com/go/ai v0.7.0 // indirect
	github.com/AssemblyAI/assemblyai-go-sdk v1.3.0 // indirect
	github.com/EndFirstCorp/peekingReader v0.0.0-20171012052444-257fb6f1a1a6 // indirect
	github.com/JalfResi/justext v0.0.0-20170829062021-c0282dea7198 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.0.0 // indirect
	github.com/advancedlogic/GoOse v0.0.0-20191112112754-e742535969c1 // indirect
	github.com/andybalholm/cascadia v1.3.2 // indirect
	github.com/araddon/dateparse v0.0.0-20200409225146-d820a6159ab1 // indirect
	github.com/avast/retry-go v3.0.0+incompatible // indirect
	github.com/aws/aws-sdk-go-v2 v1.27.2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/bedrockruntime v1.9.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/textract v1.30.11 // indirect
	github.com/aws/smithy-go v1.20.2 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/bytedance/sonic v1.11.6 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cloudflare/circl v1.3.9 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/cohere-ai/tokenizer v1.1.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/cyphar/filepath-securejoin v0.2.5 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dlclark/regexp2 v1.11.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/ebitengine/purego v0.8.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fatih/set v0.2.1 // indirect
	github.com/getkin/kin-openapi v0.128.0 // indirect
	github.com/gigawattio/window v0.0.0-20180317192513-0f5467e35573 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/glebarez/go-sqlite v1.21.2 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.5.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.20.0 // indirect
	github.com/go-resty/resty/v2 v2.3.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1 // indirect
	github.com/gobwas/ws v1.2.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/pprof v0.0.0-20230926050212-f7f687d19a98 // indirect
	github.com/googleapis/gax-go/v2 v2.12.4 // indirect
	github.com/gorilla/css v1.0.0 // indirect
	github.com/hupe1980/go-promptlayer v0.0.6 // indirect
	github.com/hupe1980/go-textractor v0.0.9 // indirect
	github.com/hupe1980/go-tiktoken v0.0.9 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/invopop/yaml v0.3.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jaytaylor/html2text v0.0.0-20230321000545-74c2419ad056 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jupiterrider/ffi v0.2.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/levigross/exp-html v0.0.0-20120902181939-8df60c69a8f5 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/microcosm-cc/bluemonday v1.0.26 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/ncruces/julianday v1.0.0 // indirect
	github.com/olekukonko/tablewriter v0.0.6-0.20230925090304-df64c4bbad77 // indirect
	github.com/otiai10/gosseract/v2 v2.2.4 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkoukk/tiktoken-go v0.1.6 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/richardlehane/mscfb v1.0.3 // indirect
	github.com/richardlehane/msoleps v1.0.3 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sashabaranov/go-openai v1.26.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/serpapi/google-search-results-golang v0.0.0-20240325113416-ec93f510648e // indirect
	github.com/skeema/knownhosts v1.2.2 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/tetratelabs/wazero v1.8.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/unidoc/unioffice v1.33.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	gitlab.com/golang-commonmark/html v0.0.0-20191124015941-a22733972181 // indirect
	gitlab.com/golang-commonmark/linkify v0.0.0-20191026162114-a0c2df6c8f82 // indirect
	gitlab.com/golang-commonmark/markdown v0.0.0-20211110145824-bf3e522c626a // indirect
	gitlab.com/golang-commonmark/mdurl v0.0.0-20191124015652-932350d1cb84 // indirect
	gitlab.com/golang-commonmark/puny v0.0.0-20191124015043-9f83538fa04f // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/exp v0.0.0-20240613232115-7f521ea00fb8 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/tools v0.23.0 // indirect
	google.golang.org/api v0.184.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240610135401-a8a62080eff3 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240610135401-a8a62080eff3 // indirect
	google.golang.org/grpc v1.64.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	modernc.org/libc v1.22.5 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/sqlite v1.23.1 // indirect
	nhooyr.io/websocket v1.8.7 // indirect
)
