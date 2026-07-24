import { Button, Empty, Modal, Space, Table, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { seedSalesOrders } from '@/config/seed'
import {
  ensureSeedData,
  listSalesOrders,
  removeSalesOrder,
} from '@/services/salesOrderApi'
import type { SalesOrder } from '@/types/salesOrder'
import { formatOrderDate } from '@/utils/dateFormat'
import { formatAmount, formatQuantity } from '@/utils/money'

export default function OrderListPage() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)
  const [orders, setOrders] = useState<SalesOrder[]>([])

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
      <Table
        rowKey="id"
        loading={loading}
        columns={columns}
        dataSource={orders}
        locale={{
          emptyText: (
            <Empty description="暂无销售单">
              <Button type="primary" onClick={() => navigate('/orders/new')}>
                新建第一张销售单
              </Button>
            </Empty>
          ),
        }}
        pagination={false}
      />
    </div>
  )
}
