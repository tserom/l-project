import { Col, DatePicker, Form, Input, Row } from 'antd'
import dayjs, { type Dayjs } from 'dayjs'
import OrderLinesTable from './OrderLinesTable'
import type { SalesOrderLine } from '@/types/salesOrder'

export type OrderFormValues = {
  orderNo: string
  orderDate: Dayjs
  customerName: string
  warehouseName: string
  deliveryType: string
  remark?: string
  lines: SalesOrderLine[]
}

type Props = {
  form: ReturnType<typeof Form.useForm<OrderFormValues>>[0]
}

export default function OrderForm({ form }: Props) {
  const lines = Form.useWatch('lines', form) ?? []

  return (
    <Form form={form} layout="vertical">
      <Row gutter={16}>
        <Col xs={24} md={8}>
          <Form.Item
            name="orderDate"
            label="日期"
            rules={[{ required: true, message: '请选择日期' }]}
          >
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
        </Col>
        <Col xs={24} md={8}>
          <Form.Item
            name="orderNo"
            label="单据号"
            rules={[{ required: true, message: '请输入单据号' }]}
          >
            <Input />
          </Form.Item>
        </Col>
        <Col xs={24} md={8}>
          <Form.Item
            name="customerName"
            label="客户"
            rules={[{ required: true, message: '请输入客户' }]}
          >
            <Input />
          </Form.Item>
        </Col>
        <Col xs={24} md={8}>
          <Form.Item name="warehouseName" label="仓库" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
        </Col>
        <Col xs={24} md={8}>
          <Form.Item name="deliveryType" label="出库类型" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
        </Col>
        <Col xs={24} md={8}>
          <Form.Item name="remark" label="备注">
            <Input />
          </Form.Item>
        </Col>
      </Row>
      <Form.Item name="lines" hidden>
        <Input />
      </Form.Item>
      <div style={{ marginBottom: 24 }}>
        <div style={{ marginBottom: 8 }}>明细</div>
        <OrderLinesTable
          value={lines}
          onChange={(next) => form.setFieldValue('lines', next)}
        />
      </div>
    </Form>
  )
}

export function toFormValues(input: {
  orderNo: string
  orderDate: string
  customerName: string
  warehouseName: string
  deliveryType: string
  remark?: string
  lines: SalesOrderLine[]
}): OrderFormValues {
  return {
    ...input,
    orderDate: dayjs(input.orderDate),
  }
}
