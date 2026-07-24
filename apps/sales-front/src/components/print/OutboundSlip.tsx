import type { CSSProperties } from 'react'
import type { PrintProfile } from '@/types/printProfile'
import type { SalesOrder } from '@/types/salesOrder'
import { formatOrderDate } from '@/utils/dateFormat'
import { formatAmount, formatQuantity } from '@/utils/money'
import { padLinesToMin } from './padLines'
import './outboundSlip.css'

type Props = {
  order: SalesOrder
  profile: PrintProfile
}

export default function OutboundSlip({ order, profile }: Props) {
  const rows = padLinesToMin(order.lines, profile.tableMinRows)
  const title = `${profile.companyName}${profile.titleSuffix}`

  return (
    <div
      className="outbound-slip"
      style={
        {
          '--slip-width': `${profile.paperWidthMm}mm`,
          '--slip-height': `${profile.paperHeightMm}mm`,
        } as CSSProperties
      }
    >
      <h1 className="outbound-slip__title">{title}</h1>
      <div className="outbound-slip__meta">
        <span>日期：{formatOrderDate(order.orderDate)}</span>
        <span>单据号：{order.orderNo}</span>
      </div>
      <div className="outbound-slip__fields">
        <span>客户：{order.customerName}</span>
        <span>仓库：{order.warehouseName}</span>
        <span>出库类型：{order.deliveryType}</span>
        {profile.bankCardNo ? (
          <span>
            农行卡号：{profile.bankCardNo}
            {profile.bankCardHolder ? `（${profile.bankCardHolder}）` : ''}
          </span>
        ) : null}
      </div>
      <table className="outbound-slip__table">
        <colgroup>
          <col style={{ width: '28%' }} />
          <col style={{ width: '14%' }} />
          <col style={{ width: '8%' }} />
          <col style={{ width: '12%' }} />
          <col style={{ width: '12%' }} />
          <col style={{ width: '14%' }} />
          <col style={{ width: '12%' }} />
        </colgroup>
        <thead>
          <tr>
            <th>物资</th>
            <th>规格型号</th>
            <th>单位</th>
            <th>数量</th>
            <th>单价</th>
            <th>金额</th>
            <th>备注</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((line, index) => (
            <tr key={line?.id ?? `empty-${index}`}>
              <td>{line?.materialName ?? ''}</td>
              <td>{line?.spec ?? ''}</td>
              <td style={{ textAlign: 'center' }}>{line?.unit ?? ''}</td>
              <td className="outbound-slip__num">
                {line ? formatQuantity(line.quantity) : ''}
              </td>
              <td className="outbound-slip__num">
                {line ? formatQuantity(line.unitPrice) : ''}
              </td>
              <td className="outbound-slip__num">
                {line ? formatAmount(line.amount) : ''}
              </td>
              <td>{line?.lineRemark ?? ''}</td>
            </tr>
          ))}
        </tbody>
      </table>
      <div className="outbound-slip__totals">
        <span>数量合计：{formatQuantity(order.totalQuantity)}</span>
        <span>合计：{formatAmount(order.totalAmount)}</span>
      </div>
      <div className="outbound-slip__footer">
        <span>地址：{profile.address}</span>
        <span>电话：{profile.phone}</span>
        <span>备注：{order.remark ?? ''}</span>
      </div>
      <div className="outbound-slip__pickup">提货人：______________</div>
      <div className="outbound-slip__note">{profile.qualityNote}</div>
    </div>
  )
}
