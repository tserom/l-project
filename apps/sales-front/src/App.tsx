import { Layout, Typography } from 'antd'
import { Outlet } from 'react-router-dom'

const { Header, Content } = Layout

export default function App() {
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header className="no-print" style={{ display: 'flex', alignItems: 'center' }}>
        <Typography.Title level={4} style={{ color: '#fff', margin: 0 }}>
          销售单
        </Typography.Title>
      </Header>
      <Content style={{ padding: 24 }}>
        <Outlet />
      </Content>
    </Layout>
  )
}
