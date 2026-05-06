package entity

type Region string

const (
	RegionAddisAbaba  Region = "ADDIS_ABABA"
	RegionDireDawa    Region = "DIRE_DAWA"
	RegionOromia      Region = "OROMIA"
	RegionAmhara      Region = "AMHARA"
	RegionSidama      Region = "SIDAMA"
	RegionSomali      Region = "SOMALI"
	RegionTigray      Region = "TIGRAY"
	RegionAfar        Region = "AFAR"
	RegionHarari      Region = "HARARI"
	RegionBenishangul Region = "BENISHANGUL_GUMUZ"
	RegionSWEPR       Region = "SWEPR" // South West Ethiopia Peoples' Region
	RegionCentral     Region = "CENTRAL_ETHIOPIA"
	RegionSouth       Region = "SOUTH_ETHIOPIA"
	RegionFederal     Region = "FEDERAL" // For federal-level regulations
)

type BusinessStage string

const (
	StageIdea         BusinessStage = "IDEA"
	StageRegistration BusinessStage = "REGISTRATION"
	StageOperational  BusinessStage = "OPERATIONAL"
	StageScaling      BusinessStage = "SCALING"
)

type TagGroup string

const (
	TagGroupLegalStructure    TagGroup = "LEGAL_STRUCTURE"
	TagGroupTaxStatus         TagGroup = "TAX_STATUS"
	TagGroupGeneralOperations TagGroup = "GENERAL_OPERATIONS"
	TagGroupEmployment        TagGroup = "EMPLOYMENT"
	TagGroupDemographics      TagGroup = "DEMOGRAPHICS"
)
