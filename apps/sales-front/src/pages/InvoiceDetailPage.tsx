import { Button, Descriptions, Space, Spin, Tag, message } from 'antd'
import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import InvoiceLinesView from '@/components/InvoiceLinesView'
import { getInvoiceDoc } from '@/services/invoiceDocApi'
import type { InvoiceDoc } from '@/types/invoiceDoc'
import { formatAmount } from '@/utils/money'

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
      <InvoiceLinesView lines={doc.lines} />
    </div>
  )
}
