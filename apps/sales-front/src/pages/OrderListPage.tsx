import {
  Button,
  Card,
  Col,
  DatePicker,
  Empty,
  Form,
  Input,
  Modal,
  Row,
  Select,
  Space,
  Table,
  message,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import type { Dayjs } from 'dayjs'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { seedSalesOrders } from '@/config/seed'
import {
  ensureSeedData,
  listSalesOrders,
  removeSalesOrder,
} from '@/services/salesOrderApi'
import type { SalesOrder } from '@/types/salesOrder'
import { formatOrderDate } from '@/utils/dateFormat'
import {
  filterSalesOrders,
  parseOrderNos,
} from '@/utils/filterSalesOrders'
import { formatAmount, formatQuantity } from '@/utils/money'

type QueryForm = {
  orderNos?: string[]
  dateRange?: [Dayjs, Dayjs] | null
  customerName?: string
}

export default function OrderListPage() {
  const navigate = useNavigate()
  const [form] = Form.useForm<QueryForm>()
  const [loading, setLoading] = useState(true)
  const [orders, setOrders] = useState<SalesOrder[]>([])
  const [applied, setApplied] = useState<QueryForm>({})

  const reload = useCallback(async () => {
    setLoading(true)
    try {
      await ensureSeedData(seedSalesOrders)
      setOrders(await listSalesOrders())
    } catch (e) {
      message.error(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void reload()
  }, [reload])

  const filtered = useMemo(() => {
    const range = applied.dateRange
    return filterSalesOrders(orders, {
      orderNos: parseOrderNos(applied.orderNos),
      dateFrom: range?.[0]?.format('YYYY-MM-DD'),
      dateTo: range?.[1]?.format('YYYY-MM-DD'),
      customerName: applied.customerName,
    })
  }, [orders, applied])

  const onSearch = () => {
    setApplied(form.getFieldsValue())
  }

  const onReset = () => {
    form.resetFields()
    setApplied({})
  }

  const columns: ColumnsType<SalesOrder> = [
    { title: '单据号', dataIndex: 'orderNo', key: 'orderNo' },
    {
      title: '日期',
      dataIndex: 'orderDate',
      key: 'orderDate',
      render: (v: string) => formatOrderDate(v),
    },
    { title: '客户', dataIndex: 'customerName', key: 'customerName' },
    { title: '仓库', dataIndex: 'warehouseName', key: 'warehouseName' },
    { title: '出库类型', dataIndex: 'deliveryType', key: 'deliveryType' },
    {
      title: '数量合计',
      dataIndex: 'totalQuantity',
      key: 'totalQuantity',
      align: 'right',
      render: (v: number) => formatQuantity(v),
    },
    {
      title: '金额合计',
      dataIndex: 'totalAmount',
      key: 'totalAmount',
      align: 'right',
      render: (v: number) => formatAmount(v),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Link to={`/orders/${record.id}`}>编辑</Link>
          <Link to={`/orders/${record.id}/print`}>预览打印</Link>
          <Button
            type="link"
            danger
            onClick={() => {
              Modal.confirm({
                title: '确认删除该销售单？',
                onOk: async () => {
                  await removeSalesOrder(record.id)
                  message.success('已删除')
                  await reload()
                },
              })
            }}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => navigate('/orders/new')}>
          新建销售单
        </Button>
      </Space>

      <Card size="small" style={{ marginBottom: 16 }}>
        <Form form={form} layout="vertical" onFinish={onSearch}>
          <Row gutter={16}>
            <Col xs={24} md={8}>
              <Form.Item
                name="orderNos"
                label="单据号"
                extra="可输入多个，回车或逗号分隔"
              >
                <Select
                  mode="tags"
                  tokenSeparators={[',', '，', ' ']}
                  placeholder="单据号，支持多值"
                  allowClear
                />
              </Form.Item>
            </Col>
            <Col xs={24} md={8}>
              <Form.Item name="dateRange" label="日期">
                <DatePicker.RangePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col xs={24} md={8}>
              <Form.Item name="customerName" label="客户">
                <Input placeholder="客户名称，模糊匹配" allowClear />
              </Form.Item>
            </Col>
          </Row>
          <Space>
            <Button type="primary" htmlType="submit">
              查询
            </Button>
            <Button onClick={onReset}>重置</Button>
          </Space>
        </Form>
      </Card>

      <Table
        rowKey="id"
        loading={loading}
        columns={columns}
        dataSource={filtered}
        locale={{
          emptyText: (
            <Empty
              description={orders.length === 0 ? '暂无销售单' : '无符合条件的销售单'}
            >
              {orders.length === 0 ? (
                <Button type="primary" onClick={() => navigate('/orders/new')}>
                  新建第一张销售单
                </Button>
              ) : null}
            </Empty>
          ),
        }}
        pagination={false}
      />
    </div>
  )
}
