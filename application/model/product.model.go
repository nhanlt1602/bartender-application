package model

type Product struct {
	Name               string `json:"name"`
	Code               string `json:"code"`
	Color              string `json:"color"`
	Material           string `json:"material"`
	ManufactureOffice  string `json:"manufacture_office"`
	ManuFactureCompany string `json:"manufacture_company"`
	ManufactureDate    string `json:"manufacture_date"`
	USSize             string `json:"us_size"`
	VNSize             string `json:"vn_size"`
	UKSize             string `json:"uk_size"`
	Gender             string `json:"gender"`
	Attribute          string `json:"attribute"`
	SizeAvailable      string `json:"size_available"`
	QrCode             string `json:"qr_code"`
	RfidBarcode        string `json:"rfid_barcode"`
	Price              string `json:"price"` // Price in string format, e.g., "100.00"
	Currency           string `json:"currency"`
}

type ProductPrinterMsgKafkaRequest struct {
	Template string     `json:"template"`
	Products []*Product `json:"products"`
}

//https://docs.google.com/spreadsheets/d/17jBvS6Gz2wkiFaxOErFhk_eN449Dev3u3e9eQAK3Wuc/edit?gid=926956614#gid=926956614
// Tương ứng với mỗi cột trong Google Sheet sẽ một product_attribute where = sku trên Google Sheet VN
// Đọc file execel sheet VN và Bảng Size để insert dữ liệu product_attribute tướng với mỗi sku trong Google Sheet VN
// Các biến product_attribute sẽ được tạo sẵn trên bom

//------------------------------Sheet VN-------------------------------
//Product Attribute 	: Sheet Column
//Name 					: PRODUCT NAME
//Code 					: STYLE#
//Color 				: COLOR
//Material 				: MATERIAL
//ManufactureOffice 	: Văn phòng: Lầu 3 555 3/2, P.8, Q.10, TP.HCM, Việt Nam
//ManuFactureCompany 	: 130 Ấp Chánh, X. Đức Lập Thượng, H. Đức Hòa, T. Long An, Việt Nam
//ManufactureDate		: Năm sản xuất: 2025
//USSize				:
//VNSize				: Tương ứng với cột Size trên Google Sheet VN
//UKSize                :
//Gender                : GENDER
//Attribute             : *Bên dưới có note rõ
//SizeAvailable         : *Bên dưới có note rõ
//QrCode                : https://hasaki.vn/

// Check bảng size sheet trên link
// Hiện tại đang làm việc với bảng size của VN
// MEN VÀ WOMEN hiện tại sẽ có cái size: XS, S, M, L, XL, 2XL, 3XL
// Field Attribute với women:
// Attribute            :  W/42-45 kg - H/150-155 cm
// Ví dụ đối với WOMEN có S thì attribute : W/42-45 kg - H/150-155 cm
// Tương ứng với mỗi size của women sẽ có một attribute khác nhau

// Field Attribute với man:
// Attribute            :  W/62-68 kg - H/166-172 cm
// Ví dụ đối với MEN có S thì attribute : W/62-68 kg - H/166-172 cm
// Tương ứng với mỗi size của men sẽ có một attribute khác nhau

// Field Attribute với kids:
// Attribute            :  W/14-16 kg - H/88-95 cm
// Ví dụ đối với KIDS có S thì attribute : W/14-16 kg - H/88-95 cm
// Tương ứng với mỗi size của kids sẽ có một attribute khác nhau

// Hiện tại đối với SizeAvailable tương ứng với mối size (VNSize, UKSize, USSize) sẽ có một giá trị khác nhau
// Hiện tại đang làm việc với bảng size của VN
// Women và men sẽ dùng chung ảnh SizeAvailable này:
// XS : https://minio.inshasaki.com/bom-prod/qc/2025/07/02/product_attribute_aQqE_20250702095012.png
// S  : https://minio.inshasaki.com/bom-prod/qc/2025/07/02/product_attribute_aQqE_20250702095012.png
// M  : https://minio.inshasaki.com/bom-prod/qc/2025/07/02/product_attribute_nDGc_20250702095516.png
// L  : https://minio.inshasaki.com/bom-prod/qc/2025/07/02/product_attribute_mhTf_20250702095555.png
// XL : https://minio.inshasaki.com/bom-prod/qc/2025/07/02/product_attribute_1gr0_20250702095638.png
// 2XL: https://minio.inshasaki.com/bom-prod/qc/2025/07/02/product_attribute_89or_20250702095722.png
// 3XL: chưa có hình ảnh, check data sheet VN hiện tại vẫn chưa thấy product có size 3XL

// Kids sẽ SizeAvailable này: 3T, 4T, 5, 6 ,7 ,8
// 2T : https://minio.inshasaki.com/bom-prod/qc/2025/07/02/product_attribute_xqXo_20250702100352.png
// 3T : https://minio.inshasaki.com/bom-prod/qc/2025/07/02/product_attribute_QvCJ_20250702100426.png
// 4T : https://minio.inshasaki.com/bom-prod/qc/2025/07/02/product_attribute_fiBO_20250702100504.png
// 5  : https://minio.inshasaki.com/bom-prod/qc/2025/07/02/product_attribute_yAXy_20250702100542.png
// 6  : https://minio.inshasaki.com/bom-prod/qc/2025/07/02/product_attribute_kNqX_20250702100606.png
// 7  : https://minio.inshasaki.com/bom-prod/qc/2025/07/02/product_attribute_lsHB_20250702100634.png
// 8  : https://minio.inshasaki.com/bom-prod/qc/2025/07/02/product_attribute_0WJk_20250702100655.png

//-------------------------------US-------------------------------
//Product Attribute 	: Sheet Column
//Name 					: PRODUCT NAME
//Code 					: STYLE#
//Color 				: COLOR NAME
//Material 				: MATERIAL
//ManufactureOffice 	: Văn phòng: Lầu 3 555 3/2, P.8, Q.10, TP.HCM, Việt Nam
//ManuFactureCompany 	: 130 Ấp Chánh, X. Đức Lập Thượng, H. Đức Hòa, T. Long An, Việt Nam
//ManufactureDate		: Năm sản xuất: 2025
//USSize				:
//VNSize				:
//UKSize                :
//Gender                : GENDER
//Attribute             :
//SizeAvailable         :
//QrCode                :
