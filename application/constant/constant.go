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

type AdultSizeAvailable string

// AdultSizeAvailable: XS, S, M, L, XL, 2XL, 3XL
const (
	AdultSizeAvailableXS  AdultSizeAvailable = "xs"
	AdultSizeAvailableS   AdultSizeAvailable = "s"
	AdultSizeAvailableM   AdultSizeAvailable = "m"
	AdultSizeAvailableL   AdultSizeAvailable = "l"
	AdultSizeAvailableXL  AdultSizeAvailable = "xl"
	AdultSizeAvailable2XL AdultSizeAvailable = "2xl"
	AdultSizeAvailable3XL AdultSizeAvailable = "3xl"
)

type KidSizeAvailable string

// KidSizeAvailable: 3T, 4T, 5, 6, 7, 8
const (
	KidSizeAvailable3T KidSizeAvailable = "3t"
	KidSizeAvailable4T KidSizeAvailable = "4t"
	KidSizeAvailable5  KidSizeAvailable = "5"
	KidSizeAvailable6  KidSizeAvailable = "6"
	KidSizeAvailable7  KidSizeAvailable = "7"
	KidSizeAvailable8  KidSizeAvailable = "8"
)

type GenderType string

const (
	GenderTypeWomen GenderType = "women"
	GenderTypeMen   GenderType = "men"
	GenderTypeKids  GenderType = "kids"
	GenderTypeKid   GenderType = "kid"
)
