import { Alert, Button, Space, Spin, message } from 'antd'
import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import OutboundSlip from '@/components/print/OutboundSlip'
import { defaultPrintProfile } from '@/config/printProfile'
import { getSalesOrder } from '@/services/salesOrderApi'
import type { SalesOrder } from '@/types/salesOrder'

export default function OrderPrintPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [order, setOrder] = useState<SalesOrder | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!id) {
      navigate('/orders')
      return
    }
    let cancelled = false
    ;(async () => {
      try {
        const data = await getSalesOrder(id)
        if (!cancelled) setOrder(data)
      } catch (e) {
        message.error(e instanceof Error ? e.message : '加载失败')
        navigate('/orders')
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [id, navigate])

  if (loading || !order) {
    return <Spin />
  }

  return (
    <div>
      <div className="no-print" style={{ marginBottom: 16 }}>
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <Space>
            <Button onClick={() => navigate('/orders')}>返回列表</Button>
            <Button type="primary" onClick={() => window.print()}>
              打印
            </Button>
            <Button onClick={() => navigate(`/orders/${order.id}`)}>编辑</Button>
          </Space>
          <Alert
            type="info"
            showIcon
            message="打印提示"
            description="请关闭浏览器页眉页脚，边距选「默认」或「最小」。若系统按 A4 出纸，票据画布靠上、周围留白可接受。"
          />
        </Space>
      </div>
      <OutboundSlip order={order} profile={defaultPrintProfile} />
    </div>
  )
}
