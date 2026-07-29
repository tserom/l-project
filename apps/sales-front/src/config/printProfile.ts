import type { PrintProfile } from '@/types/printProfile';

export const defaultPrintProfile: PrintProfile = {
  companyName: '淄博钰鑫不锈钢有限公司',
  titleSuffix: '出库单',
  bankCardNo: '621700310010367346',
  bankCardHolder: '马建平',
  address: '周村区北方不锈钢市场17-16',
  phone: '15065851366',
  qualityNote:
    '备注：如出现质量问题请于收到货7日内以书面形式提出质量异议，逾期则视为合格产品。35mm以上棒料严禁下料机（冲床）下料，出现裂纹后果自负。',
  paperWidthMm: 190,
  paperHeightMm: 140,
  tableMinRows: 10,
};
