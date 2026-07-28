import { Layout, Menu, Typography } from 'antd'
import { Link, Outlet, useLocation } from 'react-router-dom'

const { Header, Content } = Layout

export default function App() {
  const location = useLocation()
  const selected = location.pathname.startsWith('/invoices')
    ? 'invoices'
    : 'orders'

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header
        className="no-print"
        style={{ display: 'flex', alignItems: 'center', gap: 24 }}
      >
        <Typography.Title level={4} style={{ color: '#fff', margin: 0 }}>
          销售开票
        </Typography.Title>
        <Menu
          theme="dark"
          mode="horizontal"
          selectedKeys={[selected]}
          style={{ flex: 1, minWidth: 0 }}
          items={[
            { key: 'orders', label: <Link to="/orders">销售单</Link> },
            { key: 'invoices', label: <Link to="/invoices">开票单</Link> },
          ]}
        />
      </Header>
      <Content style={{ padding: 24 }}>
        <Outlet />
      </Content>
    </Layout>
  )
}
