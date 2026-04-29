package entity

type TemplateFormat string

const (
	TemplateFormatPDF             TemplateFormat = "pdf"
	TemplateFormatDOCX            TemplateFormat = "docx"
	TemplateFormatXLSX            TemplateFormat = "xlsx"
	TemplateFormatInteractiveForm TemplateFormat = "interactive_form"
)

type TierAccess string

const (
	TierAccessBasic TierAccess = "basic"
	TierAccessPro   TierAccess = "pro"
)
