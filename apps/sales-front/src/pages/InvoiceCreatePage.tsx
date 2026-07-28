import {
  Button,
  Card,
  Col,
  DatePicker,
  Form,
  Input,
  Row,
  Select,
  Space,
  Table,
  message,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import type { Dayjs } from 'dayjs'
import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  createInvoiceDoc,
  listInvoiceCandidates,
} from '@/services/invoiceDocApi'
import type { InvoiceCandidateLine } from '@/types/invoiceDoc'
import { formatOrderDate } from '@/utils/dateFormat'
import { parseOrderNos } from '@/utils/filterSalesOrders'
import { formatAmount, formatQuantity, sumAmounts, sumQuantities } from '@/utils/money'

type FilterForm = {
  customerName?: string
  dateRange?: [Dayjs, Dayjs] | null
  orderNos?: string[]
}

export default function InvoiceCreatePage() {
  const navigate = useNavigate()
  const [form] = Form.useForm<FilterForm>()
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [lines, setLines] = useState<InvoiceCandidateLine[]>([])
  const [loadedFilter, setLoadedFilter] = useState<FilterForm>({})

  const onLoad = async () => {
    const values = form.getFieldsValue()
    setLoading(true)
    try {
      const range = values.dateRange
      const rows = await listInvoiceCandidates({
        customerName: values.customerName,
        dateFrom: range?.[0]?.format('YYYY-MM-DD'),
        dateTo: range?.[1]?.format('YYYY-MM-DD'),
        orderNos: parseOrderNos(values.orderNos),
      })
      setLines(rows)
      setLoadedFilter(values)
      if (rows.length === 0) {
        message.info('没有符合条件的待开票明细')
      } else {
        message.success(`已加载 ${rows.length} 行`)
      }
    } catch (e) {
      message.error(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  const onSave = async () => {
    if (!lines.length) {
      message.error('请先加载并保留至少一行明细')
      return
    }
    setSaving(true)
    try {
      const range = loadedFilter.dateRange
      const doc = await createInvoiceDoc({
        filterCustomerName: loadedFilter.customerName,
        filterDateFrom: range?.[0]?.format('YYYY-MM-DD'),
        filterDateTo: range?.[1]?.format('YYYY-MM-DD'),
        filterOrderNos: parseOrderNos(loadedFilter.orderNos),
        lines,
      })
      message.success(`已保存开票单 ${doc.invoiceNo}`)
      navigate(`/invoices/${doc.id}`)
    } catch (e) {
      message.error(e instanceof Error ? e.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const totals = useMemo(
    () => ({
      qty: sumQuantities(lines),
      amount: sumAmounts(lines),
    }),
    [lines],
  )

  const columns: ColumnsType<InvoiceCandidateLine> = [
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
    {
      title: '',
      width: 70,
      fixed: 'right',
      render: (_, row) => (
        <Button
          type="link"
          danger
          onClick={() => setLines((prev) => prev.filter((l) => l.id !== row.id))}
        >
          删除
        </Button>
      ),
    },
  ]

  return (
    <div>
      <h2 style={{ marginTop: 0 }}>新建开票单</h2>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Form form={form} layout="vertical" onFinish={() => void onLoad()}>
          <Row gutter={16}>
            <Col xs={24} md={8}>
              <Form.Item name="customerName" label="客户（可选）">
                <Input placeholder="模糊匹配，可留空" allowClear />
              </Form.Item>
            </Col>
            <Col xs={24} md={8}>
              <Form.Item name="dateRange" label="销售单日期（可选）">
                <DatePicker.RangePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col xs={24} md={8}>
              <Form.Item
                name="orderNos"
                label="单据号（可选）"
                extra="多值，回车或逗号分隔"
              >
                <Select
                  mode="tags"
                  tokenSeparators={[',', '，', ' ']}
                  placeholder="销售单据号"
                  allowClear
                />
              </Form.Item>
            </Col>
          </Row>
          <Space>
            <Button type="primary" htmlType="submit" loading={loading}>
              加载明细
            </Button>
            <Button onClick={() => navigate('/invoices')}>返回</Button>
          </Space>
        </Form>
      </Card>

      <Table
        rowKey="id"
        size="small"
        loading={loading}
        columns={columns}
        dataSource={lines}
        pagination={false}
        scroll={{ x: 1400 }}
        summary={() => (
          <Table.Summary fixed>
            <Table.Summary.Row>
              <Table.Summary.Cell index={0} colSpan={3}>
                合计（{lines.length} 行）
              </Table.Summary.Cell>
              <Table.Summary.Cell index={1} align="right">
                {formatQuantity(totals.qty)}
              </Table.Summary.Cell>
              <Table.Summary.Cell index={2} />
              <Table.Summary.Cell index={3} align="right">
                <strong>{formatAmount(totals.amount)}</strong>
              </Table.Summary.Cell>
              <Table.Summary.Cell index={4} colSpan={7} />
            </Table.Summary.Row>
          </Table.Summary>
        )}
      />

      <Space style={{ marginTop: 16 }}>
        <Button type="primary" loading={saving} onClick={() => void onSave()}>
          保存开票单
        </Button>
      </Space>
    </div>
  )
}
