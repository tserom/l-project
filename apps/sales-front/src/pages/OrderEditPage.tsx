import { Button, Form, Space, Spin, message } from 'antd'
import dayjs from 'dayjs'
import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import OrderForm, { toFormValues, type OrderFormValues } from '@/components/OrderForm'
import {
  createSalesOrder,
  getSalesOrder,
  updateSalesOrder,
} from '@/services/salesOrderApi'
import { generateOrderNo } from '@/utils/orderNo'

function emptyLine() {
  return {
    id: crypto.randomUUID(),
    materialName: '',
    spec: '',
    unit: 'kg',
    quantity: 0,
    unitPrice: 0,
    amount: 0,
    needInvoice: false,
  }
}

export default function OrderEditPage() {
  const { id } = useParams()
  const isNew = !id
  const navigate = useNavigate()
  const [form] = Form.useForm<OrderFormValues>()
  const [loading, setLoading] = useState(!isNew)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (isNew) {
      form.setFieldsValue({
        orderNo: generateOrderNo(),
        orderDate: dayjs(),
        customerName: '',
        warehouseName: '01金阳仓库',
        deliveryType: '提货',
        remark: '',
        lines: [emptyLine()],
      })
      return
    }

    let cancelled = false
    ;(async () => {
      setLoading(true)
      try {
        const order = await getSalesOrder(id)
        if (cancelled) return
        form.setFieldsValue(
          toFormValues({
            orderNo: order.orderNo,
            orderDate: order.orderDate,
            customerName: order.customerName,
            warehouseName: order.warehouseName,
            deliveryType: order.deliveryType,
            remark: order.remark,
            lines: order.lines,
          }),
        )
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
  }, [form, id, isNew, navigate])

  const onSave = async () => {
    try {
      const values = await form.validateFields()
      if (!values.lines?.length) {
        message.error('至少需要一行明细')
        return
      }
      setSaving(true)
      const input = {
        orderNo: values.orderNo,
        orderDate: values.orderDate.format('YYYY-MM-DD'),
        customerName: values.customerName,
        warehouseName: values.warehouseName,
        deliveryType: values.deliveryType,
        remark: values.remark,
        lines: values.lines,
      }
      if (isNew) {
        await createSalesOrder(input)
      } else {
        await updateSalesOrder(id, input)
      }
      message.success('已保存')
      navigate('/orders')
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(e instanceof Error ? e.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return <Spin />
  }

  return (
    <div>
      <h2 style={{ marginTop: 0 }}>{isNew ? '新建销售单' : '编辑销售单'}</h2>
      <OrderForm form={form} />
      <Space>
        <Button type="primary" loading={saving} onClick={() => void onSave()}>
          保存
        </Button>
        <Button onClick={() => navigate('/orders')}>返回</Button>
      </Space>
    </div>
  )
}
