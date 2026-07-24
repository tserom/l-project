import type { PrintProfile } from '@/types/printProfile'

export const defaultPrintProfile: PrintProfile = {
  companyName: '泰州市金阳金属制品有限公司',
  titleSuffix: '出库单',
  bankCardNo: '6230523420034407572',
  bankCardHolder: '刘敏',
  address: '兴化市戴南镇裴马工业区',
  phone: '13705265020',
  qualityNote:
    '备注：如出现质量问题请于收到货7日内以书面形式提出质量异议，逾期则视为合格产品。35mm以上棒料严禁下料机（冲床）下料，出现裂纹后果自负。',
  paperWidthMm: 190,
  paperHeightMm: 140,
  tableMinRows: 10,
}
