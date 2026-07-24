import { Button, Input, InputNumber, Space, Table } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import type { SalesOrderLine } from '@/types/salesOrder'
import {
  calcLineAmount,
  formatAmount,
  formatQuantity,
  sumAmounts,
  sumQuantities,
} from '@/utils/money'

type Props = {
  value: SalesOrderLine[]
  onChange: (lines: SalesOrderLine[]) => void
}

function updateLine(
  lines: SalesOrderLine[],
  id: string,
  patch: Partial<SalesOrderLine>,
): SalesOrderLine[] {
  return lines.map((line) => {
    if (line.id !== id) return line
    const next = { ...line, ...patch }
    if (patch.quantity !== undefined || patch.unitPrice !== undefined) {
      next.amount = calcLineAmount(next.quantity, next.unitPrice)
    }
    return next
  })
}

export default function OrderLinesTable({ value, onChange }: Props) {
  const columns: ColumnsType<SalesOrderLine> = [
    {
      title: '物资',
      dataIndex: 'materialName',
      render: (_, row) => (
        <Input
          value={row.materialName}
          onChange={(e) =>
            onChange(updateLine(value, row.id, { materialName: e.target.value }))
          }
        />
      ),
    },
    {
      title: '规格型号',
      dataIndex: 'spec',
      width: 120,
      render: (_, row) => (
        <Input
          value={row.spec}
          onChange={(e) => onChange(updateLine(value, row.id, { spec: e.target.value }))}
        />
      ),
    },
    {
      title: '单位',
      dataIndex: 'unit',
      width: 80,
      render: (_, row) => (
        <Input
          value={row.unit}
          onChange={(e) => onChange(updateLine(value, row.id, { unit: e.target.value }))}
        />
      ),
    },
    {
      title: '数量',
      dataIndex: 'quantity',
      width: 110,
      render: (_, row) => (
        <InputNumber
          style={{ width: '100%' }}
          min={0}
          value={row.quantity}
          onChange={(v) =>
            onChange(updateLine(value, row.id, { quantity: Number(v ?? 0) }))
          }
        />
      ),
    },
    {
      title: '单价',
      dataIndex: 'unitPrice',
      width: 110,
      render: (_, row) => (
        <InputNumber
          style={{ width: '100%' }}
          min={0}
          step={0.01}
          value={row.unitPrice}
          onChange={(v) =>
            onChange(updateLine(value, row.id, { unitPrice: Number(v ?? 0) }))
          }
        />
      ),
    },
    {
      title: '金额',
      dataIndex: 'amount',
      width: 110,
      align: 'right',
      render: (v: number) => formatAmount(v),
    },
    {
      title: '备注',
      dataIndex: 'lineRemark',
      width: 120,
      render: (_, row) => (
        <Input
          value={row.lineRemark ?? ''}
          onChange={(e) =>
            onChange(updateLine(value, row.id, { lineRemark: e.target.value }))
          }
        />
      ),
    },
    {
      title: '',
      width: 70,
      render: (_, row) => (
        <Button
          type="link"
          danger
          disabled={value.length <= 1}
          onClick={() => onChange(value.filter((l) => l.id !== row.id))}
        >
          删除
        </Button>
      ),
    },
  ]

  return (
    <div>
      <Table
        rowKey="id"
        size="small"
        pagination={false}
        columns={columns}
        dataSource={value}
      />
      <Space style={{ marginTop: 12 }} wrap>
        <Button
          onClick={() =>
            onChange([
              ...value,
              {
                id: crypto.randomUUID(),
                materialName: '',
                spec: '',
                unit: 'kg',
                quantity: 0,
                unitPrice: 0,
                amount: 0,
              },
            ])
          }
        >
          添加行
        </Button>
        <span>数量合计：{formatQuantity(sumQuantities(value))}</span>
        <span>金额合计：{formatAmount(sumAmounts(value))}</span>
      </Space>
    </div>
  )
}
