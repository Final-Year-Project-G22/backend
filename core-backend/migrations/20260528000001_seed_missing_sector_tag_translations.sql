-- ---------------------------------------------------------------------------------
-- 1. CHILD SECTOR TRANSLATIONS (EN + AM)
-- ---------------------------------------------------------------------------------
INSERT INTO sector_translations (sector_id, language, name, description)
-- Trade
SELECT id, 'en', 'Retail', 'Direct-to-consumer goods sales' FROM sectors WHERE slug = 'retail' UNION ALL
SELECT id, 'am', 'ችርቻሮ', 'ለተጠቃሚዎች በቀጥታ የሚሸጡ ዕቃዎች' FROM sectors WHERE slug = 'retail' UNION ALL

SELECT id, 'en', 'Wholesale', 'Bulk goods distribution to retailers' FROM sectors WHERE slug = 'wholesale' UNION ALL
SELECT id, 'am', 'ጅምላ ንግድ', 'ለችርቻሮ ነጋዴዎች በጅምላ የሚከፋፈሉ ዕቃዎች' FROM sectors WHERE slug = 'wholesale' UNION ALL

-- Manufacturing
SELECT id, 'en', 'Food & Beverage', 'Food and drink production' FROM sectors WHERE slug = 'food-beverage' UNION ALL
SELECT id, 'am', 'ምግብ እና መጠጥ', 'የምግብ እና መጠጥ ማቀነባበሪያ' FROM sectors WHERE slug = 'food-beverage' UNION ALL

SELECT id, 'en', 'Textiles & Apparel', 'Clothing and fabric production' FROM sectors WHERE slug = 'textiles-apparel' UNION ALL
SELECT id, 'am', 'ጨርቃጨርቅ እና አልባሳት', 'የልብስ እና የጨርቃጨርቅ ምርት' FROM sectors WHERE slug = 'textiles-apparel' UNION ALL

SELECT id, 'en', 'Leather', 'Leather goods manufacturing' FROM sectors WHERE slug = 'leather' UNION ALL
SELECT id, 'am', 'ቆዳ እና ሌጦ', 'የቆዳ ምርቶች ማቀነባበሪያ' FROM sectors WHERE slug = 'leather' UNION ALL

SELECT id, 'en', 'Wood & Metal', 'Woodworking and metal fabrication' FROM sectors WHERE slug = 'wood-metal' UNION ALL
SELECT id, 'am', 'እንጨት እና ብረት', 'የእንጨት ሥራ እና የብረት ውጤቶች ማቀነባበሪያ' FROM sectors WHERE slug = 'wood-metal' UNION ALL

-- Services
SELECT id, 'en', 'IT & Technology', 'Software, hardware, and tech services' FROM sectors WHERE slug = 'it-tech' UNION ALL
SELECT id, 'am', 'አይቲ እና ቴክኖሎጂ', 'ሶፍትዌር፣ ሃርድዌር እና የቴክኖሎጂ አገልግሎቶች' FROM sectors WHERE slug = 'it-tech' UNION ALL

SELECT id, 'en', 'Hospitality', 'Hotels, restaurants, and tourism' FROM sectors WHERE slug = 'hospitality' UNION ALL
SELECT id, 'am', 'እንግዳ መቀበያ', 'ሆቴሎች፣ ምግብ ቤቶች እና ቱሪዝም' FROM sectors WHERE slug = 'hospitality' UNION ALL

SELECT id, 'en', 'Consulting', 'Professional advisory services' FROM sectors WHERE slug = 'consulting' UNION ALL
SELECT id, 'am', 'የምክር አገልግሎት', 'ሙያዊ የምክር አገልግሎቶች' FROM sectors WHERE slug = 'consulting' UNION ALL

SELECT id, 'en', 'Transport & Logistics', 'Freight and passenger transport' FROM sectors WHERE slug = 'transport-logistics' UNION ALL
SELECT id, 'am', 'ትራንስፖርት እና ሎጂስቲክስ', 'የጭነት እና የተሳፋሪዎች ትራንስፖርት' FROM sectors WHERE slug = 'transport-logistics' UNION ALL

-- Agriculture
SELECT id, 'en', 'Crop Farming', 'Cultivation of crops and grains' FROM sectors WHERE slug = 'crop-farming' UNION ALL
SELECT id, 'am', 'የሰብል ልማት', 'የሰብል እና የእህል እርሻ' FROM sectors WHERE slug = 'crop-farming' UNION ALL

SELECT id, 'en', 'Livestock & Poultry', 'Animal husbandry and poultry farming' FROM sectors WHERE slug = 'livestock-poultry' UNION ALL
SELECT id, 'am', 'የከብት እና የዶሮ እርባታ', 'የእንስሳት እና የዶሮ እርባታ' FROM sectors WHERE slug = 'livestock-poultry' UNION ALL

-- Construction
SELECT id, 'en', 'Contracting', 'General construction contracting' FROM sectors WHERE slug = 'contracting' UNION ALL
SELECT id, 'am', 'ኮንትራት', 'ጠቅላላ የግንባታ ኮንትራት ሥራዎች' FROM sectors WHERE slug = 'contracting' UNION ALL

SELECT id, 'en', 'Construction Materials', 'Building materials supply' FROM sectors WHERE slug = 'construction-mat' UNION ALL
SELECT id, 'am', 'የግንባታ ቁሳቁስ', 'የግንባታ ግብዓቶች አቅርቦት' FROM sectors WHERE slug = 'construction-mat';

-- ---------------------------------------------------------------------------------
-- 2. TAG TRANSLATIONS (EN + AM)
-- ---------------------------------------------------------------------------------
INSERT INTO tag_translations (tag_id, language, name, description)
-- Legal Structure
SELECT id, 'en', 'Share Company', 'Company owned by shareholders' FROM tags WHERE slug = 'share-company' UNION ALL
SELECT id, 'am', 'አክሲዮን ማህበር', 'በአክሲዮን ባለቤቶች የሚተዳደር ድርጅት' FROM tags WHERE slug = 'share-company' UNION ALL

SELECT id, 'en', 'Partnership', 'Business owned by two or more partners' FROM tags WHERE slug = 'partnership' UNION ALL
SELECT id, 'am', 'ሽርክና', 'በሁለት ወይም ከዚያ በላይ በሆኑ አጋሮች ባለቤትነት የተያዘ ንግድ' FROM tags WHERE slug = 'partnership' UNION ALL

SELECT id, 'en', 'Cooperative', 'Member-owned cooperative business' FROM tags WHERE slug = 'cooperative' UNION ALL
SELECT id, 'am', 'ህብረት ስራ ማህበር', 'በአባላት ባለቤትነት የሚመራ የህብረት ስራ ንግድ' FROM tags WHERE slug = 'cooperative' UNION ALL

-- Tax Status
SELECT id, 'en', 'Excise Payer', 'Business paying excise tax on specific goods' FROM tags WHERE slug = 'tax-excise' UNION ALL
SELECT id, 'am', 'ኤክሳይስ ታክስ ከፋይ', 'በተወሰኑ ዕቃዎች ላይ የኤክሳይስ ታክስ የሚከፍል ንግድ' FROM tags WHERE slug = 'tax-excise' UNION ALL

SELECT id, 'en', 'Tax Exempt', 'Business exempt from certain taxes' FROM tags WHERE slug = 'tax-exempt' UNION ALL
SELECT id, 'am', 'ከታክስ ነፃ', 'ከተወሰኑ ታክሶች ነፃ የሆነ ንግድ' FROM tags WHERE slug = 'tax-exempt' UNION ALL

-- Operations
SELECT id, 'en', 'Tender Participant', 'Business participating in government tenders' FROM tags WHERE slug = 'op-tender' UNION ALL
SELECT id, 'am', 'የጨረታ ተወዳዳሪ', 'በመንግስት ጨረታዎች የሚሳተፍ ድርጅት' FROM tags WHERE slug = 'op-tender' UNION ALL

SELECT id, 'en', 'Vehicle Operator', 'Business operating vehicles for commercial use' FROM tags WHERE slug = 'op-vehicles' UNION ALL
SELECT id, 'am', 'የተሽከርካሪ ኦፕሬተር', 'ለንግድ አገልግሎት ተሽከርካሪዎችን የሚያንቀሳቅስ ድርጅት' FROM tags WHERE slug = 'op-vehicles' UNION ALL

SELECT id, 'en', 'Handles Hazardous Materials', 'Deals with hazardous or controlled substances' FROM tags WHERE slug = 'op-hazardous' UNION ALL
SELECT id, 'am', 'አደገኛ ቁሳቁሶችን የሚያስተናግድ', 'አደገኛ ወይም ቁጥጥር የሚደረግባቸው ቁሳቁሶችን የሚጠቀም' FROM tags WHERE slug = 'op-hazardous' UNION ALL

SELECT id, 'en', 'E-Commerce', 'Online sales and digital commerce' FROM tags WHERE slug = 'op-ecommerce' UNION ALL
SELECT id, 'am', 'ኢ-ኮሜርስ', 'የኦንላይን ሽያጭ እና ዲጂታል ንግድ' FROM tags WHERE slug = 'op-ecommerce' UNION ALL

SELECT id, 'en', 'Home-Based Business', 'Business operated from home' FROM tags WHERE slug = 'op-home-based' UNION ALL
SELECT id, 'am', 'ከቤት የሚሰራ ንግድ', 'ከቤት ሆኖ የሚንቀሳቀስ ንግድ' FROM tags WHERE slug = 'op-home-based' UNION ALL

SELECT id, 'en', 'Creates Intellectual Property', 'Produces patents, trademarks, or copyrighted works' FROM tags WHERE slug = 'op-creates-ip' UNION ALL
SELECT id, 'am', 'የአእምሮ ንብረት ፈጣሪ', 'ፓተንት፣ የንግድ ምልክት ወይም የቅጂ መብት ስራዎችን የሚያመርት' FROM tags WHERE slug = 'op-creates-ip' UNION ALL

-- Demographics
SELECT id, 'en', 'Investor', 'Active angel or venture investor' FROM tags WHERE slug = 'demo-investor' UNION ALL
SELECT id, 'am', 'ባለሀብት', 'የኢንቨስትመንት ስራዎችን የሚያከናውን' FROM tags WHERE slug = 'demo-investor';