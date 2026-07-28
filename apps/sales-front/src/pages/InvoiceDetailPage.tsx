import { Button, Descriptions, Space, Spin, Table, Tag, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { getInvoiceDoc } from '@/services/invoiceDocApi'
import type { InvoiceDoc, InvoiceDocLine } from '@/types/invoiceDoc'
import { formatOrderDate } from '@/utils/dateFormat'
import { formatAmount, formatQuantity } from '@/utils/money'

export default function InvoiceDetailPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)
  const [doc, setDoc] = useState<InvoiceDoc | null>(null)

  useEffect(() => {
    if (!id) {
      navigate('/invoices')
      return
    }
    let cancelled = false
    ;(async () => {
      try {
        const data = await getInvoiceDoc(id)
        if (!cancelled) setDoc(data)
      } catch (e) {
        message.error(e instanceof Error ? e.message : '加载失败')
        navigate('/invoices')
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [id, navigate])

  if (loading || !doc) return <Spin />

  const columns: ColumnsType<InvoiceDocLine> = [
    { title: '物资', dataIndex: 'materialName' },
    { title: '规格型号', dataIndex: 'spec' },
    { title: '单位', dataIndex: 'unit', width: 60 },
    {
      title: '数量',
      dataIndex: 'quantity',
      align: 'right',
      render: (v: number) => formatQuantity(v),
    },
    {
      title: '单价',
      dataIndex: 'unitPrice',
      align: 'right',
      render: (v: number) => formatQuantity(v),
    },
    {
      title: '金额',
      dataIndex: 'amount',
      align: 'right',
      render: (v: number) => formatAmount(v),
    },
    { title: '行备注', dataIndex: 'lineRemark' },
    { title: '单据号', dataIndex: 'orderNo' },
    {
      title: '单据日期',
      dataIndex: 'orderDate',
      render: (v: string) => formatOrderDate(v),
    },
    { title: '客户', dataIndex: 'customerName' },
    { title: '仓库', dataIndex: 'warehouseName' },
    { title: '出库类型', dataIndex: 'deliveryType' },
  ]

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button onClick={() => navigate('/invoices')}>返回列表</Button>
      </Space>
      <Descriptions bordered size="small" column={2} style={{ marginBottom: 16 }}>
        <Descriptions.Item label="开票单号">{doc.invoiceNo}</Descriptions.Item>
        <Descriptions.Item label="状态">
          {doc.status === 'voided' ? (
            <Tag>已作废</Tag>
          ) : (
            <Tag color="green">已保存</Tag>
          )}
        </Descriptions.Item>
        <Descriptions.Item label="筛选客户">
          {doc.filterCustomerName || '—'}
        </Descriptions.Item>
        <Descriptions.Item label="日期范围">
          {doc.filterDateFrom || doc.filterDateTo
            ? `${doc.filterDateFrom ?? '…'} ~ ${doc.filterDateTo ?? '…'}`
            : '—'}
        </Descriptions.Item>
        <Descriptions.Item label="单据号">
          {doc.filterOrderNos?.join(', ') || '—'}
        </Descriptions.Item>
        <Descriptions.Item label="金额合计">
          {formatAmount(doc.totalAmount)}
        </Descriptions.Item>
      </Descriptions>
      <Table
        rowKey="id"
        size="small"
        columns={columns}
        dataSource={doc.lines}
        pagination={false}
        scroll={{ x: 1200 }}
      />
    </div>
  )
}
