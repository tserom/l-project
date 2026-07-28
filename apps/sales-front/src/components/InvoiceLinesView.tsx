import { Segmented, Table } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useMemo, useState } from 'react'
import type { InvoiceDocLine } from '@/types/invoiceDoc'
import { formatOrderDate } from '@/utils/dateFormat'
import { mergeInvoiceLines, type MergedInvoiceLine } from '@/utils/mergeInvoiceLines'
import { formatAmount, formatQuantity } from '@/utils/money'

type ViewMode = 'detail' | 'merged'

type Props = {
  lines: InvoiceDocLine[]
  /** 明细视图右侧操作列（如删除） */
  detailExtraColumns?: ColumnsType<InvoiceDocLine>
}

export default function InvoiceLinesView({ lines, detailExtraColumns }: Props) {
  const [mode, setMode] = useState<ViewMode>('detail')
  const merged = useMemo(() => mergeInvoiceLines(lines), [lines])

  const detailColumns: ColumnsType<InvoiceDocLine> = [
    { title: '物资', dataIndex: 'materialName' },
    { title: '规格型号', dataIndex: 'spec', width: 100 },
    { title: '单位', dataIndex: 'unit', width: 60 },
    {
      title: '数量',
      dataIndex: 'quantity',
      width: 90,
      align: 'right',
      render: (v: number) => formatQuantity(v),
    },
    {
      title: '单价',
      dataIndex: 'unitPrice',
      width: 90,
      align: 'right',
      render: (v: number) => formatQuantity(v),
    },
    {
      title: '金额',
      dataIndex: 'amount',
      width: 100,
      align: 'right',
      render: (v: number) => formatAmount(v),
    },
    { title: '行备注', dataIndex: 'lineRemark', width: 100 },
    { title: '单据号', dataIndex: 'orderNo', width: 120 },
    {
      title: '单据日期',
      dataIndex: 'orderDate',
      width: 120,
      render: (v: string) => formatOrderDate(v),
    },
    { title: '客户', dataIndex: 'customerName', width: 140 },
    { title: '仓库', dataIndex: 'warehouseName', width: 110 },
    { title: '出库类型', dataIndex: 'deliveryType', width: 90 },
    ...(detailExtraColumns ?? []),
  ]

  const mergedColumns: ColumnsType<MergedInvoiceLine> = [
    { title: '物资', dataIndex: 'materialName' },
    { title: '规格型号', dataIndex: 'spec', width: 100 },
    { title: '单位', dataIndex: 'unit', width: 60 },
    {
      title: '数量',
      dataIndex: 'quantity',
      width: 90,
      align: 'right',
      render: (v: number) => formatQuantity(v),
    },
    {
      title: '单价',
      dataIndex: 'unitPrice',
      width: 90,
      align: 'right',
      render: (v: number) => formatQuantity(v),
    },
    {
      title: '金额',
      dataIndex: 'amount',
      width: 100,
      align: 'right',
      render: (v: number) => formatAmount(v),
    },
    {
      title: '合并笔数',
      dataIndex: 'sourceCount',
      width: 90,
      align: 'right',
    },
    { title: '单据号', dataIndex: 'orderNos', width: 160 },
    {
      title: '单据日期',
      dataIndex: 'orderDates',
      width: 160,
      render: (v: string) =>
        v
          .split('、')
          .map((d) => {
            try {
              return formatOrderDate(d)
            } catch {
              return d
            }
          })
          .join('、'),
    },
    { title: '客户', dataIndex: 'customerNames', width: 140 },
    { title: '仓库', dataIndex: 'warehouseNames', width: 110 },
    { title: '出库类型', dataIndex: 'deliveryTypes', width: 90 },
  ]

  return (
    <div>
      <div style={{ marginBottom: 12 }}>
        <Segmented
          value={mode}
          onChange={(v) => setMode(v as ViewMode)}
          options={[
            { label: '明细视图', value: 'detail' },
            { label: '合并视图', value: 'merged' },
          ]}
        />
        <span style={{ marginLeft: 12, color: '#666', fontSize: 13 }}>
          合并规则：物资 + 规格型号 + 单价相同则合并数量
        </span>
      </div>
      {mode === 'detail' ? (
        <Table
          rowKey="id"
          size="small"
          columns={detailColumns}
          dataSource={lines}
          pagination={false}
          scroll={{ x: 1400 }}
        />
      ) : (
        <Table
          rowKey="key"
          size="small"
          columns={mergedColumns}
          dataSource={merged}
          pagination={false}
          scroll={{ x: 1300 }}
        />
      )}
    </div>
  )
}
