import { Button, Empty, Modal, Space, Table, Tag, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { listInvoiceDocs, voidInvoiceDoc } from '@/services/invoiceDocApi'
import type { InvoiceDoc } from '@/types/invoiceDoc'
import { formatOrderDate } from '@/utils/dateFormat'
import { formatAmount, formatQuantity } from '@/utils/money'

export default function InvoiceListPage() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)
  const [docs, setDocs] = useState<InvoiceDoc[]>([])

  const reload = useCallback(async () => {
    setLoading(true)
    try {
      setDocs(await listInvoiceDocs())
    } catch (e) {
      message.error(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void reload()
  }, [reload])

  const columns: ColumnsType<InvoiceDoc> = [
    { title: '开票单号', dataIndex: 'invoiceNo', key: 'invoiceNo' },
    {
      title: '状态',
      dataIndex: 'status',
      render: (s: InvoiceDoc['status']) =>
        s === 'voided' ? <Tag>已作废</Tag> : <Tag color="green">已保存</Tag>,
    },
    {
      title: '筛选客户',
      dataIndex: 'filterCustomerName',
      render: (v?: string) => v || '—',
    },
    {
      title: '日期范围',
      key: 'dates',
      render: (_, r) => {
        if (!r.filterDateFrom && !r.filterDateTo) return '—'
        const a = r.filterDateFrom ? formatOrderDate(r.filterDateFrom) : '…'
        const b = r.filterDateTo ? formatOrderDate(r.filterDateTo) : '…'
        return `${a} ~ ${b}`
      },
    },
    {
      title: '数量合计',
      dataIndex: 'totalQuantity',
      align: 'right',
      render: (v: number) => formatQuantity(v),
    },
    {
      title: '金额合计',
      dataIndex: 'totalAmount',
      align: 'right',
      render: (v: number) => formatAmount(v),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Link to={`/invoices/${record.id}`}>查看</Link>
          {record.status === 'saved' ? (
            <Button
              type="link"
              danger
              onClick={() => {
                Modal.confirm({
                  title: '确认作废该开票单？',
                  content: '作废后，相关销售明细将重新变为待开票。',
                  onOk: async () => {
                    await voidInvoiceDoc(record.id)
                    message.success('已作废')
                    await reload()
                  },
                })
              }}
            >
              作废
            </Button>
          ) : null}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => navigate('/invoices/new')}>
          新建开票单
        </Button>
      </Space>
      <Table
        rowKey="id"
        loading={loading}
        columns={columns}
        dataSource={docs}
        pagination={false}
        locale={{
          emptyText: <Empty description="暂无开票单" />,
        }}
      />
    </div>
  )
}
