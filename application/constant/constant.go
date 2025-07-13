package constant

type FileType string

type TemplateTypePath string

const (
	FileTypeTxt FileType = "txt"
	FileTypeCsv FileType = "csv"
)

const (
	KidVn   TemplateTypePath = "kidvn"
	AdultVn TemplateTypePath = "adultvn"
	KidUs   TemplateTypePath = "kidus"
	AdultUs TemplateTypePath = "adultus"
)
